package api

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/tauraamui/dragondaemon/common"
	"github.com/tauraamui/dragondaemon/media"
)

func init() {
	rpc.Register(Session{})
}

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort string
}

type Session struct {
	Token string
}

func (s Session) GetToken(args string, resp *string) error {
	*resp = s.Token
	return nil
}

type MediaServer struct {
	interrupt     chan os.Signal
	s             *media.Server
	httpServer    *http.Server
	rpcListenPort string
}

func New(interrupt chan os.Signal, server *media.Server, opts Options) *MediaServer {
	return &MediaServer{
		interrupt:     interrupt,
		s:             server,
		httpServer:    &http.Server{},
		rpcListenPort: opts.RPCListenPort,
	}
}

func StartRPC(m *MediaServer) error {
	err := rpc.Register(m)
	if err != nil {
		return err
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", m.rpcListenPort)
	if err != nil {
		return err
	}

	errs := make(chan error)
	go func() {
		httpErr := m.httpServer.Serve(l)
		if httpErr != nil {
			errs <- httpErr
		}
		errs <- nil
	}()

	select {
	case err := <-errs:
		return err
	default:
		return nil
	}
}

func ShutdownRPC(m *MediaServer) error {
	return m.httpServer.Close()
}

// Exposed API
func (m *MediaServer) ActiveConnections(sess *Session, resp *[]common.ConnectionData) error {
	*resp = m.s.APIFetchActiveConnections()
	return nil
}

func (m *MediaServer) Shutdown(sess *Session, resp *bool) error {
	*resp = true
	defer func() {
		time.Sleep(time.Second * 1)
		m.interrupt <- os.Interrupt
	}()
	return nil
}
