package moebotApi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_www/auth"
	"github.com/camd67/moebot/moebot_www/db"
)

type MoebotHeartbeatApi struct {
	amw *auth.AuthenticationMiddleware
	moebotDatabaseEndpoint
}

func NewHeartbeatApi(amw *auth.AuthenticationMiddleware, mdb *db.MoeWebDb) *MoebotHeartbeatApi {
	return &MoebotHeartbeatApi{amw, moebotDatabaseEndpoint{mdb}}
}

func (a *MoebotHeartbeatApi) Binds() []bind {
	return []bind{
		bind{Path: "/heartbeat", PermissionLevel: botDb.PermAll, Methods: []string{"GET"}, Handler: a.Heartbeat},
	}
}

func (a *MoebotHeartbeatApi) Heartbeat(w http.ResponseWriter, r *http.Request) {
	token, err := a.amw.GetToken(r.Header.Get("Authorization"))
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if time.Unix(token.Claims.(*auth.MoeCustomClaims).ExpiresAt, 0).Sub(time.Now()).Hours() < 24*7 {
		authType, _ := strconv.ParseInt(r.Header.Get("X-AuthType"), 10, 32)
		ss, err := a.amw.CreateSignedToken(r.Header.Get("X-UserID"), auth.AuthType(authType))
		response, err := json.Marshal(struct{ Jwt string }{Jwt: ss})
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}
