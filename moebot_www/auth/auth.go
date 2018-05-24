package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_www/db"
	"github.com/lib/pq"
)

type AuthManager struct {
	amw           *AuthenticationMiddleware
	db            *db.MoeWebDb
	clientID      string
	clientSecret  string
	redirectURI   string
	oAuthLoginURI string
}

type discordToken struct {
	Access_token  string
	Token_type    string
	Expires_in    int64
	Refresh_token string
	Scope         string
}

type discordOAuthCode struct {
	Code  string
	State string
}

func NewAuthManager(amw *AuthenticationMiddleware, db *db.MoeWebDb, clientID string, clientSecret string, redirectURI string, oAuthLoginURI string) *AuthManager {
	return &AuthManager{amw: amw, db: db, clientID: clientID, clientSecret: clientSecret, redirectURI: redirectURI, oAuthLoginURI: oAuthLoginURI}
}

func (a *AuthManager) PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	request := &struct {
		Username string
		Password string
	}{}
	var response []byte
	json.NewDecoder(r.Body).Decode(request)
	user, err := a.db.Users.SelectByUsername(request.Username, request.Password)
	if user == nil || err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Invalid Username or Password"})
		w.Write(response)
		return
	}
	ss, err := a.amw.CreateSignedToken(user.ID, PasswordAuth)
	response, err = json.Marshal(struct{ Jwt string }{Jwt: ss})
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *AuthManager) DiscordBeginOAuth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.oAuthLoginURI, http.StatusSeeOther)
}

func (a *AuthManager) DiscordCompleteOAuth(w http.ResponseWriter, r *http.Request) {
	codeGrant := discordOAuthCode{}
	var response []byte
	json.NewDecoder(r.Body).Decode(&codeGrant)
	discordToken := a.discordRequestToken(codeGrant.Code)
	session, _ := discordgo.New("Bearer " + discordToken.Access_token)
	defer session.Close()
	discordUser, err := session.User("@me")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := a.db.Users.SelectByDiscordUID(discordUser.ID)
	if user == nil || err != nil {
		user = &db.User{
			Username:               discordUser.ID,
			DiscordUID:             sql.NullString{String: discordUser.ID, Valid: true},
			DiscordAuthToken:       sql.NullString{String: discordToken.Access_token, Valid: true},
			DiscordTokenExpiration: pq.NullTime{Time: time.Now().Add(time.Second * time.Duration(discordToken.Expires_in)), Valid: true},
		}
		user, err = a.db.Users.InsertUser(user, sql.NullString{}, sql.NullString{String: discordToken.Refresh_token, Valid: true})
	}
	ss, err := a.amw.CreateSignedToken(user.ID, DiscordAuth)
	response, err = json.Marshal(struct{ Jwt string }{Jwt: ss})
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *AuthManager) PasswordRegister(w http.ResponseWriter, r *http.Request) {
	request := &struct {
		Username string
		Password string
		Email    string
	}{}
	var response []byte
	json.NewDecoder(r.Body).Decode(request)
	if len(request.Username) < 4 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Username too short (minimum 4 characters)"})
		w.Write(response)
		return
	}
	if len(request.Password) < 7 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Password too short (minimum 7 characters)"})
		w.Write(response)
		return
	}
	if !regexp.MustCompile("^\\w+([.-]?\\w+)*@\\w+([.-]?\\w+)*(\\.\\w{2,3})+$").MatchString(request.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Invalid Email"})
		w.Write(response)
		return
	}
	user := &db.User{
		Username: request.Username,
		Email:    sql.NullString{String: request.Email, Valid: true},
	}
	user, err := a.db.Users.InsertUser(user, sql.NullString{String: request.Password, Valid: true}, sql.NullString{})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "User already exists"})
		w.Write(response)
		return
	}
	ss, err := a.amw.CreateSignedToken(user.ID, PasswordAuth)
	response, err = json.Marshal(struct{ Jwt string }{Jwt: ss})
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *AuthManager) discordRequestToken(codeGrant string) discordToken {
	token := discordToken{}
	client := &http.Client{}
	r, _ := client.PostForm("https://discordapp.com/api/oauth2/token", url.Values{
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {codeGrant},
		"redirect_uri":  {a.redirectURI},
	})
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		json.NewDecoder(r.Body).Decode(&token)
	}
	return token
}
