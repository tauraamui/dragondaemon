package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type customClaims struct {
	UserUUID string `json:"useruuid"`
	jwt.StandardClaims
}

var TimeNow = func() time.Time {
	return time.Now()
}

func GenToken(secret, username string) (string, error) {
	claims := customClaims{
		UserUUID: username,
		StandardClaims: jwt.StandardClaims{
			Audience:  "dragondaemon",
			ExpiresAt: TimeNow().UTC().Add(time.Minute * 15).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(secret, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return "", errors.New("unable to pass claims")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return "", errors.New("auth token has expired")
	}
	return claims.UserUUID, nil
}
