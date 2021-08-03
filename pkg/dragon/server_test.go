package dragon_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type testConfigResolver struct {
	resolveCallback func()
	resolveConfigs  func() configdef.Values
	resolveError    error
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

type testVideoBackend struct {
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	return testVideoConnection{}, nil
}

func (tvb testVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
}

func (tvc testVideoConnection) Read(frame video.Frame) error {
	return nil
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

func TestServerLoadConfigGivesErrorOnResolveError(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{
		resolveError: errors.New("test unable to resolve config"),
	}, testVideoBackend{})
	err := s.LoadConfiguration()
	require.EqualError(t, err, "test unable to resolve config")
}

func TestServerConnect(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{}, testVideoBackend{})
	err := s.LoadConfiguration()
	require.NoError(t, err)
	errs := s.Connect()
	assert.Len(t, errs, 0)

	timeout := time.After(3 * time.Second)
	done := make(chan interface{})
	go func(t *testing.T, s dragon.Server) {
		<-s.Shutdown()
		close(done)
	}(t, s)

	select {
	case <-timeout:
		t.Fatal("Timeout exceeded. Shutdown is blocking for too long...")
	case <-done:
	}
}
