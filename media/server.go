package media

import (
	"sync"
)

// Server manages receiving RTSP streams and persisting clips to disk
type Server struct {
	inShutdown  int32
	mu          sync.Mutex
	connections map[*Connection]struct{}
}

// NewServer returns a pointer to media server instance
func NewServer() *Server {
	return &Server{}
}

func (s *Server) IsRunning() bool {
	return !s.shuttingDown()
}
