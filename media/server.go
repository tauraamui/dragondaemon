package media

import (
	"context"
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
	inShutdown           int32
	mu                   sync.Mutex
	ctx                  context.Context
	ctxCancel            context.CancelFunc
	stoppedAll           chan struct{}
	stopStreaming        chan struct{}
	stoppedStreaming     []chan struct{}
	stopRemovingClips    chan struct{}
	stoppedRemovingClips chan struct{}
	connections          map[*Connection]struct{}
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

func (s *Server) Run() {
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.stoppedAll = make(chan struct{})

	go func(ctx context.Context, stopped chan struct{}) {
		streamingCtx, cancelStreaming := context.WithCancel(context.Background())
		savingCtx, cancelSaving := context.WithCancel(context.Background())

		// streaming connections is core and is the dependancy to all subsequent processes
		stoppedStreaming := s.beginStreaming(streamingCtx)
		stoppedSaving := s.saveStreams(savingCtx)

		// wait for shutdown signal
		<-ctx.Done()

		cancelSaving()
		logging.Info("Waiting for persist process to finish...")
		// wait for saving streams to stop
		<-stoppedSaving

		// stopping the streaming process should be done last
		// stop all streaming
		cancelStreaming()
		logging.Info("Waiting for streams to close...")
		// wait for all streams to stop
		// TODO(:tauraamui) Move each stream stop signal wait onto separate goroutine
		for _, stoppedStreamSig := range stoppedStreaming {
			<-stoppedStreamSig
		}

		// send signal saying shutdown process has finished
		close(stopped)
	}(s.ctx, s.stoppedAll)
}

func (s *Server) RemoveOldClips(maxClipAgeInDays int) {
	if s.stopRemovingClips == nil {
		s.stopRemovingClips = make(chan struct{})
	}

	s.stoppedRemovingClips = s.removeOldClips(maxClipAgeInDays, s.stopRemovingClips)

}

func (s *Server) beginStreaming(ctx context.Context) []chan struct{} {
	var stoppedStreaming []chan struct{}
	for _, conn := range s.activeConnections() {
		logging.Info("Reading stream from connection [%s]", conn.title)
		stoppedStreaming = append(stoppedStreaming, conn.stream(ctx))
	}
	return stoppedStreaming
}

func (s *Server) saveStreams(ctx context.Context) chan struct{} {
	stopping := make(chan struct{})

	reachedShutdownCase := false
	go func(ctx context.Context, stopping chan struct{}) {
		for {
			time.Sleep(time.Millisecond * 1)
			select {
			case <-ctx.Done():
				// TODO(:tauraamui) Investigate why this case is reached more than once anyway
				if reachedShutdownCase == false {
					reachedShutdownCase = true
					close(stopping)
					break
				}
				break
			default:
				start := make(chan struct{})
				wg := sync.WaitGroup{}
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
	}(ctx, stopping)

	return stopping
}

func (s *Server) Shutdown() chan struct{} {
	atomic.StoreInt32(&s.inShutdown, 1)
	s.ctxCancel()
	return s.stoppedAll
}

func (s *Server) Close() error {
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

func (s *Server) removeOldClips(maxClipAgeInDays int, stop chan struct{}) chan struct{} {
	stopping := make(chan struct{})

	ticker := time.NewTicker(5 * time.Second)

	go func(stop chan struct{}) {
		var currentConnection int
		for {
			time.Sleep(time.Millisecond * 10)
			select {
			case <-ticker.C:
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
			case _, notStopped := <-stop:
				if !notStopped {
					ticker.Stop()
					close(s.stoppedRemovingClips)
					return
				}
			}
		}
	}(stop)

	return stopping
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
