package xerror_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/internal/xerror"
)

func TestNewErrorGivesErrInstance(t *testing.T) {
	is := is.New(t)

	err := xerror.New().Msg("order was declined").WithParams(
		map[string]string{"cheese": "smelt awful"},
	)
	is.True(err != nil)
}
