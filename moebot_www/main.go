package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/camd67/moebot/moebot_www/auth"
	"github.com/camd67/moebot/moebot_www/moebot_api"
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
)

type authenticationMiddleware struct {
	secretKey []byte
}

func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(strings.TrimPrefix(tokenString, "Bearer "), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return amw.secretKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	amw := authenticationMiddleware{}
	api.HandleFunc("/user", moebotApi.UserInfo)
	api.Use(amw.Middleware)
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
