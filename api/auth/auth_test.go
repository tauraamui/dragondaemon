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

func overloadTimeNow() func() {
	return auth.OverloadTimeNow(
		func() time.Time {
			return time.Date(2001, 1, 1, 12, 0, 0, 0, time.UTC)
		},
	)
}
