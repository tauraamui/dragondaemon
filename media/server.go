package media

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gocv.io/x/gocv"
)

// Server manages receiving RTSP streams and persisting clips to disk
type Server struct {
	inShutdown    int32
	mu            sync.Mutex
	stopStreaming chan struct{}
	connections   map[*Connection]struct{}
}

// NewServer returns a pointer to media server instance
func NewServer() *Server {
	return &Server{}
}

func (s *Server) IsRunning() bool {
	time.Sleep(time.Millisecond * 10)
	return !s.shuttingDown()
}

func (s *Server) Connect(
	stdlog, errlog *log.Logger,
	title string,
	rtspStream string,
	persistLocation string,
	fps int,
	secondsPerClip int,
) {
	vc, err := gocv.OpenVideoCapture(rtspStream)
	if err != nil {
		errlog.Printf("Unable to connect to stream [%s] at [%s]: %v\n", title, rtspStream, err)
		return
	}

	stdlog.Printf("Connected to stream [%s] at [%s]\n", title, rtspStream)
	conn := NewConnection(
		stdlog, errlog,
		title,
		persistLocation,
		fps,
		secondsPerClip,
		vc,
	)
	s.trackConnection(conn, true)
}

func (s *Server) BeginStreaming() {
	s.stopStreaming = make(chan struct{})
	for _, conn := range s.activeConnections() {
		go conn.stream(s.stopStreaming)
	}
}

func (s *Server) SaveStreams(rwg *sync.WaitGroup) {
	if rwg != nil {
		rwg.Add(1)
	}
	defer func() {
		if rwg != nil {
			rwg.Done()
		}
	}()
	wg := sync.WaitGroup{}
	for s.IsRunning() {
		start := make(chan struct{})
		for _, conn := range s.activeConnections() {
			wg.Add(1)
			go func(conn *Connection) {
				// immediately pause thread
				<-start
				// save 1-3 seconds worth of footage to clip file
				conn.persistToDisk()
				wg.Done()
			}(conn)
		}
		// unpause all threads at the same time
		close(start)
		wg.Wait()
	}
}

func (s *Server) Shutdown() {
	atomic.StoreInt32(&s.inShutdown, 1)
}

func (s *Server) Close() error {
	close(s.stopStreaming)
	time.Sleep(time.Millisecond * 10)
	return s.closeConnectionsLocked()
}

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

func (s *Server) activeConnections() []*Connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	connections := make([]*Connection, 0, len(s.connections))
	for k := range s.connections {
		connections = append(connections, k)
	}

	return connections
}

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
