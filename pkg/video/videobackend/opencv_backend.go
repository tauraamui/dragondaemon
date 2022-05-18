package videobackend

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
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
	prefix := make([]byte, 12)
	prefix[0] = 0x31
	prefix[1] = 0x13

	frameMatData := frame.mat.ToBytes()
	storeUint32(prefix, 2, len(frameMatData)) // store size of frame data
	storeUint16(prefix, 6, frame.mat.Rows())
	storeUint16(prefix, 8, frame.mat.Cols())
	storeUint16(prefix, 10, int(frame.mat.Type()))

	return append(prefix, frameMatData...)
}

func storeUint16(data []byte, index int, value int) {
	if datal := len(data); index+2 > datal {
		panic(fmt.Errorf("run out of space to store two bytes: MAX(%d), INDEX(%d), %d bytes left over", datal, index, datal-index))
	}

	binary.LittleEndian.PutUint16(data[index:index+2], uint16(value))
}

func storeUint32(data []byte, index int, value int) {
	if datal := len(data); index+4 > datal {
		panic(fmt.Errorf("run out of space to store four bytes: MAX(%d), INDEX(%d), %d bytes left over", datal, index, datal-index))
	}

	binary.LittleEndian.PutUint32(data[index:index+4], uint32(value))
}

func loadUint16(data []byte, index int) uint16 {
	if datal := len(data); index+2 > datal {
		panic(fmt.Errorf("run out of space to load two bytes: MAX(%d), INDEX(%d), %d bytes left over", datal, index, datal-index))
	}

	return binary.LittleEndian.Uint16(data[index : index+2])
}

func loadUint32(data []byte, index int) uint32 {
	if datal := len(data); index+4 > datal {
		panic(fmt.Errorf("run out of space to load four bytes: MAX(%d), INDEX(%d), %d bytes left over", datal, index, datal-index))
	}

	return binary.LittleEndian.Uint32(data[index : index+4])
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
	if len(d) < 12 { // an encoded OpenCV frame should contain a prefix of 12 bytes so...
		return nil, errors.New("OpenCV frame expects at least 12 bytes to load")
	}

	prefix := d[:12]
	if int(prefix[0]) != 0x31 || int(prefix[1]) != 0x13 {
		return nil, errors.New("OpenCV frame bytes missing leading prefix")
	}

	s := loadUint32(prefix, 2)
	r := loadUint16(prefix, 6)
	c := loadUint16(prefix, 8)
	mattype := gocv.MatType(loadUint16(prefix, 10))
	frameMatData := d[12:s]

	mat, err := gocv.NewMatFromBytes(int(r), int(c), mattype, frameMatData)
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
