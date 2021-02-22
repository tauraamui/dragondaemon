package api

import (
	"github.com/tauraamui/dragondaemon/media"
)

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort int
}

type MediaServer struct {
	s *media.Server
}

func New(server *media.Server) *MediaServer {
	return &MediaServer{s: server}
}

func (i *MediaServer) ActiveConnections(args string, resp *[]media.ConnectionData) error {
	*resp = i.s.APIFetchActiveConnections()
	return nil
}
