package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
)

type MoeCustomClaims struct {
	jwt.StandardClaims
}

type User struct {
	Username string
	Password string
}

type DiscordCode struct {
	Code  string
	State string
}

func PasswordLoginHandler(w http.ResponseWriter, r *http.Request) {
	signKey := []byte("test") //replace with file read
	user := User{}
	var response []byte
	json.NewDecoder(r.Body).Decode(&user)
	fmt.Println(user.Username + " " + user.Password)
	if user.Username != "test" || user.Password != "test" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response, _ = json.Marshal(struct{ Err string }{Err: "Invalid Username or Password"})
		w.Write(response)
		return
	}
	claims := MoeCustomClaims{
		jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "MoeBot API",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	codeGrant := DiscordCode{}
	var response []byte
	json.NewDecoder(r.Body).Decode(&codeGrant)
	client := &http.Client{}
	resp, _ := client.PostForm("https://discordapp.com/api/oauth2/token", url.Values{
		"client_id":     {"432291649754759169"},
		"client_secret": {"c6BViHBOTqmUm4tPOYPB9Ku0iUX8AjI-"},
		"grant_type":    {"authorization_code"},
		"code":          {codeGrant.Code},
		"redirect_uri":  {"http://localhost:8080/login/discord"},
	})
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
	}
	claims := MoeCustomClaims{
		jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "MoeBot API",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
