package media_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

const KB int64 = 1024

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
				&testMockVideoCapture{
					closeFunc: func() error { return nil },
				},
				"fake-stream-addr",
			)

			Expect(conn).ToNot(BeNil())
			Expect(conn.Close()).To(BeNil())
		})

		It("Should return connection instance with missing reolink connection", func() {
			mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)
			conn := media.NewConnection(
				"TestConnection",
				media.ConnectonSettings{
					PersistLocation: "/testroot/clips",
					FPS:             30,
					SecondsPerClip:  2,
					Schedule:        schedule.Schedule{},
					Reolink:         config.ReolinkAdvanced{Enabled: true},
				},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			Expect(conn.ReolinkControl()).To(BeNil())
		})

		Context("Using a connection instance", func() {
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

			AfterSuite(func() {
				Expect(conn.Close()).To(BeNil())
				conn = nil
			})

			It("Should have populated UUID", func() {
				Expect(conn.UUID()).ToNot(BeEmpty())
			})

			It("Should have populated title", func() {
				Expect(conn.Title()).To(Equal("TestConnectionInstance"))
			})

			Context("Calling unitizeSize directly", func() {
				It("Should return different units and sizes given byte counts", func() {
					size, unit := media.UnitizeSize(KB)
					Expect(size).To(BeNumerically("==", 1))
					Expect(unit).To(Equal("KB"))

					size, unit = media.UnitizeSize(KB * KB)
					Expect(size).To(BeNumerically("==", 1))
					Expect(unit).To(Equal("MB"))

					size, unit = media.UnitizeSize(KB * KB * KB)
					Expect(size).To(BeNumerically("==", 1))
					Expect(unit).To(Equal("GB"))
				})
			})

			Context("Connect checking total file size in persist dir", func() {
				It("Should return total size on disk as EOF with empty size and unit values", func() {
					size, err := conn.SizeOnDisk()
					Expect(size).To(Equal("0KB"))
					Expect(err).To(MatchError(io.EOF))
				})

				It("Should return total size on disk which matches real total size", func() {
					clipsDirPath := "/testroot/clips/TestConnectionInstance"

					By("Creating file on disk of size 9KB")
					binFile, err := mockFs.Create(filepath.Join(clipsDirPath, "mock.bin"))
					Expect(err).To(BeNil())
					defer binFile.Close()
					err = binFile.Truncate(KB * 9)
					Expect(err).To(BeNil())

					By("Querying disk size with just created file on disk")
					size, err := conn.SizeOnDisk()
					Expect(size).To(Equal("9KB"))
					Expect(err).To(BeNil())
				})

				It("Should return total size on disk from checking disk and then reading from cache", func() {
					clipsDirPath := "/testroot/clips/TestConnectionInstance"
					mockFs.MkdirAll(clipsDirPath, os.ModeDir|os.ModePerm)

					By("Creating file on disk of size 9KB")
					binFile, err := mockFs.Create(filepath.Join(clipsDirPath, "mock.bin"))
					Expect(err).To(BeNil())
					defer binFile.Close()
					err = binFile.Truncate(KB * 9)
					Expect(err).To(BeNil())

					By("Querying disk size with just created file on disk")
					size, err := conn.SizeOnDisk()
					Expect(size).To(Equal("9KB"))
					Expect(err).To(BeNil())

					By("Querying disk size after deleting all files from disk")
					mockFs.Remove(binFile.Name())

					size, err = conn.SizeOnDisk()
					Expect(size).To(Equal("9KB"))
					Expect(err).To(BeNil())
				})

				It("Should return total size on disk including sub dirs within persist dir", func() {
					clipsRootDirPath := "/testroot/clips/TestConnectionInstance"

					clipsSubDirPath1 := "/testroot/clips/TestConnectionInstance/subdir1"
					mockFs.MkdirAll(clipsSubDirPath1, os.ModeDir|os.ModePerm)

					clipsSubDirPath2 := "/testroot/clips/TestConnectionInstance/subdir2"
					mockFs.MkdirAll(clipsSubDirPath2, os.ModeDir|os.ModePerm)

					By("Creating file on disk within root dir of size 6KB")
					rootBinFile, err := mockFs.Create(filepath.Join(clipsRootDirPath, "mock.bin"))
					Expect(err).To(BeNil())
					defer rootBinFile.Close()
					err = rootBinFile.Truncate(KB * 6)
					Expect(err).To(BeNil())

					By("Creating file on disk within sub dir 1 of size 6KB")
					subBinFile1, err := mockFs.Create(filepath.Join(clipsSubDirPath1, "mock.bin"))
					Expect(err).To(BeNil())
					defer subBinFile1.Close()
					err = subBinFile1.Truncate(KB * 6)
					Expect(err).To(BeNil())

					By("Creating file on disk within sub dir 2 of size 6KB")
					subBinFile2, err := mockFs.Create(filepath.Join(clipsSubDirPath2, "mock.bin"))
					Expect(err).To(BeNil())
					defer subBinFile2.Close()
					err = subBinFile2.Truncate(KB * 6)
					Expect(err).To(BeNil())

					By("Querying disk size with all just created files on disk")
					size, err := conn.SizeOnDisk()
					Expect(size).To(Equal("18KB"))
					Expect(err).To(BeNil())
				})

				It("Should return total size of 150 files within root persist dir", func() {
					clipsDirPath := "/testroot/clips/TestConnectionInstance"

					By("Creating 150 files on disk of size 1MB")
					for i := 0; i < 150; i++ {
						binFile, err := mockFs.Create(filepath.Join(clipsDirPath, fmt.Sprintf("mock%d.bin", i)))
						Expect(err).To(BeNil())
						defer binFile.Close()
						err = binFile.Truncate(KB * KB)
						Expect(err).To(BeNil())
					}

					By("Querying disk size with all just created files on disk")
					size, err := conn.SizeOnDisk()
					Expect(size).To(Equal("150MB"))
					Expect(err).To(BeNil())
				})
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
						time.Sleep(50 * time.Millisecond)
						runningBufferLength = len(conn.Buffer())
						readMat = <-conn.Buffer()
						cancelStreaming()
					}()

					Eventually(stopping).Should(BeClosed())
					// make sure the buffer is filled max to 6
					Expect(runningBufferLength).To(BeNumerically("==", 6))
					// make sure mat read matches mat written
					Expect(readMat.Sum().Val1).To(BeNumerically("==", matSumVal1))
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
					Expect(openVideoCaptureCallCount).To(BeNumerically("==", 10))
					Expect(closeCallCount).To(BeNumerically("==", 10))
				})
			})
		})
	})
})
