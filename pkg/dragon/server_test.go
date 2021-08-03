package dragon_test

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
)

type testConfigResolver struct {
	resolveCallback func()
}

func (tcc testConfigResolver) Resolve() (configdef.Values, error) {
	tcc.resolveCallback()
	return configdef.Values{
		Cameras: []configdef.Camera{{Title: "Test camera"}},
	}, nil
}

func TestNewServer(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{})
	if s == nil {
		t.Error("New server's response cannot be nil pointer")
	}
}

func TestServerLoadConfig(t *testing.T) {
	configResolved := false
	cb := func() {
		configResolved = true
	}
	s := dragon.NewServer(testConfigResolver{resolveCallback: cb})
	err := s.LoadConfiguration()

	if err != nil {
		t.Error("Server load config returned error despite that being impossible")
	}

	if !configResolved {
		t.Error("Server load config does not call resolve against config resolver")
	}
}
