package media

type ConnectionData struct {
	uuid, title string
}

func (c ConnectionData) UUID() string {
	return c.uuid
}

func (c ConnectionData) Title() string {
	return c.title
}

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []ConnectionData {
	connections := []ConnectionData{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, ConnectionData{
				uuid:  connPtr.uuid,
				title: connPtr.title,
			})
		}
	}
	return connections
}
