package auth_test

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/api/auth"
)

type testCustomClaims struct {
	UserUUID  string `json:"useruuid"`
	Audience  string `json:"aud"`
	ExpiresAt int    `json:"exp"`
}

var _ = Describe("Auth", func() {

	const TESTING_SECRET = "testsecret"
	existingLoggingLevel := logging.CurrentLoggingLevel

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
	})

	Context("GenToken", func() {
		It("Should return a JWT which contains payload data with given username and is signed correctly", func() {
			auth.TimeNow = func() time.Time {
				return time.Date(2001, 1, 1, 12, 0, 0, 0, time.UTC)
			}

			defer func() { auth.TimeNow = time.Now }()

			token, err := auth.GenToken(TESTING_SECRET, "testuser")
			Expect(err).To(BeNil())
			Expect(token).ToNot(BeEmpty())

			tokenSegments := strings.Split(token, ".")
			Expect(tokenSegments).To(HaveLen(3))
			decodedClaims, err := jwt.DecodeSegment(tokenSegments[1])
			Expect(err).To(BeNil())

			testCustomClaims := testCustomClaims{}
			json.Unmarshal(decodedClaims, &testCustomClaims)

			Expect(testCustomClaims.UserUUID).To(Equal("testuser"))
			Expect(testCustomClaims.Audience).To(Equal("dragondaemon"))
			Expect(testCustomClaims.ExpiresAt).ToNot(BeZero())
			Expect(time.Unix(int64(testCustomClaims.ExpiresAt), 0).Minute()).To(Equal(15))
		})

		It("Should validate token and return correct user UUID from payload segment", func() {
			token, err := auth.GenToken(TESTING_SECRET, "testuser")
			Expect(err).To(BeNil())
			Expect(token).ToNot(BeEmpty())

			userUUID, err := auth.ValidateToken(TESTING_SECRET, token)
			Expect(err).To(BeNil())
			Expect(userUUID).To(Equal("testuser"))
		})
	})
})
