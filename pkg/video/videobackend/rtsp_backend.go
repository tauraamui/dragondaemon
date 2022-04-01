package videobackend

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

type rtspRequest struct{}

type rtspClient struct {
	supportedSchemes []string
	conn             net.Conn
}

func NewRTSPClient() rtspClient {
	return rtspClient{
		supportedSchemes: []string{"rtsp"},
	}
}

func (r *rtspClient) Connect(cancel context.Context, addr string) error {
	return r.TCPConnect(cancel, addr)
}

func processURL(addr string, supportedSchemes []string) (string, error) {
	if len(addr) == 0 {
		return "", errors.New("connection address is undefined")
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}

	scheme := u.Scheme
	if ok := containsString(scheme, supportedSchemes); !ok {
		return "", fmt.Errorf("scheme: %s is unsupported", scheme)
	}

	if !strings.Contains(u.Host, ":") {
		u.Host += ":554"
	}

	return u.Host, nil
}

func (r *rtspClient) TCPConnect(cancel context.Context, addr string) error {
	return r.connect(cancel, "tcp", addr)
}

func (r *rtspClient) UDPConnect(cancel context.Context, addr string) error {
	return r.connect(cancel, "udp", addr)
}

func containsString(str string, strs []string) bool {
	for _, s := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func (r *rtspClient) connect(cancel context.Context, proto, addr string) error {
	url, err := processURL(addr, r.supportedSchemes)
	if err != nil {
		return err
	}

	d := &net.Dialer{}
	conn, err := d.DialContext(cancel, proto, url)
	if err != nil {
		return err
	}

	r.conn = conn

	return nil
}

func (r *rtspClient) options(addr string) {

}

func (r *rtspClient) describe(addr string) {
}

type rtspBackend struct{}

func (b *rtspBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	return &rtspConnection{}, nil
}

type vidConn interface {
	IsOpened() bool
	Close() error
}

type rtspConnection struct {
	uuid   string
	mu     sync.Mutex
	isOpen bool
}

func (c *rtspConnection) connect(cancel context.Context, addr string) error {
	return errors.New("rtsp backend connection not yet implemented")
}

func (c *rtspConnection) UUID() string {
	if len(c.uuid) == 0 {
		c.uuid = uuid.NewString()
	}
	return c.uuid
}

func (c *rtspConnection) Read(frame videoframe.Frame) error {
	// mat, ok := frame.DataRef().(*gocv.Mat)
	// if !ok {
	// 	return xerror.New("must pass OpenCV frame to OpenCV connection read")
	// }
	// c.mu.Lock()
	// defer c.mu.Unlock()
	// ok = readFromVideoConnection(c.vc, mat)
	// if !ok {
	// 	return xerror.New("unable to read from video connection")
	// }
	return errors.New("reading from rtsp vid conn not yet implemented")
}

func (c *rtspConnection) IsOpen() bool { return false }

func (c *rtspConnection) Close() error { return nil }
