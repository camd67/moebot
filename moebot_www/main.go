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
	router := mux.NewRouter()
	moeWebDb := db.NewDatabase(config.DbHost, config.DbRootPwd, config.DbUsername, config.DbPassword)
	moeWebDb.Initialize()

	botDb.SetupDatabase(config.DbHost, config.DbRootPwd, config.DbPassword)

	session, err := discordgo.New("Bot " + config.MoeBotToken)
	if err != nil {
		log.Fatal("Error starting discord...", err)
	}
	amw := auth.NewAuthMiddleware(key)
	pmw := auth.NewPermissionsMiddleware(session, moeWebDb)

	configureAuthEndpoints(router, amw, moeWebDb, config)

	configureAPIEndpoints(router, amw, pmw, session, moeWebDb)

	router.PathPrefix("/static").Handler(handlers.CompressHandler(cacheControlWrapper(http.FileServer(http.Dir("./dist/")))))
	router.PathPrefix("/").HandlerFunc(vueHandler)

	log.Println("Starting MoeBot API on address " + config.Address + "...")

	log.Fatal(http.ListenAndServe(config.Address, router))
}

func configureAuthEndpoints(router *mux.Router, amw *auth.AuthenticationMiddleware, moeWebDb *db.MoeWebDb, config *config) {
	authManager := auth.NewAuthManager(amw, moeWebDb, config.ClientID, config.ClientSecret, config.RedirectURI, config.OAuthLoginURI)
	router.HandleFunc("/auth/password", authManager.PasswordLoginHandler).Methods("POST")
	router.HandleFunc("/auth/register", authManager.PasswordRegister).Methods("POST")
	router.HandleFunc("/auth/discord", authManager.DiscordBeginOAuth).Methods("GET")
	router.HandleFunc("/auth/discordOAuth", authManager.DiscordCompleteOAuth).Methods("POST")
}

func configureAPIEndpoints(router *mux.Router, amw *auth.AuthenticationMiddleware, pmw *auth.PermissionsMiddleware, session *discordgo.Session, moeWebDb *db.MoeWebDb) {
	apiList := createAPIEndpoints(amw, pmw, session, moeWebDb)

	apiRouter := router.PathPrefix("/api").Subrouter()
	for _, a := range apiList {
		for _, bind := range a.Binds() {
			pmw.AddRoute(apiRouter.HandleFunc(bind.Path, bind.Handler).Methods(bind.Methods...), bind.PermissionLevel)
		}
	}
	apiRouter.Use(amw.Middleware)
	apiRouter.Use(pmw.Middleware)
}

func createAPIEndpoints(amw *auth.AuthenticationMiddleware, pmw *auth.PermissionsMiddleware, session *discordgo.Session, moeWebDb *db.MoeWebDb) []moebotApi.MoebotBindableApi {
	return []moebotApi.MoebotBindableApi{
		moebotApi.NewUserApi(pmw, moeWebDb, session),
		moebotApi.NewHeartbeatApi(amw, moeWebDb),
		moebotApi.NewEventsApi(moeWebDb, session),
	}
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
