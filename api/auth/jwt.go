package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tauraamui/xerror"
)

type customClaims struct {
	UserUUID string `json:"useruuid"`
	jwt.StandardClaims
}

var timeNow = func() time.Time {
	return time.Now()
}

func GenToken(secret, username string) (string, error) {
	expiresAt := timeNow().UTC().Add(time.Minute * 15).Unix()
	claims := customClaims{
		UserUUID: username,
		StandardClaims: jwt.StandardClaims{
			Audience:  "dragondaemon",
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(secret, tokenString string) (string, error) {
	jwt.TimeFunc = timeNow
	token, err := jwt.ParseWithClaims(
		tokenString,
		&customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return "", xerror.Errorf("unable to validate token: %w", err)
	}

	return checkClaims(token.Claims)
}

func checkClaims(claims jwt.Claims) (string, error) {
	cc, ok := claims.(*customClaims)
	if !ok {
		return "", xerror.New("unable to parse claims")
	}

	if cc.ExpiresAt < timeNow().UTC().Unix() {
		return "", xerror.New("auth token has expired")
	}

	return cc.UserUUID, nil
}
