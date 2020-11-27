package media

import (
	"sync"

	"gocv.io/x/gocv"
)

// Server manages receiving RTSP streams and persisting clips to disk
type Server struct {
	inShutdown  int32
	mu          sync.Mutex
	connections map[*gocv.VideoCapture]struct{}
}

// NewServer returns a pointer to media server instance
func NewServer() *Server {
	return &Server{}
}

func (s *Server) IsRunning() bool {
	return !s.shuttingDown()
}
