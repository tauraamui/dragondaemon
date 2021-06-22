package media_test

import (
	"context"
	"errors"
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
	if tmvc.readFunc != nil {
		return tmvc.isOpenedFunc()
	}
	panic(errors.New("call to missing test mock video capture isOpened function"))
}

func (tmvc *testMockVideoCapture) Read(m *gocv.Mat) bool {
	if tmvc.readFunc != nil {
		return tmvc.readFunc(m)
	}
	panic(errors.New("call to missing test mock video capture read function"))
}

func (tmvc *testMockVideoCapture) Close() error {
	if tmvc.closeFunc != nil {
		return tmvc.closeFunc()
	}
	panic(errors.New("call to missing test mock video capture close function"))
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
			var resetVidCapOverload func()
			var openVidCapCallback func()

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

				resetVidCapOverload = media.OverloadOpenVideoCapture(
					func(string, string, int, bool, string) (media.VideoCapturable, error) {
						if openVidCapCallback != nil {
							openVidCapCallback()
						}
						return videoCapture, nil
					},
				)
			})

			AfterEach(func() {
				resetVidCapOverload()
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
					// make sure the buffer is filled max to 6
					Expect(runningBufferLength).To(Equal(6))
					// make sure mat read matches mat written
					Expect(readMat.Sum().Val1).To(Equal(matSumVal1))
					// make sure buffer is flushed on stream shutdown
					Expect(conn.Buffer()).To(HaveLen(0))
				})

				It("Should attempt to reconnect until read eventually returns ok", func() {
					videoCapture.isOpenedFunc = func() bool { return true }

					openVideoCaptureCallCount := 0
					openVidCapCallback = func() { openVideoCaptureCallCount++ }

					closeCallCount := 0
					videoCapture.closeFunc = func() error {
						closeCallCount++
						return nil
					}

					readCallCount := 0
					videoCapture.readFunc = func(m *gocv.Mat) bool {
						readCallCount++
						return readCallCount > 10
					}

					ctx, cancelStreaming := context.WithCancel(context.Background())
					stopping := conn.Stream(ctx)

					go func() {
						defer cancelStreaming()
						for {
							if openVideoCaptureCallCount >= 10 && closeCallCount >= 10 {
								break
							}
						}
					}()

					Eventually(stopping).Should(BeClosed())
					Expect(openVideoCaptureCallCount).To(Equal(10))
					Expect(closeCallCount).To(Equal(10))
				})
			})
		})
	})
})
