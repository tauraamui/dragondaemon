package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/api/auth"
	"github.com/tauraamui/dragondaemon/common"
	db "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"github.com/tauraamui/dragondaemon/pkg/media"
	"gorm.io/gorm"
)

func init() {
	rpc.Register(Session{})
}

const SIGREMOTE = Signal(0x1)

type Signal int

func (s Signal) Signal() {}

func (s Signal) String() string {
	return "remote-shutdown"
}

// TODO(:tauraamui) declare this ahead of time to indicate intention on how to pass these
type Options struct {
	RPCListenPort string
	SigningSecret string
}

type Session struct {
	Token      string
	CameraUUID string
}

func (s Session) GetToken(args string, resp *string) error {
	*resp = s.Token
	return nil
}

type MediaServer struct {
	interrupt     chan os.Signal
	s             *media.Server
	httpServer    *http.Server
	rpcListenPort string
	signingSecret string
	db            *gorm.DB
}

func New(interrupt chan os.Signal, server *media.Server, opts Options) (*MediaServer, error) {
	db, err := db.Connect()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to DB, try running the setup: %w", err)
	}
	return &MediaServer{
		interrupt:     interrupt,
		s:             server,
		httpServer:    &http.Server{},
		rpcListenPort: opts.RPCListenPort,
		signingSecret: opts.SigningSecret,
		db:            db,
	}, nil
}

func StartRPC(m *MediaServer) error {
	err := rpc.Register(m)
	if err != nil {
		return err
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", m.rpcListenPort)
	if err != nil {
		return err
	}

	errs := make(chan error)
	go func() {
		httpErr := m.httpServer.Serve(l)
		if httpErr != nil {
			errs <- httpErr
		}
		errs <- nil
	}()

	select {
	case err := <-errs:
		return err
	default:
		return nil
	}
}

func ShutdownRPC(m *MediaServer) error {
	if m != nil && m.httpServer != nil {
		return m.httpServer.Close()
	}
	return errors.New("API server not running")
}

func (m *MediaServer) Authenticate(authContents []string, resp *string) error {
	if err := validateAuth(authContents); err != nil {
		return err
	}

	username := authContents[0]
	password := authContents[1]

	userRepo := repos.UserRepository{DB: m.db}
	user, err := userRepo.FindByName(username)
	if err != nil {
		return err
	}

	if err := user.ComparePassword(password); err != nil {
		return err
	}

	token, err := auth.GenToken(m.signingSecret, user.UUID)
	if err != nil {
		return err
	}

	*resp = token
	return nil
}

// Exposed API
func (m *MediaServer) ActiveConnections(sess *Session, resp *[]common.ConnectionData) error {
	_, err := validateSession(m.signingSecret, *sess)
	if err != nil {
		return err
	}
	*resp = m.s.APIFetchActiveConnections()
	return nil
}

func (m *MediaServer) RebootConnection(sess *Session, resp *bool) error {
	_, err := validateSession(m.signingSecret, *sess)
	if err != nil {
		return err
	}

	logging.Warn("Received remote reboot connection request...")
	err = m.s.APIRebootConnection(sess.CameraUUID)
	if err != nil {
		*resp = false
		return err
	}

	*resp = true
	return nil
}

func (m *MediaServer) Shutdown(sess *Session, resp *bool) error {
	_, err := validateSession(m.signingSecret, *sess)
	if err != nil {
		return err
	}

	*resp = true
	logging.Warn("Received remote shutdown request...")
	defer func() {
		time.Sleep(time.Second * 1)
		m.interrupt <- SIGREMOTE
	}()
	return nil
}

func validateSession(signingSecret string, sess Session) (string, error) {
	return auth.ValidateToken(signingSecret, sess.Token)
}

func validateAuth(auth []string) error {
	if len(auth) == 0 {
		return errors.New("cannot retrieve username and password from blank input")
	}

	if len(auth[0]) == 0 || len(auth[1]) == 0 {
		return errors.New("unable to correctly retrieve username and password from malformed input")
	}

	return nil
}
