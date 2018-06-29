package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_www/db"
	"github.com/gorilla/mux"
)

type PermissionsMiddleware struct {
	routePermissionsMap map[*mux.Route]botDb.Permission
	session             *discordgo.Session
	moeWebDb            *db.MoeWebDb
}

func NewPermissionsMiddleware(s *discordgo.Session, mdb *db.MoeWebDb) *PermissionsMiddleware {
	return &PermissionsMiddleware{map[*mux.Route]botDb.Permission{}, s, mdb}
}

func (mw *PermissionsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentRoute := mux.CurrentRoute(r)
		vars := mux.Vars(r)
		var p botDb.Permission
		if p, ok := mw.routePermissionsMap[currentRoute]; !ok || p == botDb.PermNone {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if guildUID, ok := vars["guildUID"]; ok {
			if check, err := mw.CheckGuildPermissions(r.Header.Get("X-UserID"), guildUID, p); !check || err != nil {
				log.Println(err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (mw *PermissionsMiddleware) AddRoute(route *mux.Route, perm botDb.Permission) {
	mw.routePermissionsMap[route] = perm
}

func (mw *PermissionsMiddleware) CheckGuildPermissions(userID string, guildUID string, permLevel botDb.Permission) (bool, error) {
	isMoeGuild, err := mw.CheckMoeGuild(guildUID)
	if !isMoeGuild || err != nil {
		if err == nil {
			err = fmt.Errorf("No such guild in moebot database")
		}
		return false, err
	}
	discordUser, err := mw.GetDiscordUser(userID)
	if err != nil {
		return false, err
	}
	guild, err := mw.session.Guild(guildUID)
	if err != nil {
		log.Println("Cannot retreive guild - ", err)
		return false, err
	}
	member, err := mw.session.GuildMember(guildUID, discordUser.ID)
	if err != nil {
		log.Println("Cannot retreive guild member - ", err)
		return false, err
	}
	return GetPermissionLevel(member, guild) >= permLevel, nil
}

func (mw *PermissionsMiddleware) CheckMoeGuild(guildUID string) (bool, error) {
	moeGuilds, err := mw.session.UserGuilds(100, "", "")
	if err != nil {
		log.Println("Cannot retreive moebot guilds - ", err)
		return false, err
	}
	for _, g := range moeGuilds {
		if g.ID == guildUID {
			return false, nil
		}
	}
	return false, nil
}

func (mw *PermissionsMiddleware) GetDiscordUser(userID string) (*discordgo.User, error) {
	dbUser, err := mw.moeWebDb.Users.SelectByID(userID)
	if err != nil {
		return nil, err
	}

	if !dbUser.DiscordUID.Valid {
		return nil, fmt.Errorf("No discord user informations available")
	}

	discordUser, err := mw.session.User(dbUser.DiscordUID.String)
	if err != nil {
		log.Println("Cannot retreive discord User - ", err)
		return nil, err
	}
	return discordUser, nil
}

func GetPermissionLevel(member *discordgo.Member, guild *discordgo.Guild) botDb.Permission {
	if member.GuildID != guild.ID {
		return botDb.PermNone
	}
	if guild.OwnerID == member.User.ID {
		return botDb.PermGuildOwner
	}

	perms := botDb.RoleQueryPermission(member.Roles)
	highestPerm := botDb.PermAll
	// Find the highest permission level this user has
	for _, userPerm := range perms {
		if userPerm > highestPerm {
			highestPerm = userPerm
		}
	}
	return highestPerm
}
