package api

import (
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

type MediaServer struct {
	s *media.Server
}

type Session struct {
	Token string
}

func (s Session) GetToken(args string, resp *string) error {
	*resp = s.Token
	return nil
}

func New(server *media.Server) *MediaServer {
	return &MediaServer{s: server}
}

func (i *MediaServer) ActiveConnections(sess *Session, resp *[]common.ConnectionData) error {
	*resp = i.s.APIFetchActiveConnections()
	return nil
}
