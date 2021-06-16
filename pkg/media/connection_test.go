package media_test

import (
	"context"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/media"
	"gocv.io/x/gocv"
)

type testMockVideoCapture struct {
	isOpenedFunc func() bool
	readFunc     func(m *gocv.Mat) bool
	closeFunc    func() error
}

// SetP doesn't do anything it exists to satisfy VideoCapturable interface
func (tmvc *testMockVideoCapture) SetP(_ *gocv.VideoCapture) {}

// IsOpened always returns true.
func (tmvc *testMockVideoCapture) IsOpened() bool {
	return tmvc.isOpenedFunc()
}

func (tmvc *testMockVideoCapture) Read(m *gocv.Mat) bool {
	return tmvc.readFunc(m)
}

func (tmvc *testMockVideoCapture) Close() error {
	return tmvc.closeFunc()
}

var _ = Describe("Connection", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	Context("NewConnection", func() {
		var resetFs func() = nil
		var mockFs afero.Fs = nil

		BeforeEach(func() {
			logging.CurrentLoggingLevel = logging.SilentLevel
			mockFs = afero.NewMemMapFs()
			resetFs = media.OverloadFS(mockFs)
		})

		AfterEach(func() {
			logging.CurrentLoggingLevel = existingLoggingLevel
			resetFs()
			mockFs = nil
		})

		It("Should return a new connection instance", func() {
			mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)
			conn := media.NewConnection(
				"TestConnection",
				media.ConnectonSettings{
					PersistLocation: "/testroot/clips",
					FPS:             30,
					SecondsPerClip:  2,
					Schedule:        schedule.Schedule{},
					Reolink:         config.ReolinkAdvanced{Enabled: false},
				},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			Expect(conn).ToNot(BeNil())
		})

		Context("Connection instance", func() {
			var conn *media.Connection
			var videoCapture *testMockVideoCapture

			BeforeEach(func() {
				mockFs.MkdirAll("/testroot/clips/TestConnectionInstance", os.ModeDir|os.ModePerm)
				conn = media.NewConnection(
					"TestConnectionInstance",
					media.ConnectonSettings{
						PersistLocation: "/testroot/clips",
						FPS:             15,
						SecondsPerClip:  3,
						Schedule:        schedule.Schedule{},
						Reolink:         config.ReolinkAdvanced{Enabled: false},
					},
					videoCapture,
					"test-connection-instance-addr",
				)
			})

			AfterEach(func() {
				videoCapture = &testMockVideoCapture{}
			})

			It("Should have populated UUID", func() {
				Expect(conn.UUID()).ToNot(BeEmpty())
			})

			It("Should have populated title", func() {
				Expect(conn.Title()).To(Equal("TestConnectionInstance"))
			})

			It("Should return total size on disk as EOF with empty size and unit values", func() {
				size, unit, err := conn.SizeOnDisk()
				Expect(int(size)).To(Equal(0))
				Expect(unit).To(BeEmpty())
				Expect(err).To(MatchError(io.EOF))
			})

			Context("Connection streaming frames to channel", func() {
				It("Should read video frames from given connection into buffer channel", func() {
					var matSumVal1 float64
					videoCapture.isOpenedFunc = func() bool { return true }
					videoCapture.readFunc = func(m *gocv.Mat) bool {
						mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
						defer mat.Close()
						mat.AddFloat(3.15)
						matSumVal1 = mat.Sum().Val1
						mat.CopyTo(m)
						return true
					}

					ctx, cancelStreaming := context.WithCancel(context.Background())
					stopping := conn.Stream(ctx)

					runningBufferLength := 0
					var readMat gocv.Mat
					defer readMat.Close()

					go func() {
						time.Sleep(500 * time.Millisecond)
						runningBufferLength = len(conn.Buffer())
						readMat = <-conn.Buffer()
						cancelStreaming()
					}()

					Eventually(stopping).Should(BeClosed())
					Expect(runningBufferLength).To(Equal(6))
					Expect(readMat.Sum().Val1).To(Equal(matSumVal1))
					Expect(conn.Buffer()).To(HaveLen(0))
				})
			})
		})
	})
})
