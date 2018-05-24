package moebotApi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/camd67/moebot/moebot_www/auth"
)

type MoebotApi struct {
	Amw *auth.AuthenticationMiddleware
}

func (a *MoebotApi) UserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response, _ := json.Marshal(&struct {
		Username string
		Avatar   string
	}{Username: r.Header.Get("X-UserID"), Avatar: "/static/baseline_person_black_18dp.png"})
	w.Write(response)
}

func (a *MoebotApi) Heartbeat(w http.ResponseWriter, r *http.Request) {
	token, err := a.Amw.GetToken(r.Header.Get("Authorization"))
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if time.Unix(token.Claims.(*auth.MoeCustomClaims).ExpiresAt, 0).Sub(time.Now()).Hours() < 24*7 {
		authType, _ := strconv.ParseInt(r.Header.Get("X-AuthType"), 10, 32)
		ss, err := a.Amw.CreateSignedToken(r.Header.Get("X-UserID"), auth.AuthType(authType))
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
