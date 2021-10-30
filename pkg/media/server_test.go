package media

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tauraamui/xerror"
)

type testMockProcess struct {
	stopCallback func()
	waitCallback func()
}

func (t *testMockProcess) stop() {
	t.stopCallback()
}

func (t *testMockProcess) wait() {
	t.waitCallback()
}

var _ = Describe("Server", func() {
	Describe("Unit tests for instance methods", func() {
		var server *Server
		BeforeEach(func() {
			server = NewServer(false)
			Expect(server).ToNot(BeNil())
		})

		Context("Calling server run directly", func() {
			It("Should call beginProcesses method", func() {
				var passedServer *Server
				resetBeginProcesses := OverloadBeginProcesses(func(c context.Context, o Options, s *Server) []processable {
					passedServer = s
					return []processable{}
				})
				defer resetBeginProcesses()

				server.Run(Options{})
				<-server.Shutdown()

				Expect(passedServer).To(Equal(server))
			})

			It("Should call stop and wait in that order on processes returned from beginProcess", func() {
				var stopCalled bool
				var waitCalled bool

				resetBeginProcesses := OverloadBeginProcesses(func(c context.Context, o Options, s *Server) []processable {
					return []processable{
						&testMockProcess{
							stopCallback: func() { stopCalled = !waitCalled },
							waitCallback: func() { waitCalled = stopCalled },
						},
					}
				})
				defer resetBeginProcesses()

				server.Run(Options{})
				<-server.Shutdown()

				Expect(stopCalled).To(BeTrue())
				Expect(waitCalled).To(BeTrue())
			})
		})

		Context("Server connect opens a connection and tracks it", func() {
			It("Should make mock connection and store instance", func() {
				infoLogs := []string{}
				resetLogInfo := OverloadLogInfo(func(format string, args ...interface{}) {
					infoLogs = append(infoLogs, fmt.Sprintf(format, args...))
				})
				defer resetLogInfo()

				server.Connect(
					"TestConnection",
					"fake-rtsp-addr",
					ConnectonSettings{
						MockWriter:      true,
						MockCapturer:    true,
						PersistLocation: "/testroot/clips",
					},
				)

				Expect(server.connections).To(HaveLen(1))
				Expect(infoLogs[0]).To(ContainSubstring(
					fmt.Sprintf(
						"Connected to stream [%s] at [%s]",
						"TestConnection", "fake-rtsp-addr",
					),
				))
			})

			It("Should error on connecting to stream and not track a connection", func() {
				resetOpenVidCapOverload := OverloadOpenVideoCapture(
					func(string, string, int, bool, string, bool) (VideoCapturable, error) {
						return nil, xerror.New("test fail to open connection")
					},
				)
				defer resetOpenVidCapOverload()

				infoLogs := []string{}
				resetLogInfo := OverloadLogInfo(func(format string, args ...interface{}) {
					infoLogs = append(infoLogs, fmt.Sprintf(format, args...))
				})
				defer resetLogInfo()

				errorLogs := []string{}
				resetLogError := OverloadLogError(func(format string, args ...interface{}) {
					errorLogs = append(errorLogs, fmt.Sprintf(format, args...))
				})
				defer resetLogError()

				server.Connect(
					"TestConnection",
					"fake-rtsp-addr",
					ConnectonSettings{
						MockWriter:      true,
						MockCapturer:    true,
						PersistLocation: "/testroot/clips",
					},
				)

				Expect(server.connections).To(HaveLen(0))
				Expect(infoLogs).To(HaveLen(0))
				Expect(errorLogs).To(HaveLen(1))
				Expect(errorLogs[0]).To(ContainSubstring(
					fmt.Sprintf(
						"Unable to connect to stream [%s] at [%s]",
						"TestConnection", "fake-rtsp-addr",
					),
				))
			})
		})
	})
})
