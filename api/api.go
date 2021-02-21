package api

import (
	"github.com/tauraamui/dragondaemon/media"
)

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

type Connection interface {
	UUID() string
	Title() string
}

type connectionData struct {
	uuid, title string
}

func (c connectionData) UUID() string  { return c.uuid }
func (c connectionData) Title() string { return c.title }

func (i *Instance) ActiveConnections() []Connection {
	connections := []Connection{}
	for _, conn := range i.s.APIFetchActiveConnections() {
		connections = append(connections, connectionData{
			uuid:  conn.UUID(),
			title: conn.Title(),
		})
	}
	return connections
}
