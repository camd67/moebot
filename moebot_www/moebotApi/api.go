package moebotApi

import (
	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"
	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_www/db"
)

type MoebotBindableApi interface {
	Binds() []bind
}

type bind struct {
	PermissionLevel botDb.Permission
	Handler         func(http.ResponseWriter, *http.Request)
	Path            string
	Methods         []string
}

type moebotDatabaseEndpoint struct {
	moeWebDb *db.MoeWebDb
}

type moebotDiscordEndpoint struct {
	session *discordgo.Session
}

func errorResponse(errMessage string) []byte {
	resp, _ := json.Marshal(&struct {
		Error string `json:"error"`
	}{Error: errMessage})
	return resp
}
