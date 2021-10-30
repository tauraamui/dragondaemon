package dragon_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/dragondaemon/pkg/xis"
	"github.com/tauraamui/xerror"
)

type testConfigResolver struct {
	resolveCallback func()
	resolveConfigs  func() configdef.Values
	resolveError    error
}

func (tcc testConfigResolver) Create() error {
	return xerror.New("create not implemented")
}

func (tcc testConfigResolver) Resolve() (configdef.Values, error) {
	if tcc.resolveCallback != nil {
		tcc.resolveCallback()
	}

	if tcc.resolveError != nil {
		return configdef.Values{}, tcc.resolveError
	}

	if tcc.resolveConfigs != nil {
		return tcc.resolveConfigs(), nil
	}

	return configdef.Values{
		Cameras: []configdef.Camera{{Title: "Test camera"}},
	}, nil
}

func (tcc testConfigResolver) Destroy() error {
	return xerror.New("destroy not implemented")
}

type testVideoBackend struct {
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	return testVideoConnection{}, nil
}

func (tvb testVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

func (tvb testVideoBackend) NewWriter() video.ClipWriter {
	return nil
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Dimensions() video.FrameDimension {
	return video.FrameDimension{W: 100, H: 50}
}

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
}

func (tvc testVideoConnection) UUID() string {
	return "test-conn-uuid"
}

func (tvc testVideoConnection) Read(frame video.Frame) error {
	return nil
}

func (tvc testVideoConnection) IsOpen() bool {
	return true
}

func (tvc testVideoConnection) Close() error {
	return nil
}

func TestNewServer(t *testing.T) {
	is := is.New(t)
	s, err := dragon.NewServer(testConfigResolver{}, video.DefaultBackend())
	is.NoErr(err)
	is.True(s != nil) // new server's response cannot be nil pointer
}

func TestServerLoadConfig(t *testing.T) {
	is := is.New(t)
	configResolved := false
	cb := func() {
		configResolved = true
	}
	s, err := dragon.NewServer(testConfigResolver{resolveCallback: cb}, testVideoBackend{})

	is.True(s != nil)
	is.NoErr(err) // resolve returned error when this should be impossible
	is.True(configResolved)
}

func TestServerLoadConfigWithDisabledsLogs(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	var warnLogs []string
	resetLogWarn := overloadWarnLog(
		func(format string, a ...interface{}) {
			warnLogs = append(warnLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogWarn()

	s, err := dragon.NewServer(
		testConfigResolver{
			resolveConfigs: func() configdef.Values {
				return configdef.Values{
					Cameras: []configdef.Camera{
						{Title: "Disabled camera", Disabled: true},
					},
				}
			},
		},
		testVideoBackend{},
	)

	is.True(s != nil)
	is.NoErr(err)
	is.Equal(len(s.Connect()), 0)

	is.Equal(len(warnLogs), 1)
	is.True(xis.Contains(warnLogs, "Camera [Disabled camera] is disabled... skipping..."))
}

func TestServerLoadConfigGivesErrorOnResolveError(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s, err := dragon.NewServer(testConfigResolver{
		resolveError: xerror.New("test low level resolve error"),
	}, testVideoBackend{})
	is.True(s == nil)
	is.Equal(err.Error(), "unable to resolve config: test low level resolve error")
}

func TestServerConnect(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s, err := dragon.NewServer(testConfigResolver{}, testVideoBackend{})
	is.True(s != nil)
	is.NoErr(err)
	errs := s.Connect()
	is.Equal(len(errs), 0)
}

type testWaitsOnCancelVideoBackend struct {
}

func (b testWaitsOnCancelVideoBackend) Connect(ctx context.Context, addr string) (video.Connection, error) {
	<-ctx.Done()
	return testVideoConnection{}, xerror.New("test unable to connect, context cancelled")
}

func (b testWaitsOnCancelVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

func (b testWaitsOnCancelVideoBackend) NewWriter() video.ClipWriter {
	return nil
}

// TODO(tauraamui): these can potentially block the test run forever, add timeout
func TestServerConnectWithImmediateCancelInvoke(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s, err := dragon.NewServer(testConfigResolver{}, testWaitsOnCancelVideoBackend{})

	is.True(s != nil)
	is.NoErr(err)

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []error)
	go func(ctx context.Context) {
		errs <- s.ConnectWithCancel(ctx)
	}(ctx)
	cancel()

	connErrs := <-errs
	is.Equal(len(connErrs), 0)
}

// TODO(tauraamui): these can potentially block the test run forever, add timeout
func TestServerConnectWithDelayedCancelInvoke(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s, err := dragon.NewServer(testConfigResolver{}, testWaitsOnCancelVideoBackend{})
	is.True(s != nil)
	is.NoErr(err)

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []error)
	go func(ctx context.Context) {
		errs <- s.ConnectWithCancel(ctx)
	}(ctx)
	time.Sleep(1 * time.Millisecond)
	cancel()

	connErrs := <-errs
	is.Equal(len(connErrs), 1)
	is.Equal(connErrs[0].Error(), "Unable to connect to camera [Test camera]: test unable to connect, context cancelled")
}

func TestServerShutdown(t *testing.T) {
	is := is.New(t)
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	var warnLogs []string
	resetLogWarn := overloadWarnLog(
		func(format string, a ...interface{}) {
			warnLogs = append(warnLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogWarn()

	s, err := dragon.NewServer(testConfigResolver{}, testVideoBackend{})
	is.True(s != nil)
	is.NoErr(err)
	is.Equal(len(s.Connect()), 0)

	timeout := time.After(3 * time.Second)
	done := make(chan interface{})
	go func(t *testing.T, s *dragon.Server, done chan interface{}) {
		defer close(done)
		<-s.Shutdown()
	}(t, s, done)

	select {
	case <-timeout:
		t.Fatal("Timeout exceeded. Shutdown is blocking for too long...")
	case <-done:
	}

	is.Equal(len(warnLogs), 1)
	is.Equal(warnLogs[0], "Closing camera connection: [Test camera]...")
}
