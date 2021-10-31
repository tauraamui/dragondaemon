package auth_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/api/auth"
)

const testingSecret = "testsecret"

type testCustomClaims struct {
	UserUUID  string `json:"useruuid"`
	Audience  string `json:"aud"`
	ExpiresAt int    `json:"exp"`
}

func TestGenTokenReturnsSignedJWTContainingPayload(t *testing.T) {
	reset := overloadTimeNow()
	defer reset()

	is := is.New(t)
	token, err := auth.GenToken(testingSecret, "testuser")
	is.NoErr(err)
	is.True(len(token) > 0)

	segs := strings.Split(token, ".")
	is.Equal(len(segs), 3)

	claims, err := jwt.DecodeSegment(segs[1])
	is.NoErr(err)
	tc := testCustomClaims{}
	is.NoErr(json.Unmarshal(claims, &tc))

	is.Equal(tc.UserUUID, "testuser")
	is.Equal(tc.Audience, "dragondaemon")
	is.True(tc.ExpiresAt != 0)
	println(tc.ExpiresAt)
	println(time.Unix(int64(tc.ExpiresAt), 0).String())
	is.Equal(time.Unix(int64(tc.ExpiresAt), 0).Minute(), 15)
}

func TestValidateTokenReturnCurrentUserUUIDFromPayload(t *testing.T) {
	is := is.New(t)

	token, err := auth.GenToken(testingSecret, "testuser")
	is.NoErr(err)
	is.True(len(token) > 0)

	uuid, err := auth.ValidateToken(testingSecret, token)
	is.NoErr(err)
	is.Equal(uuid, "testuser")
}

func TestHandleValidationErrorGracefullyAndReturnWrappedError(t *testing.T) {
	is := is.New(t)

	token, err := auth.GenToken(testingSecret, "testuser")
	is.NoErr(err)
	is.True(len(token) > 0)

	uuid, err := auth.ValidateToken("incorrect-secret", token)
	is.True(err != nil)
	is.Equal(err.Error(), "unable to validate token: signature is invalid")
	is.True(len(uuid) == 0)
}

func TestHandleUnableToParseClaimsErrorGracefullyAndReturnError(t *testing.T) {
	is := is.New(t)

	uuid, err := auth.CheckClaims(auth.NotPointerForCustomClaims)
	is.True(err != nil)
	is.Equal(err.Error(), "unable to parse claims")
	is.True(len(uuid) == 0)
}

func TestHandleTokenExpiredErrorGracefullyAndReturnError(t *testing.T) {
	is := is.New(t)

	auth.CustomClaims.StandardClaims = jwt.StandardClaims{
		ExpiresAt: 1,
	}

	uuid, err := auth.CheckClaims(auth.CustomClaims)
	is.True(err != nil)
	is.Equal(err.Error(), "auth token has expired")
	is.True(len(uuid) == 0)
}

func overloadTimeNow() func() {
	return auth.OverloadTimeNow(
		func() time.Time {
			return time.Date(2001, 1, 1, 12, 0, 0, 0, time.UTC)
		},
	)
}
