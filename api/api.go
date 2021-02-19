package api

import "github.com/tauraamui/dragondaemon/media"

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort int
}

type Instance struct {
	s *media.Server
}

type Connection struct {
	UUID  string
	Title string
}

func New(server *media.Server) *Instance {
	return &Instance{s: server}
}

func (i *Instance) ActiveConnections() []Connection {
	connections := []Connection{}
	for _, conn := range i.s.APIFetchActiveConnections() {
		connections = append(connections, Connection{
			UUID:  conn.UUID,
			Title: conn.Title,
		})
	}
	return connections
}
