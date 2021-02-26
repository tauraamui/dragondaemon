package media

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/tacusci/logging/v2"
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
				Size: func(c *Connection) string {
					size, unit, err := c.SizeOnDisk()
					if err != nil {
						return "N/A"
					}
					logging.Debug("SIZE -> CONN: %s, SIZE: %d%s", connPtr.uuid, size, unit)
					return fmt.Sprintf("%s%s", strconv.FormatInt(size, 10), unit)
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
