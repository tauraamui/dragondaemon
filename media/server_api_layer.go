package media

// FetchActiveConnections returns list of copies of current active connections
func (s *Server) APIFetchActiveConnections() []Connection {
	connections := []Connection{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, *connPtr)
		}
	}
	return connections
}
