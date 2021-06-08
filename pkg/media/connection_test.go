package media_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/media"
	"gocv.io/x/gocv"
)

type testMockVideoCapture struct {
}

// SetP doesn't do anything it exists to satisfy VideoCapturable interface
func (tmvc *testMockVideoCapture) SetP(_ *gocv.VideoCapture) {}

// IsOpened always returns true.
func (tmvc *testMockVideoCapture) IsOpened() bool {
	return true
}

func (tmvc *testMockVideoCapture) Read(m *gocv.Mat) bool {
	return false
}

func (tmvc *testMockVideoCapture) Close() error {
	return nil
}

var _ = Describe("Connection", func() {
	Context("NewConnection", func() {
		It("Should return a new connection instance", func() {
			conn := media.NewConnection(
				"TestConnection",
				"/testroot/clips",
				30,
				2,
				schedule.Schedule{},
				config.ReolinkAdvanced{Enabled: false},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			Expect(conn).ToNot(BeNil())
			Expect(conn.UUID()).ToNot(BeEmpty())
			Expect(conn.Title()).To(Equal("TestConnection"))
		})
	})
})
