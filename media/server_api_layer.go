package media

type apiConnection struct {
	UUID  string
	Title string
}

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []apiConnection {
	connections := []apiConnection{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, apiConnection{
				UUID:  connPtr.uuid,
				Title: connPtr.title,
			})
		}
	}
	return connections
}
