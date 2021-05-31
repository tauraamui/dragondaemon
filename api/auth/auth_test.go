package auth_test

import (
	"encoding/json"
	"strings"

	"github.com/dgrijalva/jwt-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/api/auth"
)

type testCustomClaims struct {
	UserUUID string `json:"useruuid"`
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
		})
	})
})
