package media

import (
	"errors"
	"fmt"
	"strconv"

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
					s, err := c.SizeOnDisk()
					if err != nil {
						return "N/A"
					}
					return fmt.Sprintf("%sMb", strconv.FormatInt(s, 10))
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
