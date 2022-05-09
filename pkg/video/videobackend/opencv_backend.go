package videobackend

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
	"gocv.io/x/gocv"
)

type openCVFrame struct {
	isClosed  bool
	mat       gocv.Mat
	timestamp int64
}

func (frame *openCVFrame) Timestamp() int64 { return frame.timestamp }

func (frame *openCVFrame) DataRef() interface{} {
	return &frame.mat
}

func (frame *openCVFrame) ToBytes() []byte {
	var r, c, mt uint16
	// store the OpenCV matrix rows, columns and type
	r = uint16(frame.mat.Rows())  // 2 bytes
	c = uint16(frame.mat.Cols())  // 2 bytes
	mt = uint16(frame.mat.Type()) // 2 byte

	suffix := make([]byte, 8)
	binary.LittleEndian.PutUint16(suffix[:2], r)
	binary.LittleEndian.PutUint16(suffix[2:4], c)
	binary.LittleEndian.PutUint16(suffix[4:6], mt)
	suffix[6] = 0x13
	suffix[7] = 0x31

	return append(frame.mat.ToBytes(), suffix...)
}

func (frame *openCVFrame) Dimensions() videoframe.Dimensions {
	return videoframe.Dimensions{W: frame.mat.Cols(), H: frame.mat.Rows()}
}

func (frame *openCVFrame) Close() {
	if !frame.isClosed {
		frame.mat.Close()
		frame.isClosed = true
	}
}

type openCVBackend struct{}

func (b *openCVBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	conn := openCVConnection{}
	err := conn.connect(cancel, addr)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (b *openCVBackend) NewFrame() videoframe.Frame {
	return &openCVFrame{mat: gocv.NewMat()}
}

func (b *openCVBackend) NewFrameFromBytes(d []byte) (videoframe.Frame, error) {
	if len(d) < 8 {
		return nil, errors.New("OpenCV frame expects at least 8 bytes to load")
	}

	dl := len(d)
	suffix := d[dl-8:]

	if int(suffix[6]) != 0x13 || int(suffix[7]) != 0x31 {
		return nil, errors.New("OpenCV frame bytes missing trailing suffix")
	}

	r := binary.LittleEndian.Uint16(suffix[:2])
	c := binary.LittleEndian.Uint16(suffix[2:4])
	mtypeid := binary.LittleEndian.Uint16(suffix[4:6])
	mattype := gocv.MatType(mtypeid)

	if dl-8 < 8 {
		d = []byte{}
	} else {
		d = d[:dl-8]
	}

	mat, err := gocv.NewMatFromBytes(int(r), int(c), mattype, d)
	if err != nil {
		return nil, err
	}
	return &openCVFrame{mat: mat}, nil
}

func (b *openCVBackend) NewWriter() videoclip.Writer {
	return &openCVClipWriter{
		onWriteInitDone: false,
	}
}

const codec = "avc1.4d001e"

type openCVClipWriter struct {
	onWriteInitDone bool
	vw              *gocv.VideoWriter
	clip            videoclip.NoCloser
}

func (w *openCVClipWriter) init(clip videoclip.NoCloser) error {
	if err := ensureDirectoryPathExists(clip.RootPath()); err != nil {
		return err
	}
	w.clip = clip

	dimensions, err := clip.Dimensions()
	if err != nil {
		return err
	}

	vw, err := openVideoWriter(
		clip.FileName(), codec, float64(clip.FPS()), dimensions.W, dimensions.H, true,
	)
	if err != nil {
		return err
	}
	w.vw = vw
	return nil
}

var openVideoWriter = func(filename, codec string, fps float64, width, height int, isColor bool) (*gocv.VideoWriter, error) {
	return gocv.VideoWriterFile(filename, codec, fps, width, height, isColor)
}

func ensureDirectoryPathExists(path string) error {
	err := fs.MkdirAll(path, os.ModePerm|os.ModeDir)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func (w *openCVClipWriter) reset() {
	w.vw.Close()
	w.vw = nil
}

func (w *openCVClipWriter) Write(clip videoclip.NoCloser) error {
	// TODO(tauraamui):
	// make clip frames fetch be statically referenced from here
	// instead of referring to internal instance again
	if len(clip.Frames()) == 0 {
		return xerror.New("cannot write empty clip")
	}
	if err := w.init(clip); err != nil {
		return err
	}
	defer w.reset()
	for _, frame := range clip.Frames() {
		if err := w.writeFrame(frame); err != nil {
			return err
		}
	}
	return nil
}

func (w *openCVClipWriter) writeFrame(frame videoframe.NoCloser) error {
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return xerror.New("must pass OpenCV frame to OpenCV writer")
	}
	return w.vw.Write(*mat)
}

type openCVConnection struct {
	uuid   string
	mu     sync.Mutex
	isOpen bool
	vc     *gocv.VideoCapture
}

func (c *openCVConnection) connect(cancel context.Context, addr string) error {
	connAndError := make(chan openVideoStreamResult)
	go openVideoStream(addr, connAndError)
	select {
	case r := <-connAndError:
		if r.err != nil {
			return r.err
		}
		c.vc = r.vc
		c.isOpen = true
		return nil
	case <-cancel.Done():
		return xerror.New("connection cancelled")
	}
}

type openVideoStreamResult struct {
	vc  *gocv.VideoCapture
	err error
}

func openVideoStream(addr string, d chan openVideoStreamResult) {
	vc, err := openVideoCapture(addr)
	result := openVideoStreamResult{vc: vc, err: err}
	d <- result
}

var openVideoCapture = func(addr string) (*gocv.VideoCapture, error) {
	return gocv.OpenVideoCapture(addr)
}

var readFromVideoConnection = func(vc *gocv.VideoCapture, mat *gocv.Mat) bool {
	if vc.IsOpened() {
		return vc.Read(mat)
	}
	return false
}

func (c *openCVConnection) UUID() string {
	if len(c.uuid) == 0 {
		c.uuid = uuid.NewString()
	}
	return c.uuid
}

func (c *openCVConnection) Read(frame videoframe.Frame) error {
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return xerror.New("must pass OpenCV frame to OpenCV connection read")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ok = readFromVideoConnection(c.vc, mat)
	if !ok {
		return xerror.New("unable to read from video connection")
	}
	return nil
}

func (c *openCVConnection) IsOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isOpen {
		return c.vc.IsOpened()
	}
	return false
}

func (c *openCVConnection) Close() error {
	c.mu.Lock()
	c.isOpen = false
	c.mu.Unlock()
	return c.vc.Close()
}
