package media

import (
	"errors"

	"github.com/tauraamui/dragondaemon/common"
)

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []common.ConnectionData {
	connections := []common.ConnectionData{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, common.ConnectionData{
				UUID:  connPtr.uuid,
				Title: connPtr.title,
			})
		}
	}
	return connections
}

func (s *Server) APIRestartConnection(cameraUUID string) error {
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil && connPtr.reolinkControl != nil {
			if connPtr.uuid == cameraUUID {
				_, err := connPtr.reolinkControl.RebootCamera()(connPtr.reolinkControl.RestHandler)
				return err
			}
		}
	}

	return errors.New("unable to find Reolink contoller for any connection")
}
