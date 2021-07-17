package media

import (
	"errors"

	"github.com/tauraamui/dragondaemon/common"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

// APIFetchActiveConnections returns list of current active connection titles
func (s *Server) APIFetchActiveConnections() []common.ConnectionData {
	connections := []common.ConnectionData{}
	for _, connPtr := range s.activeConnections() {
		if connPtr != nil {
			connections = append(connections, common.ConnectionData{
				UUID:  connPtr.uuid,
				Title: connPtr.title,
				Size: func(c *Connection) string {
					size, err := c.SizeOnDisk()
					if err != nil {
						return "N/A"
					}
					log.Debug("SIZE -> CONN: %s, SIZE: %s", connPtr.uuid, size) //nolint
					return size
				}(connPtr),
			})
		}
	}
	return connections
}

func (s *Server) APIRebootConnection(cameraUUID string) error {
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
