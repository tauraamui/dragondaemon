package video_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func TestVideoBackendDefaultBackend(t *testing.T) {
	assert.NotNil(t, video.DefaultBackend())
}
