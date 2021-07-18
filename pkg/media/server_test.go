package media

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("Unit tests for instance methods", func() {
		var server *Server
		BeforeEach(func() {
			server = NewServer(false)
			Expect(server).ToNot(BeNil())
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
		})
	})
})
