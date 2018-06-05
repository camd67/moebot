package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"

	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_www/auth"
	"github.com/camd67/moebot/moebot_www/db"
	"github.com/camd67/moebot/moebot_www/moebotApi"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type config struct {
	Base64SecretKey string
	ClientID        string
	ClientSecret    string
	MoeBotToken     string
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
	moeWebDb := db.NewDatabase(config.DbHost, config.DbRootPwd, config.DbUsername, config.DbPassword)
	moeWebDb.Initialize()

	botDb.SetupDatabase(config.DbHost, config.DbRootPwd, config.DbPassword)

	routeAuthMap := map[*mux.Route]botDb.Permission{}

	amw := auth.NewAuthMiddleware(key)
	authManager := auth.NewAuthManager(amw, moeWebDb, config.ClientID, config.ClientSecret, config.RedirectURI, config.OAuthLoginURI)
	session, err := discordgo.New("Bot " + config.MoeBotToken)
	if err != nil {
		log.Fatal("Error starting discord...", err)
	}

	api := &moebotApi.MoebotApi{Amw: amw, Session: session, MoeWebDb: moeWebDb}

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	routeAuthMap[apiRouter.HandleFunc("/user", api.UserInfo).Methods("GET")] = botDb.PermAll
	routeAuthMap[apiRouter.HandleFunc("/heartbeat", api.Heartbeat).Methods("GET")] = botDb.PermAll
	routeAuthMap[apiRouter.HandleFunc("/serverlist", api.ServerList).Methods("GET")] = botDb.PermAll
	routeAuthMap[apiRouter.HandleFunc("/{guildUID}/events", api.Heartbeat).Methods("GET")] = botDb.PermMod
	apiRouter.Use(amw.Middleware)

	router.HandleFunc("/auth/password", authManager.PasswordLoginHandler).Methods("POST")
	router.HandleFunc("/auth/register", authManager.PasswordRegister).Methods("POST")
	router.HandleFunc("/auth/discord", authManager.DiscordBeginOAuth).Methods("GET")
	router.HandleFunc("/auth/discordOAuth", authManager.DiscordCompleteOAuth).Methods("POST")
	router.PathPrefix("/static").Handler(handlers.CompressHandler(cacheControlWrapper(http.FileServer(http.Dir("./dist/")))))
	router.PathPrefix("/").HandlerFunc(vueHandler)

	log.Println("Starting MoeBot API on address " + config.Address + "...")

	log.Fatal(http.ListenAndServe(config.Address, router))
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

func cacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000")
		h.ServeHTTP(w, r)
	})
}
