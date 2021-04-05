package auth

import "github.com/dgrijalva/jwt-go"

type customClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenToken(secret, username string) (string, error) {
	claims := customClaims{
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
