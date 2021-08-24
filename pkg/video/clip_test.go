package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClip(t *testing.T) {
	clip := NewClip("/testroot/clips/TestConn/2010-02-2/2010-02-02 19:45:00")
	assert.NotNil(t, clip)
}
