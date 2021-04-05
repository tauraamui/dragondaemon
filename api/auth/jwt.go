package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type customClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenToken(secret, username string) (string, error) {
	claims := customClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			Audience:  "dragondaemon",
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
