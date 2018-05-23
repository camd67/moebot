package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	jwt "github.com/dgrijalva/jwt-go"
)

type MoeCustomClaims struct {
	jwt.StandardClaims
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Avatar   string `json:"avatar"`
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

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	signKey := []byte("test") //replace with file read
	user := User{}
	var response []byte
	json.NewDecoder(r.Body).Decode(&user)
	if user.Username != "test" || user.Password != "test" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Invalid Username or Password"})
		w.Write(response)
		return
	}
	token := createNewToken(user.Username)
	ss, _ := token.SignedString(signKey)
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

func DiscordBeginOAuth(w http.ResponseWriter, r *http.Request) {
	const redirectURL = "https://discordapp.com/api/oauth2/authorize?client_id=432291649754759169&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Flogin%2Fdiscord&response_type=code&scope=identify%20guilds"
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func DiscordCompleteOAuth(w http.ResponseWriter, r *http.Request) {
	signKey := []byte("test") //replace with file read
	codeGrant := discordOAuthCode{}
	var response []byte
	json.NewDecoder(r.Body).Decode(&codeGrant)
	discordToken := discordRequestToken(codeGrant.Code)
	session, _ := discordgo.New("Bearer " + discordToken.Access_token)
	defer session.Close()
	user, err := session.User("@me")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token := createNewToken(user.Username)
	ss, _ := token.SignedString(signKey)
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

func createNewToken(username string) *jwt.Token {
	currentTime := time.Now()
	claims := MoeCustomClaims{
		jwt.StandardClaims{
			IssuedAt:  currentTime.Unix(),
			ExpiresAt: currentTime.Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "MoeBot API",
			Subject:   username,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}

func discordRequestToken(codeGrant string) discordToken {
	token := discordToken{}
	client := &http.Client{}
	r, _ := client.PostForm("https://discordapp.com/api/oauth2/token", url.Values{
		"client_id":     {"432291649754759169"},
		"client_secret": {"c6BViHBOTqmUm4tPOYPB9Ku0iUX8AjI-"},
		"grant_type":    {"authorization_code"},
		"code":          {codeGrant},
		"redirect_uri":  {"http://localhost:8080/login/discord"},
	})
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		json.NewDecoder(r.Body).Decode(&token)
	}
	return token
}

func CheckToken(tokenString string) (*jwt.Token, error) {
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, fmt.Errorf("Unrecognized Authorization Header")
	}
	token, err := jwt.Parse(strings.TrimPrefix(tokenString, "Bearer "), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("test"), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
