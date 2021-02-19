package api

import "github.com/tauraamui/dragondaemon/media"

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort int
}

type Instance struct {
	s *media.Server
}

func New(server *media.Server) *Instance {
	return &Instance{s: server}
}

func (i *Instance) ActiveConnections() []media.Connection {
	return i.s.APIFetchActiveConnections()
}
