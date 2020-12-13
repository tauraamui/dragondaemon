package media

import (
	"sync/atomic"
)

func (s *Server) closeConnectionsLocked() error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *Server) Shutdown() {
	atomic.StoreInt32(&s.inShutdown, 1)
}

func (s *Server) Close() error {
	return s.closeConnectionsLocked()
}
