package auth

import "github.com/dgrijalva/jwt-go"

var CheckClaims = checkClaims
var CustomClaims = &customClaims{
	UserUUID:       "",
	StandardClaims: jwt.StandardClaims{},
}
var NotPointerForCustomClaims customClaims = customClaims{}
