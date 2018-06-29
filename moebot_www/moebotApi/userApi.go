package moebotApi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_www/auth"
	"github.com/camd67/moebot/moebot_www/db"
)

type MoebotUserApi struct {
	pmw *auth.PermissionsMiddleware
	moebotDatabaseEndpoint
	moebotDiscordEndpoint
}

func NewUserApi(pmw *auth.PermissionsMiddleware, mdb *db.MoeWebDb, session *discordgo.Session) *MoebotUserApi {
	return &MoebotUserApi{pmw, moebotDatabaseEndpoint{mdb}, moebotDiscordEndpoint{session}}
}

func (a *MoebotUserApi) Binds() []bind {
	return []bind{
		bind{Path: "/user", PermissionLevel: botDb.PermAll, Methods: []string{"GET"}, Handler: a.UserInfo},
		bind{Path: "/user/guilds", PermissionLevel: botDb.PermAll, Methods: []string{"GET"}, Handler: a.ServerList},
	}
}

func (a *MoebotUserApi) UserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-UserID")
	dbUser, err := a.moeWebDb.Users.SelectByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errorResponse("Cannot read user from Database"))
		return
	}
	var response []byte
	if dbUser.DiscordUID.Valid {
		discordUser, err := a.session.User(dbUser.DiscordUID.String)
		if err != nil {
			log.Println("Failed to load Discord user - ", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errorResponse("Cannot find matching Discord User"))
			return
		}
		w.WriteHeader(http.StatusOK)
		response, _ = json.Marshal(&struct {
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
		}{Username: discordUser.Username, Avatar: fmt.Sprintf("https://cdn.discordapp.com/avatars/%v/%v.gif", discordUser.ID, discordUser.Avatar)})
	} else {
		w.WriteHeader(http.StatusOK)
		response, _ = json.Marshal(&struct {
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
		}{Username: dbUser.Username, Avatar: "/static/defaultDiscordAvatar.png"})
	}
	w.Write(response)
}

func (a *MoebotUserApi) ServerList(w http.ResponseWriter, r *http.Request) {
	type guildData struct {
		ServerUID  string `json:"id"`
		Icon       string `json:"icon"`
		Name       string `json:"name"`
		UserRights int    `json:"userRights"`
	}
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-UserID")
	responseData := []*guildData{}
	var response []byte

	discordUser, err := a.pmw.GetDiscordUser(userID)
	if err != nil {
		log.Println("Failed to load Discord user", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write(errorResponse("Cannot find matching Discord User"))
		return
	}

	guilds, err := a.session.UserGuilds(100, "", "")
	if err != nil {
		log.Println("Failed to load Guilds - ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errorResponse("Cannot find user guilds"))
		return
	}

	for _, g := range guilds {
		if m, err := a.session.GuildMember(g.ID, discordUser.ID); err == nil {
			guild, err := a.session.Guild(g.ID)
			if err == nil {
				m.GuildID = guild.ID
				var icon string
				if g.Icon != "" {
					icon = fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", guild.ID, g.Icon)
				} else {
					icon = "/static/defaultDiscordAvatar.png"
				}
				responseData = append(responseData, &guildData{ServerUID: g.ID, Icon: icon, Name: g.Name, UserRights: int(auth.GetPermissionLevel(m, guild))})
			}
		}
	}
	response, _ = json.Marshal(responseData)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	return
}
