package media

import "github.com/tauraamui/dragondaemon/common"

type connectionData struct {
	uuid, title string
}

func (c connectionData) UUID() string {
	return c.uuid
}

func (c connectionData) Title() string {
	return c.title
}

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []common.ConnectionData {
	connections := []common.ConnectionData{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, connectionData{
				uuid:  connPtr.uuid,
				title: connPtr.title,
			})
		}
	}
	return connections
}
