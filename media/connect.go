package media

import (
	"fmt"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
)

func (s *Server) trackConnection(conn *gocv.VideoCapture, add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections == nil {
		s.connections = make(map[*gocv.VideoCapture]struct{})
	}

	if add {
		if s.shuttingDown() {
			return false
		}
		s.connections[conn] = struct{}{}
	} else {
		delete(s.connections, conn)
	}
	return true
}

func (s *Server) Connect(rtspStream string) error {
	conn, err := gocv.OpenVideoCapture(rtspStream)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to connect to stream at [%s]: %v", rtspStream, err))
		return err
	}

	logging.Info(fmt.Sprintf("Connected to stream at [%s]", rtspStream))
	s.trackConnection(conn, true)

	return nil
}
