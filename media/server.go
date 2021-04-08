package media

import (
	"context"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gocv.io/x/gocv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

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

type Options struct {
	MaxClipAgeInDays int
}

type videoCapture struct {
	p                 *gocv.VideoCapture
	drawDateTimeLabel bool
}

func (vc *videoCapture) SetP(c *gocv.VideoCapture) {
	vc.p = c
}

func (vc *videoCapture) IsOpened() bool {
	return vc.p.IsOpened()
}

func (vc *videoCapture) Read(m *gocv.Mat) bool {
	read := vc.p.Read(m)
	if read && vc.drawDateTimeLabel {
		gocv.PutText(
			m,
			time.Now().Format("2006/01/02 15:04:05.999999999"),
			image.Pt(15, 50),
			gocv.FontHersheyPlain,
			3,
			color.RGBA{255, 255, 255, 255},
			int(gocv.Line4),
		)
	}
	return read
}

func (vc *videoCapture) Close() error {
	return vc.p.Close()
}

type mockVideoCapture struct {
	title       string
	stream      gocv.Mat
	initialised bool
	baseImage   image.Image
}

func (mvc *mockVideoCapture) SetP(_ *gocv.VideoCapture) {}

func (mvc *mockVideoCapture) IsOpened() bool {
	return true
}

func (mvc *mockVideoCapture) Read(m *gocv.Mat) bool {
	if !mvc.initialised {
		var w, h int = 1400, 1200
		var hw, hh float64 = float64(w / 2), float64(h / 2)
		r := 200.0
		θ := 2 * math.Pi / 3
		cr := &circle{hw - r*math.Sin(0), hh - r*math.Cos(0), 300}
		cg := &circle{hw - r*math.Sin(θ), hh - r*math.Cos(θ), 300}
		cb := &circle{hw - r*math.Sin(-θ), hh - r*math.Cos(-θ), 300}

		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				c := color.RGBA{
					cr.Brightness(float64(x), float64(y)),
					cg.Brightness(float64(x), float64(y)),
					cb.Brightness(float64(x), float64(y)),
					255,
				}
				img.Set(x, y, c)
			}
		}
		mvc.baseImage = img
		mvc.initialised = true
	}

	baseClone := cloneImage(mvc.baseImage)
	drawText(baseClone, 5, 50, "DD_OFFLINE_STREAM")
	drawText(baseClone, 5, 180, mvc.title)
	drawText(baseClone, 5, 310, time.Now().Format("2006-01-02 15:04:05.999999999"))

	mat, err := gocv.ImageToMatRGB(baseClone)
	if err != nil {
		logging.Fatal("Unable to convert Go image into OpenCV mat")
	}

	time.Sleep(time.Millisecond * 100)
	mat.CopyTo(m)
	mvc.stream = mat
	return mvc.initialised
}

func cloneImage(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

func drawText(canvas *image.RGBA, x, y int, text string) error {
	var (
		fgColor  image.Image
		fontFace *truetype.Font
		err      error
		fontSize = 64.0
	)
	fgColor = image.White
	fontFace, err = freetype.ParseFont(goregular.TTF)
	fontDrawer := &font.Drawer{
		Dst: canvas,
		Src: fgColor,
		Face: truetype.NewFace(fontFace, &truetype.Options{
			Size:    fontSize,
			Hinting: font.HintingFull,
		}),
	}
	textBounds, _ := fontDrawer.BoundString(text)
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	yPosition := fixed.I((y)-textHeight.Ceil())/2 + fixed.I(textHeight.Ceil())
	fontDrawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: yPosition,
	}
	fontDrawer.DrawString(text)
	return err
}

func (mvc *mockVideoCapture) Close() error {
	mvc.initialised = false
	mvc.stream.Close()
	return nil
}

// NewServer returns a pointer to media server instance
func NewServer(debugMode bool) *Server {
	return &Server{debugMode: debugMode}
}

func (s *Server) IsRunning() bool {
	return !s.shuttingDown()
}

func (s *Server) Connect(
	title string,
	rtspStream string,
	persistLocation string,
	fps int,
	dateTimeLabel bool,
	secondsPerClip int,
	schedule schedule.Schedule,
	reolink config.ReolinkAdvanced,
) {
	vc, err := openVideoCapture(rtspStream, title, fps, dateTimeLabel)
	if err != nil {
		logging.Error("Unable to connect to stream [%s] at [%s]", title, rtspStream)
		return
	}

	logging.Info("Connected to stream [%s] at [%s]", title, rtspStream)
	if len(persistLocation) == 0 {
		persistLocation = "."
	}
	conn := NewConnection(
		title,
		persistLocation,
		fps,
		secondsPerClip,
		schedule,
		reolink,
		vc,
		rtspStream,
	)

	go func(c *Connection) {
		logging.Info("Fetching connection size on disk...")
		size, unit, err := c.SizeOnDisk()
		if err != nil {
			logging.Error("Unable to fetch size on disk: %v", err)
			return
		}
		logging.Info("Connection [%s] size on disk: %d%s", conn.title, size, unit)
	}(conn)

	s.trackConnection(conn, true)
}

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
		logging.Info("Waiting for persist process to finish...")
		// wait for saving streams to stop
		<-stoppedSavingClips

		// stopping the streaming process should be done last
		// stop all streaming
		cancelStreaming()
		logging.Info("Waiting for streams to terminate...")
		// wait for all streams to stop
		// TODO(:tauraamui) Move each stream stop signal wait onto separate goroutine
		for _, stoppedStreamSig := range stoppedStreaming {
			<-stoppedStreamSig
		}

		cancelRemovingClips()
		logging.Info("Waiting for removing clips process to finish...")
		<-stoppedRemovingClips

		// send signal saying shutdown process has finished
		close(stopped)
	}(s.ctx, s.stoppedAll)
}

func (s *Server) Shutdown() chan struct{} {
	atomic.StoreInt32(&s.inShutdown, 1)
	s.ctxCancel()
	return s.stoppedAll
}

func (s *Server) Close() error {
	return s.closeConnectionsLocked()
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

	go func(ctx context.Context, stopping chan struct{}) {
		for {
			time.Sleep(time.Millisecond * 1)
			select {
			case <-ctx.Done():
				close(stopping)
				return
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
				if currentConnection >= len(activeConnections) {
					currentConnection = 0
				}

				if conn := activeConnections[currentConnection]; conn != nil {
					fullPersistLocation := fmt.Sprintf("%s%c%s", conn.persistLocation, os.PathSeparator, conn.title)
					files, err := ioutil.ReadDir(fullPersistLocation)
					if err != nil {
						logging.Error("Unable to read contents of connection persist location %s: %v", fullPersistLocation, err)
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

func openVideoCapture(rtspStream string, title string, fps int, dateTimeLabel bool) (VideoCapturable, error) {
	mockVidStream, foundEnv := os.LookupEnv("DRAGON_DAEMON_MOCK_VIDEO_STREAM")
	if foundEnv && mockVidStream == "1" {
		return &mockVideoCapture{title: title}, nil
	}

	vc, err := gocv.OpenVideoCapture(rtspStream)
	if err != nil {
		return nil, err
	}

	vc.Set(gocv.VideoCaptureFPS, float64(fps))
	return &videoCapture{p: vc, drawDateTimeLabel: dateTimeLabel}, err
}

type circle struct {
	X, Y, R float64
}

func (c *circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	} else {
		return 255
	}
}
