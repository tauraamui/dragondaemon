package media

import (
	"sync/atomic"
)

func (s *Server) closeConnectionsLocked() error {
	var err error
	for conn := range s.connections {
		if cerr := (*conn).Close(); cerr != nil && err == nil {
			err = cerr
		}
		delete(s.connections, conn)
	}
	return err
}

func (s *Server) shuttingDown() bool {
	return atomic.LoadInt32(&s.inShutdown) != 0
}

func (s *Server) Shutdown() error {
	atomic.StoreInt32(&s.inShutdown, 1)
	return s.closeConnectionsLocked()
}
