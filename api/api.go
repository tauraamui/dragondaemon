package api

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/tauraamui/dragondaemon/common"
	"github.com/tauraamui/dragondaemon/media"
)

func init() {
	rpc.Register(Session{})
}

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort int
}

type Session struct {
	Token string
}

func (s Session) GetToken(args string, resp *string) error {
	*resp = s.Token
	return nil
}

type MediaServer struct {
	s             *media.Server
	httpServer    *http.Server
	rpcListenPort int
}

func New(server *media.Server, opts Options) *MediaServer {
	return &MediaServer{s: server, rpcListenPort: opts.RPCListenPort}
}

func StartRPC(m *MediaServer) error {
	err := rpc.Register(m)
	if err != nil {
		return err
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", m.rpcListenPort))
	if err != nil {
		return err
	}

	errs := make(chan error)
	go func() {
		m.httpServer = &http.Server{}
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

func (m *MediaServer) ActiveConnections(sess *Session, resp *[]common.ConnectionData) error {
	*resp = m.s.APIFetchActiveConnections()
	return nil
}
