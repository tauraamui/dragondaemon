package media_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/media"
	"gocv.io/x/gocv"
)

const KB = 1024

type testMockVideoCapture struct {
	isOpenedFunc func() bool
	readFunc     func(m *gocv.Mat) bool
	closeFunc    func() error
}

// SetP doesn't do anything it exists to satisfy VideoCapturable interface
func (tmvc *testMockVideoCapture) SetP(_ *gocv.VideoCapture) {}

// IsOpened always returns true.
func (tmvc *testMockVideoCapture) IsOpened() bool {
	if tmvc.isOpenedFunc != nil {
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

type testMockVideoWriter struct {
	isOpenedFunc func() bool
	writeFunc    func(m gocv.Mat) error
	closeFunc    func() error
}

func (tmvw *testMockVideoWriter) SetP(_ *gocv.VideoWriter) {}

func (tmvw *testMockVideoWriter) IsOpened() bool {
	if tmvw.isOpenedFunc != nil {
		return tmvw.isOpenedFunc()
	}
	panic(errors.New("call to missing test mock video writer isOpened function"))
}

func (tmvw *testMockVideoWriter) Write(m gocv.Mat) error {
	if tmvw.writeFunc != nil {
		return tmvw.writeFunc(m)
	}
	panic(errors.New("call to missing test mock video writer write function"))
}

func (tmvw *testMockVideoWriter) Close() error {
	if tmvw.closeFunc != nil {
		return tmvw.closeFunc()
	}
	panic(errors.New("call to missing test mock video writer close function"))
}

var _ = Describe("Connection", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel

	var mockFs afero.Fs = nil
	var resetFs func() = nil

	BeforeSuite(func() {
		mockFs = afero.NewMemMapFs()
		resetFs = media.OverloadFS(mockFs)
	})

	AfterEach(func() {
		Expect(mockFs.RemoveAll("*")).To(BeNil())
		logging.CurrentLoggingLevel = logging.SilentLevel
	})

	AfterSuite(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
		resetFs()
		mockFs = nil
	})

	Context("NewConnection", func() {
		It("Should return a new connection instance", func() {
			Expect(mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)).To(BeNil())
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
			resetErrorLog := media.OverloadLogError(func(format string, _ ...interface{}) {
				Expect(format).To(ContainSubstring(
					"unable to connect to camera API: Post \"http://fake-reolink-api/cgi-bin/api.cgi?cmd=Login&token=\":",
				))
			})
			defer resetErrorLog()

			Expect(mockFs.MkdirAll("/testroot/clips/TestConnection", os.ModeDir|os.ModePerm)).To(BeNil())
			conn := media.NewConnection(
				"TestConnection",
				media.ConnectonSettings{
					PersistLocation: "/testroot/clips",
					FPS:             30,
					SecondsPerClip:  2,
					Schedule:        schedule.Schedule{},
					Reolink: config.ReolinkAdvanced{
						APIAddress: "fake-reolink-api",
						Enabled:    true,
					},
				},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			Expect(conn.ReolinkControl()).To(BeNil())
		})

		It("Should return connection instance with missing cache", func() {
			var errorLogs []string
			resetErrorLog := media.OverloadLogError(func(format string, args ...interface{}) {
				errorLogs = append(errorLogs, fmt.Sprintf(format, args...))
			})
			defer resetErrorLog()

			resetInitCache := media.OverloadNewCache(func() (*bigcache.BigCache, error) {
				return nil, errors.New("test error: unable to init cache")
			})
			defer resetInitCache()

			clipsDirPath := "/testroot/clips/TestConnection"
			Expect(mockFs.MkdirAll(clipsDirPath, os.ModeDir|os.ModePerm)).To(BeNil())

			By("Creating file on disk of size 9KB")
			binFile, err := mockFs.Create(filepath.Join(clipsDirPath, "mock.bin"))
			Expect(err).To(BeNil())
			defer binFile.Close()
			err = binFile.Truncate(KB * 9)
			Expect(err).To(BeNil())
			conn := media.NewConnection(
				"TestConnection",
				media.ConnectonSettings{
					PersistLocation: "/testroot/clips",
					FPS:             30,
					SecondsPerClip:  2,
					Schedule:        schedule.Schedule{},
					Reolink: config.ReolinkAdvanced{
						Enabled: false,
					},
				},
				&testMockVideoCapture{},
				"fake-stream-addr",
			)

			size, err := conn.SizeOnDisk()
			Expect(err).To(MatchError("nil pointer to cache"))
			Expect(size).To(Equal("9KB"))

			Expect(conn.Cache()).To(BeNil())
			Expect(errorLogs).To(HaveLen(2))
			Expect(errorLogs[0]).To(Equal(
				"unable to initialise connection cache: test error: unable to init cache",
			))
			Expect(errorLogs[1]).To(ContainSubstring("unable to load disk size from cache"))
		})
	})

	Context("Using a connection instance", func() {
		var conn *media.Connection
		var videoCapture *testMockVideoCapture
		var resetVidCapOverload func()
		var openVidCapCallback func()

		BeforeEach(func() {
			videoCapture = &testMockVideoCapture{
				// TODO(tauraamui): Move this to above definition as default
				/*
					Will always need to be present so force add always.
					Some tests will want to not return nil, but it must
					at least exist as a function def all the time.
					This should be handled by the mock definition really.
				*/
				isOpenedFunc: func() bool { return true },
				closeFunc:    func() error { return nil },
			}

			Expect(mockFs.MkdirAll("/testroot/clips/TestConnectionInstance", os.ModeDir|os.ModePerm)).To(BeNil())
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
					Expect(mockFs.MkdirAll(clipsDirPath, os.ModeDir|os.ModePerm)).To(BeNil())

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
					Expect(mockFs.Remove(binFile.Name())).To(BeNil())

					size, err = conn.SizeOnDisk()
					Expect(size).To(Equal("9KB"))
					Expect(err).To(BeNil())
				})

				It("Should return total size on disk including sub dirs within persist dir", func() {
					clipsRootDirPath := "/testroot/clips/TestConnectionInstance"

					clipsSubDirPath1 := "/testroot/clips/TestConnectionInstance/subdir1"
					Expect(mockFs.MkdirAll(clipsSubDirPath1, os.ModeDir|os.ModePerm)).To(BeNil())

					clipsSubDirPath2 := "/testroot/clips/TestConnectionInstance/subdir2"
					Expect(mockFs.MkdirAll(clipsSubDirPath2, os.ModeDir|os.ModePerm)).To(BeNil())

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
		})

		Context("Calling connection readFromStream directly", func() {
			It("Should return true from readFromStream", func() {
				videoCapture.isOpenedFunc = func() bool {
					return true
				}
				readCallCount := 0
				videoCapture.readFunc = func(m *gocv.Mat) bool {
					readCallCount++
					return true
				}
				mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
				defer mat.Close()
				Expect(media.ReadFromStream(conn, &mat)).To(BeTrue())
				Expect(readCallCount).To(BeNumerically("==", 1))
			})

			It("Should return false from readFromStream", func() {
				videoCapture.isOpenedFunc = func() bool {
					return false
				}
				mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
				defer mat.Close()
				Expect(media.ReadFromStream(conn, &mat)).To(BeFalse())
			})
		})

		Context("Connection streaming frames to channel", func() {
			It("Should read video frames from given connection into buffer channel", func() {
				var matSumVal1 float64
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
					time.Sleep(80 * time.Millisecond)
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
				infoLogs := []string{}
				resetLogInfo := media.OverloadLogInfo(func(format string, args ...interface{}) {
					infoLogs = append(infoLogs, fmt.Sprintf(format, args...))
				})
				defer resetLogInfo()

				failedToCloseErrorLogs := []string{}
				unableToReconnectErrorLogs := []string{}
				otherErrorLogs := []string{}
				resetLogError := media.OverloadLogError(func(format string, args ...interface{}) {
					errorLog := fmt.Sprintf(format, args...)
					if strings.Contains(
						errorLog, "Failed to close connection... ERROR: test connection close error",
					) {
						failedToCloseErrorLogs = append(failedToCloseErrorLogs, errorLog)
						return
					} else if strings.Contains(errorLog, "Unable to reconnect to") {
						unableToReconnectErrorLogs = append(unableToReconnectErrorLogs, errorLog)
						return
					}
					otherErrorLogs = append(otherErrorLogs, errorLog)
				})
				defer resetLogError()

				processCallCount := 0
				ableToRead := true
				videoCapture.closeFunc = func() error {
					return errors.New("test connection close error")
				}
				videoCapture.readFunc = func(m *gocv.Mat) bool {
					processCallCount++
					return ableToRead
				}

				makeVidCapReturnFalseFromRead := func() func() { ableToRead = false; return func() { ableToRead = true } }
				makeOpenVidCapReturnError := func() func() {
					return media.OverloadOpenVideoCapture(
						func(string, string, int, bool, string) (media.VideoCapturable, error) {
							processCallCount++
							return nil, errors.New("test unable to open video connection")
						},
					)
				}

				ctx, cancelStreaming := context.WithCancel(context.Background())
				stopping := conn.Stream(ctx)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					madeVidCapReturnFalse := false
					madeOpenVidCapReturnError := false

					var resetMakeVidCapReturnFalse, resetMakeOpenVidCapReturnError func()
					defer wg.Done()

					for processCallCount < 50 {
						// initial 20 reads read as normal

						// after 20 reads and less than 30 reads make read fail
						if processCallCount > 20 && processCallCount < 30 {
							if !madeVidCapReturnFalse {
								resetMakeVidCapReturnFalse = makeVidCapReturnFalseFromRead()
								madeVidCapReturnFalse = true
							}

							if !madeOpenVidCapReturnError {
								resetMakeOpenVidCapReturnError = makeOpenVidCapReturnError()
								madeOpenVidCapReturnError = true
							}
						}

						if processCallCount > 30 {
							if madeVidCapReturnFalse && resetMakeVidCapReturnFalse != nil {
								resetMakeVidCapReturnFalse()
								madeVidCapReturnFalse = false
							}
							if madeOpenVidCapReturnError && resetMakeOpenVidCapReturnError != nil {
								resetMakeOpenVidCapReturnError()
								madeOpenVidCapReturnError = false
							}
						}
					}
					cancelStreaming()
				}(&wg)

				wg.Wait()
				Eventually(stopping).Should(BeClosed())
				Expect(failedToCloseErrorLogs).To(HaveLen(10))
				Expect(failedToCloseErrorLogs).To(HaveCap(16))
				Expect(unableToReconnectErrorLogs).To(HaveLen(9))
				Expect(unableToReconnectErrorLogs).To(HaveCap(16))
				Expect(infoLogs).To(HaveLen(11))
				Expect(infoLogs).To(HaveCap(16))
				for i := 0; i < 10; i++ {
					Expect(infoLogs[i]).To(Equal("Attempting to reconnect to [TestConnectionInstance]"))
				}
				Expect(infoLogs[10]).To(Equal("Re-connected to [TestConnectionInstance]..."))
			})
		})

		Context("Connection writing stream to disk", func() {
			var videoWriter testMockVideoWriter
			var resetVidWriterOverload func()

			BeforeEach(func() {
				resetVidWriterOverload = media.OverloadOpenVideoWriter(
					func(string, string, float64, int, int, bool) (media.VideoWriteable, error) {
						return &videoWriter, nil
					},
				)
			})

			AfterEach(func() { resetVidWriterOverload() })

			It("Should read same frame data from stream as written when calling save", func() {
				// improve behaviour test by verifying frames given to writer
				// match expected number of clip's total frames to write
				var sentMatSumVal1 float64
				videoCapture.isOpenedFunc = func() bool { return true }
				videoCapture.readFunc = func(m *gocv.Mat) bool {
					mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
					defer mat.Close()
					mat.AddFloat(11.54)
					sentMatSumVal1 = mat.Sum().Val1
					mat.CopyTo(m)
					return true
				}

				var readMatSumVal1 float64
				videoWriter.isOpenedFunc = func() bool { return true }
				videoWriter.writeFunc = func(m gocv.Mat) error {
					readMatSumVal1 = m.Sum().Val1
					return nil
				}
				videoWriter.closeFunc = func() error { return nil }

				ctx, cancelStreaming := context.WithCancel(context.Background())
				stoppingStreaming := conn.Stream(ctx)

				ctx, cancelWriteStreamToClips := context.WithCancel(context.Background())
				stoppingWriteStreamIntoClips := conn.WriteStreamToClips(ctx)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					time.Sleep(10 * time.Millisecond)
					wg.Done()
				}(&wg)

				wg.Wait()

				cancelWriteStreamToClips()
				Eventually(stoppingWriteStreamIntoClips).Should(BeClosed())

				cancelStreaming()
				Eventually(stoppingStreaming).Should(BeClosed())

				Expect(sentMatSumVal1).To(BeNumerically("==", readMatSumVal1))
			})

			// TODO(tauraamui): Fix or remove this test
			It("Should use real video writer and save footage into clips in correct dir on disk", func() {
				os.Setenv("DRAGON_DAEMON_MOCK_VIDEO_STREAM", "1")
				timeNow, err := time.Parse("2006-01-02 15.04.05", "2021-02-02 10.00.00")
				resetNowOverload := media.OverloadNow(func() time.Time {
					timeNow = timeNow.Add(time.Second * 1)
					return timeNow
				})
				defer resetNowOverload()

				Expect(err).To(BeNil())
				Expect(timeNow).ToNot(BeNil())

				errorLogs := []string{}
				resetLogError := media.OverloadLogError(func(format string, args ...interface{}) {
					errorLogs = append(errorLogs, fmt.Sprintf(format, args...))
				})
				defer resetLogError()

				// make video writer and capturer use real implementation
				resetVidWriterOverload()
				resetVidCapOverload()

				videoCapture.readFunc = func(m *gocv.Mat) bool {
					mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
					defer mat.Close()
					// mat.AddFloat(11.54)
					mat.CopyTo(m)
					return true
				}

				ctx, cancelStreaming := context.WithCancel(context.Background())
				stoppingStreaming := conn.Stream(ctx)

				ctx, cancelWriteStreamToClips := context.WithCancel(context.Background())
				stoppingWriteStreamIntoClips := conn.WriteStreamToClips(ctx)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					time.Sleep(800 * time.Millisecond)
					wg.Done()
				}(&wg)

				wg.Wait()

				cancelWriteStreamToClips()
				Eventually(stoppingWriteStreamIntoClips).Should(BeClosed())

				cancelStreaming()
				Eventually(stoppingStreaming).Should(BeClosed())

				Expect(errorLogs).To(HaveLen(0))
				clipsDir, err := mockFs.Open("/testroot/clips/TestConnectionInstance/2021-02-02")
				Expect(err).To(BeNil())
				Expect(clipsDir).ToNot(BeNil())

				files, err := clipsDir.Readdir(-1)
				Expect(err).To(BeNil())
				Expect(files).To(HaveLen(0))
			})

			It("Should fail to open video writer and therefore not write to disk", func() {
				errorLogs := []string{}
				resetErrorLog := media.OverloadLogError(func(format string, args ...interface{}) {
					errorLogs = append(errorLogs, fmt.Sprintf(format, args...))
				})
				defer resetErrorLog()

				videoCapture.isOpenedFunc = func() bool { return true }
				videoCapture.readFunc = func(m *gocv.Mat) bool {
					mat := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV32F)
					defer mat.Close()
					mat.CopyTo(m)
					return true
				}

				resetVidWriterOverload = media.OverloadOpenVideoWriter(
					func(string, string, float64, int, int, bool) (media.VideoWriteable, error) {
						return nil, errors.New("test error: unable to open video writer")
					},
				)
				defer resetVidWriterOverload()

				ctx, cancelStreaming := context.WithCancel(context.Background())
				stoppingStreaming := conn.Stream(ctx)

				ctx, cancelWriteStreamToClips := context.WithCancel(context.Background())
				stoppingWriteStreamIntoClips := conn.WriteStreamToClips(ctx)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					time.Sleep(3 * time.Millisecond)
					wg.Done()
				}(&wg)

				wg.Wait()

				cancelWriteStreamToClips()
				Eventually(stoppingWriteStreamIntoClips, 3*time.Second).Should(BeClosed())

				cancelStreaming()
				Eventually(stoppingStreaming).Should(BeClosed())

				Expect(errorLogs).To(HaveLen(1))
				Expect(errorLogs).To(HaveCap(1))
				Expect(errorLogs[0]).To(ContainSubstring(
					"Unable to write video clip /testroot/clips/TestConnectionInstance/",
				))
				Expect(errorLogs[0]).To(ContainSubstring(
					".mp4 to disk: test error: unable to open video writer",
				))
			})
		})
	})
})
