package media

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Options struct {
	MaxClipAgeInDays int
}

// Server manages receiving RTSP streams and persisting clips to disk
type Server struct {
	debugMode   bool
	inShutdown  int32
	mu          sync.Mutex
	ctx         context.Context
	ctxCancel   context.CancelFunc
	stoppedAll  chan struct{}
	connections map[*Connection]struct{}
}

// NewServer allocates a new server struct.
func NewServer(debugMode bool) *Server {
	return &Server{debugMode: debugMode}
}

// IsRunning reports whether the server is running or not.
func (s *Server) IsRunning() bool {
	return !s.shuttingDown()
}

// Connect allocates a new connection struct which server tracks internally.
// A new video stream connection is opened to the target stream location
// which is saved within the connection struct allocation.
func (s *Server) Connect(
	title string,
	rtspStream string,
	sett ConnectonSettings,
) {
	vc, err := openVideoCapture(
		rtspStream, title, sett.FPS, sett.DateTimeLabel, sett.DateTimeFormat, sett.MockCapturer,
	)

	if err != nil {
		log.Error("Unable to connect to stream [%s] at [%s]", title, rtspStream) //nolint
		return
	}

	log.Info("Connected to stream [%s] at [%s]", title, rtspStream) //nolint
	conn := NewConnection(
		title,
		sett,
		vc,
		rtspStream,
	)

	go outputConnectionSizeOnDisk(conn)

	s.trackConnection(conn, true)
}

// Run beings the server process using the provided options.
// The server will save connection streams to disk, manages
// the individual connection streams and removes old clips.
func (s *Server) Run(opts Options) {
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.stoppedAll = make(chan struct{})

	go func(ctx context.Context, stopped chan struct{}) {
		streamingCtx, cancelStreaming := context.WithCancel(context.Background())
		savingClipsCtx, cancelSavingClips := context.WithCancel(context.Background())
		removingClipsCtx, cancelRemovingClips := context.WithCancel(context.Background())

		// streaming connections is core and is the dependancy to all subsequent processes
		stoppedStreaming := s.beginStreaming(streamingCtx)
		stoppedSavingClips := s.saveStreams(savingClipsCtx)
		stoppedRemovingClips := s.removeOldClips(removingClipsCtx, opts.MaxClipAgeInDays)

		// wait for shutdown signal
		<-ctx.Done()

		cancelSavingClips()
		log.Info("Waiting for persist process to finish...") //nolint
		// wait for saving streams to stop
		for _, stoppedPersistSig := range stoppedSavingClips {
			<-stoppedPersistSig
		}

		// stopping the streaming process should be done last
		// stop all streaming
		cancelStreaming()
		log.Info("Waiting for streams to terminate...") //nolint
		// wait for all streams to stop
		// TODO(:tauraamui) Move each stream stop signal wait onto separate goroutine
		for _, stoppedStreamSig := range stoppedStreaming {
			<-stoppedStreamSig
		}

		cancelRemovingClips()
		log.Info("Waiting for removing clips process to finish...") //nolint
		<-stoppedRemovingClips

		// send signal saying shutdown process has finished
		close(stopped)
	}(s.ctx, s.stoppedAll)
}

// Shutdown kills the server process and begins terminating all of it's
// child processes.
func (s *Server) Shutdown() chan struct{} {
	atomic.StoreInt32(&s.inShutdown, 1)
	s.ctxCancel()
	return s.stoppedAll
}

// Close closes all open/active video stream connections.
func (s *Server) Close() error {
	return s.closeConnectionsLocked()
}

func (s *Server) beginStreaming(ctx context.Context) []chan struct{} {
	var stoppedStreaming []chan struct{}
	for _, conn := range s.activeConnections() {
		log.Info("Reading stream from connection [%s]", conn.title) //nolint
		stoppedStreaming = append(stoppedStreaming, conn.stream(ctx))
	}
	return stoppedStreaming
}

func (s *Server) saveStreams(ctx context.Context) []chan interface{} {
	var stoppedPersisting []chan interface{}
	for _, conn := range s.activeConnections() {
		stoppedPersisting = append(stoppedPersisting, conn.writeStreamToClips(ctx))
	}
	return stoppedPersisting
}

func (s *Server) removeOldClips(ctx context.Context, maxClipAgeInDays int) chan struct{} {
	stopping := make(chan struct{})

	ticker := time.NewTicker(5 * time.Second)

	go func(ctx context.Context, stopping chan struct{}) {
		var currentConnection int
		for {
			time.Sleep(time.Millisecond * 10)
			select {
			case <-ticker.C:
				activeConnections := s.activeConnections()
				if len(activeConnections) == 0 {
					continue
				}
				if currentConnection >= len(activeConnections) {
					currentConnection = 0
				}

				if conn := activeConnections[currentConnection]; conn != nil {
					fullPersistLocation := fmt.Sprintf("%s%c%s", conn.sett.PersistLocation, os.PathSeparator, conn.title)
					files, err := ioutil.ReadDir(fullPersistLocation)
					if err != nil {
						log.Error("Unable to read contents of connection persist location %s: %v", fullPersistLocation, err) //nolint
					}

					for _, file := range files {
						date, err := time.Parse("2006-01-02", file.Name())
						if err != nil {
							continue
						}

						oldestAllowedDay := time.Now().AddDate(0, 0, -1*maxClipAgeInDays)
						if date.Before(oldestAllowedDay) {
							dirToRemove := fmt.Sprintf("%s%c%s", fullPersistLocation, os.PathSeparator, file.Name())
							log.Info("REMOVING DIR %s", dirToRemove) //nolint
							err := os.RemoveAll(dirToRemove)
							if err != nil {
								log.Error("Failed to RemoveAll %s", dirToRemove) //nolint
							}
						}
					}
				}

				currentConnection++
			case <-ctx.Done():
				ticker.Stop()
				close(stopping)
				return
			}
		}
	}(ctx, stopping)

	return stopping
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

func outputConnectionSizeOnDisk(c *Connection) {
	log.Info("Fetching connection size on disk...") //nolint
	size, err := c.SizeOnDisk()
	if err != nil {
		log.Error("Unable to fetch size on disk: %v", err) //nolint
		return
	}
	log.Info("Connection [%s] size on disk: %s", c.title, size) //nolint
}
