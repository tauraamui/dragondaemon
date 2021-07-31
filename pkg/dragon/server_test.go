package dragon_test

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/dragon"
)

func TestNewServer(t *testing.T) {
	s := dragon.NewServer(dragon.DefaultConfigResolver())
	if s == nil {
		t.Error("New server's response cannot be nil pointer")
	}
}
