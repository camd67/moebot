package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type AuthenticationMiddleware struct {
	secretKey []byte
}

type MoeCustomClaims struct {
	AuthenticationType AuthType `json:authenticationType`
	jwt.StandardClaims
}

type AuthType int64

const (
	PasswordAuth AuthType = iota
	DiscordAuth  AuthType = iota
)

func NewAuthMiddleware(key []byte) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{key}
}

func (amw *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := amw.GetToken(r.Header.Get("Authorization"))
		if err != nil || !token.Valid {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-UserID", token.Claims.(*MoeCustomClaims).Subject)
		r.Header.Set("X-AuthType", fmt.Sprintf("%v", token.Claims.(*MoeCustomClaims).AuthenticationType))
		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) GetToken(authHeader string) (*jwt.Token, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("No token in header")
	}
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(authHeader, "Bearer "), &MoeCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return amw.secretKey, nil
	})
	return token, err
}

func (amw *AuthenticationMiddleware) CreateSignedToken(username string, authType AuthType) (string, error) {
	currentTime := time.Now()
	claims := MoeCustomClaims{
		authType,
		jwt.StandardClaims{
			IssuedAt:  currentTime.Unix(),
			ExpiresAt: currentTime.Add(time.Hour * 24 * 30).Unix(),
			Issuer:    "MoeBot API",
			Subject:   username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(amw.secretKey)
}
