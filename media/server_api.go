package media

import mediaapi "github.com/tauraamui/dragondaemon/api/media"

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []mediaapi.Connection {
	connections := []mediaapi.Connection{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, &Connection{
				uuid:  connPtr.uuid,
				title: connPtr.title,
			})
		}
	}
	return connections
}
