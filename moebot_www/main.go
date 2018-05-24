package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/camd67/moebot/moebot_www/auth"
	"github.com/camd67/moebot/moebot_www/db"
	"github.com/camd67/moebot/moebot_www/moebotApi"

	"github.com/gorilla/mux"
)

type config struct {
	Base64SecretKey string
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	OAuthLoginURI   string
	Address         string
	DbRootPwd       string
	DbUsername      string
	DbPassword      string
	DbHost          string
}

func main() {
	os.Setenv("MOEBOT_WWW_CONFIG_PATH", "config.json")
	configPath := os.Getenv("MOEBOT_WWW_CONFIG_PATH")
	config, err := readConfig(configPath)
	if err != nil {
		log.Println("Cannot read config file!")
		return
	}
	key, err := base64.StdEncoding.DecodeString(config.Base64SecretKey)
	if err != nil {
		log.Println("Invalid token generation key")
		return
	}
	db := db.NewDatabase(config.DbHost, config.DbRootPwd, config.DbUsername, config.DbPassword)
	db.Initialize()
	amw := auth.NewAuthMiddleware(key)
	authManager := auth.NewAuthManager(amw, db, config.ClientID, config.ClientSecret, config.RedirectURI, config.OAuthLoginURI)
	api := &moebotApi.MoebotApi{amw}

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/user", api.UserInfo)
	apiRouter.HandleFunc("/heartbeat", api.Heartbeat)
	apiRouter.Use(amw.Middleware)

	router.HandleFunc("/auth/password", authManager.PasswordLoginHandler)
	router.HandleFunc("/auth/register", authManager.PasswordRegister)
	router.HandleFunc("/auth/discord", authManager.DiscordBeginOAuth)
	router.HandleFunc("/auth/discordOAuth", authManager.DiscordCompleteOAuth)
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./dist/")))
	router.PathPrefix("/").HandlerFunc(vueHandler)

	log.Println("Starting MoeBot API on address " + config.Address + "...")

	http.ListenAndServe(config.Address, router)
}

func readConfig(path string) (*config, error) {
	result := &config{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func vueHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./dist/index.html")
}
