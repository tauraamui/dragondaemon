package media

import (
	"fmt"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
)

func (s *Server) trackConnection(conn *Connection, add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections == nil {
		s.connections = make(map[*Connection]struct{})
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

func (s *Server) Connect(
	title string,
	rtspStream string,
	persistLocation string,
	secondsPerClip int,
) (*Connection, error) {
	vc, err := gocv.OpenVideoCapture(rtspStream)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to connect to stream at [%s]: %v", rtspStream, err))
		return nil, err
	}

	logging.Info(fmt.Sprintf("Connected to stream at [%s]", rtspStream))
	conn := NewConnection(
		title,
		persistLocation,
		secondsPerClip,
		vc,
	)
	s.trackConnection(conn, true)

	return conn, nil
}
