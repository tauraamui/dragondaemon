package media

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []Connection {
	connections := []Connection{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, Connection{
				uuid:  connPtr.uuid,
				title: connPtr.title,
			})
		}
	}
	return connections
}
