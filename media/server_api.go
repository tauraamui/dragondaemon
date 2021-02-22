package media

import "net/rpc"

func init() {
	rpc.Register(ConnectionData{})
}

type ConnectionData struct {
	UUID, Title string
}

func (c ConnectionData) GetUUID(args string, dst *string) error {
	*dst = c.UUID
	return nil
}

func (c ConnectionData) GetTitle(args string, dst *string) error {
	*dst = c.Title
	return nil
}

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []ConnectionData {
	connections := []ConnectionData{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, ConnectionData{
				connPtr.uuid,
				connPtr.title,
			})
		}
	}
	return connections
}
