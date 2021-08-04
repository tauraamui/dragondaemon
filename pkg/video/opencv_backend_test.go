package video

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocv.io/x/gocv"
)

func TestOpenVideoStreamInvokesOpenVideoCapture(t *testing.T) {
	openVideoCapture = func(addr string) (*gocv.VideoCapture, error) {
		return nil, errors.New("test connect error")
	}
	conn := openCVConnection{}
	err := conn.connect(context.TODO(), "TestAddr")
	assert.EqualError(t, err, "test connect error")
}
