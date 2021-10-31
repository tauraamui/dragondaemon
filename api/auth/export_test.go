package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var CheckClaims = checkClaims
var CustomClaims = &customClaims{
	UserUUID:       "",
	StandardClaims: jwt.StandardClaims{},
}
var NotPointerForCustomClaims customClaims = customClaims{}

func OverloadTimeNow(o func() time.Time) func() {
	timeNowRef := timeNow
	timeNow = o
	return func() { timeNow = timeNowRef }
}
