package dragon_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type testConfigResolver struct {
	resolveCallback func()
	resolveConfigs  func() configdef.Values
	resolveError    error
}

func (tcc testConfigResolver) Create() error {
	return errors.New("create not implemented")
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
	return errors.New("destroy not implemented")
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
	s := dragon.NewServer(testConfigResolver{}, video.DefaultBackend())
	assert.NotNil(t, s, "new server's response cannot be nil pointer")
}

func TestServerLoadConfig(t *testing.T) {
	configResolved := false
	cb := func() {
		configResolved = true
	}
	s := dragon.NewServer(testConfigResolver{resolveCallback: cb}, testVideoBackend{})
	err := s.LoadConfiguration()

	require.NoError(t, err, "Resolve returned error when this should be impossible")
	assert.True(t, configResolved)
}

func TestServerLoadConfigWithDisabledsLogs(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	var warnLogs []string
	resetLogWarn := overloadWarnLog(
		func(format string, a ...interface{}) {
			warnLogs = append(warnLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogWarn()

	s := dragon.NewServer(
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
	require.NoError(t, s.LoadConfiguration())
	require.Len(t, s.Connect(), 0)

	assert.Len(t, warnLogs, 1)
	assert.Contains(t, warnLogs, "Camera [Disabled camera] is disabled... skipping...")
}

func TestServerLoadConfigGivesErrorOnResolveError(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s := dragon.NewServer(testConfigResolver{
		resolveError: errors.New("test unable to resolve config"),
	}, testVideoBackend{})
	err := s.LoadConfiguration()
	require.EqualError(t, err, "test unable to resolve config")
}

func TestServerConnect(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s := dragon.NewServer(testConfigResolver{}, testVideoBackend{})
	err := s.LoadConfiguration()
	require.NoError(t, err)
	errs := s.Connect()
	assert.Len(t, errs, 0)
}

type testWaitsOnCancelVideoBackend struct {
}

func (b testWaitsOnCancelVideoBackend) Connect(ctx context.Context, addr string) (video.Connection, error) {
	<-ctx.Done()
	return testVideoConnection{}, errors.New("test unable to connect, context cancelled")
}

func (b testWaitsOnCancelVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

func (b testWaitsOnCancelVideoBackend) NewWriter() video.ClipWriter {
	return nil
}

// TODO(tauraamui): these can potentially block the test run forever, add timeout
func TestServerConnectWithImmediateCancelInvoke(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s := dragon.NewServer(testConfigResolver{}, testWaitsOnCancelVideoBackend{})
	require.NoError(t, s.LoadConfiguration())

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []error)
	go func(ctx context.Context) {
		errs <- s.ConnectWithCancel(ctx)
	}(ctx)
	cancel()

	connErrs := <-errs
	assert.Len(t, connErrs, 0)
}

// TODO(tauraamui): these can potentially block the test run forever, add timeout
func TestServerConnectWithDelayedCancelInvoke(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	s := dragon.NewServer(testConfigResolver{}, testWaitsOnCancelVideoBackend{})
	require.NoError(t, s.LoadConfiguration())

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []error)
	go func(ctx context.Context) {
		errs <- s.ConnectWithCancel(ctx)
	}(ctx)
	time.Sleep(1 * time.Millisecond)
	cancel()

	connErrs := <-errs
	require.Len(t, connErrs, 1)
	assert.EqualError(
		t, connErrs[0], "Unable to connect to camera [Test camera]: test unable to connect, context cancelled",
	)
}

func TestServerShutdown(t *testing.T) {
	logging.CurrentLoggingLevel = logging.SilentLevel
	defer func() { logging.CurrentLoggingLevel = logging.WarnLevel }()
	var warnLogs []string
	resetLogWarn := overloadWarnLog(
		func(format string, a ...interface{}) {
			warnLogs = append(warnLogs, fmt.Sprintf(format, a...))
		},
	)
	defer resetLogWarn()

	s := dragon.NewServer(testConfigResolver{}, testVideoBackend{})
	require.NoError(t, s.LoadConfiguration())
	require.Len(t, s.Connect(), 0)

	timeout := time.After(3 * time.Second)
	done := make(chan interface{})
	go func(t *testing.T, s dragon.Server, done chan interface{}) {
		defer close(done)
		<-s.Shutdown()
	}(t, s, done)

	select {
	case <-timeout:
		t.Fatal("Timeout exceeded. Shutdown is blocking for too long...")
	case <-done:
	}

	assert.Len(t, warnLogs, 1)
	assert.Contains(t, warnLogs, "Closing camera connection: [Test camera]...")
}
