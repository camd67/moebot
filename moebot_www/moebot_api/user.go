package moebotApi

import (
	"encoding/json"
	"net/http"

	"github.com/camd67/moebot/moebot_www/auth"
)

func UserInfo(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CheckToken(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response, _ := json.Marshal(auth.User{Username: "test", Avatar: "/static/baseline_person_black_18dp.png"})
	w.Write(response)
}
