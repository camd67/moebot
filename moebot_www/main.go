package main

import (
	"net/http"

	"github.com/camd67/moebot/moebot_www/auth"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/auth/password", auth.PasswordLoginHandler)
	router.HandleFunc("/auth/discord", auth.DiscordBeginOAuth)
	router.HandleFunc("/auth/discordOAuth", auth.DiscordCompleteOAuth)
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./dist/")))
	router.PathPrefix("/").HandlerFunc(vueHandler)

	http.ListenAndServe(":8081", router)
}

func vueHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./dist/index.html")
}
