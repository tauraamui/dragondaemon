package process_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
)

func TestNewCoreProcess(t *testing.T) {
	is := is.New(t)
	conn := mockCameraConn{}
	writer := mockClipWriter{}
	proc := process.NewCoreProcess(&conn, &writer)

	is.True(proc != nil)
}
