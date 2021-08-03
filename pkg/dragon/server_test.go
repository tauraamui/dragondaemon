package dragon_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
)

type testConfigResolver struct {
	resolveCallback func()
	resolveError    error
}

func (tcc testConfigResolver) Resolve() (configdef.Values, error) {
	if tcc.resolveCallback != nil {
		tcc.resolveCallback()
	}

	if tcc.resolveError != nil {
		return configdef.Values{}, tcc.resolveError
	}

	return configdef.Values{
		Cameras: []configdef.Camera{{Title: "Test camera"}},
	}, nil
}

func TestNewServer(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{})
	assert.NotNil(t, s, "new server's response cannot be nil pointer")
}

func TestServerLoadConfig(t *testing.T) {
	configResolved := false
	cb := func() {
		configResolved = true
	}
	s := dragon.NewServer(testConfigResolver{resolveCallback: cb})
	err := s.LoadConfiguration()

	assert.Nil(t, err, "Resolve returned error when this should be impossible")
	assert.True(t, configResolved)
}

func TestServerLoadConfigGivesErrorOnResolveError(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{
		resolveError: errors.New("test unable to resolve config"),
	})
	err := s.LoadConfiguration()
	assert.NotNil(t, err)
	assert.EqualError(t, err, "test unable to resolve config")
}
