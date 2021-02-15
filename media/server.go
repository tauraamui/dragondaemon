package media

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gocv.io/x/gocv"
)

// Server manages receiving RTSP streams and persisting clips to disk
type Server struct {
	inShutdown        int32
	mu                sync.Mutex
	wg                sync.WaitGroup
	t                 *time.Ticker
	stopStreaming     chan struct{}
	stoppedStreaming  chan struct{}
	stopRemovingClips chan struct{}
	connections       map[*Connection]struct{}
}

// NewServer returns a pointer to media server instance
func NewServer() *Server {
	return &Server{}
}

func (s *Server) IsRunning() bool {
	return !s.shuttingDown()
}

func (s *Server) Connect(
	title string,
	rtspStream string,
	persistLocation string,
	fps int,
	secondsPerClip int,
	schedule schedule.Schedule,
) {
	vc, err := gocv.OpenVideoCapture(rtspStream)
	vc.Set(gocv.VideoCaptureFPS, float64(fps))
	if err != nil {
		logging.Error("Unable to connect to stream [%s] at [%s]", title, rtspStream)
		return
	}

	logging.Info("Connected to stream [%s] at [%s]", title, rtspStream)
	conn := NewConnection(
		title,
		persistLocation,
		fps,
		secondsPerClip,
		schedule,
		vc,
		rtspStream,
	)
	s.trackConnection(conn, true)
}

func (s *Server) BeginStreaming() {
	s.wg = sync.WaitGroup{}
	s.stopStreaming = make(chan struct{})
	for _, conn := range s.activeConnections() {
		logging.Info("Reading stream from connection [%s]", conn.title)
		s.stoppedStreaming = conn.stream(s.stopStreaming)
	}
}

func (s *Server) RemoveOldClips(maxClipAgeInDays int) {
	if s.t == nil {
		s.t = time.NewTicker(1 * time.Minute)
	}

	if s.stopRemovingClips == nil {
		s.stopRemovingClips = make(chan struct{})
	}

	var currentConnection int
	for {
		time.Sleep(time.Millisecond * 10)
		select {
		case <-s.t.C:
			activeConnections := s.activeConnections()
			if currentConnection >= len(activeConnections) {
				currentConnection = 0
			}

			if conn := activeConnections[currentConnection]; conn != nil {
				fullPersistLocation := fmt.Sprintf("%s%c%s", conn.persistLocation, os.PathSeparator, conn.title)
				files, err := ioutil.ReadDir(fullPersistLocation)
				if err != nil {
					logging.Error("Unable to read contents of connection persist location %s", fullPersistLocation)
				}

				for _, file := range files {
					date, err := time.Parse("2006-01-02", file.Name())
					if err != nil {
						continue
					}

					oldestAllowedDay := time.Now().AddDate(0, 0, -1*maxClipAgeInDays)
					if date.Before(oldestAllowedDay) {
						dirToRemove := fmt.Sprintf("%s%c%s", fullPersistLocation, os.PathSeparator, file.Name())
						logging.Info("REMOVING DIR %s", dirToRemove)
						err := os.RemoveAll(dirToRemove)
						if err != nil {
							logging.Error("Failed to RemoveAll %s", dirToRemove)
						}
					}
				}
			}

			currentConnection++
		case <-s.stopRemovingClips:
			s.t.Stop()
			return
		}
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
	close(s.stopRemovingClips)
	<-s.stoppedStreaming
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
