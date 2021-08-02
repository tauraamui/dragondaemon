package dragon_test

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
)

type testConfigResolver struct{}

func (tcc testConfigResolver) Resolve() (configdef.Values, error) {
	return configdef.Values{}, nil
}

func TestNewServer(t *testing.T) {
	s := dragon.NewServer(testConfigResolver{})
	if s == nil {
		t.Error("New server's response cannot be nil pointer")
	}
}
