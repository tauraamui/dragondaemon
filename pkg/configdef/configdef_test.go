package configdef

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmptyConfigPasses(t *testing.T) {
	body := `{}`
	config := Values{}
	json.Unmarshal([]byte(body), &config)

	assert.NoError(t, config.Validate())
}

func TestHasDupCameraTitlesDoesNotFindDuplicates(t *testing.T) {
	cameras := []Camera{}
	assert.False(t, hasDupCameraTitles(cameras))

	cameras = []Camera{
		{Title: "TestCam1"},
		{Title: "TestCam2"},
		{Title: "TestCam3"},
	}

	assert.False(t, hasDupCameraTitles(cameras))
}

func TestHasDupCameraTitlesDoesFindDuplicates(t *testing.T) {
	cameras := []Camera{
		{Title: "TestCam1"},
		{Title: "TestCam2"},
		{Title: "TestCam3"},
		{Title: "TestCam3"},
		{Title: "TestCam4"},
	}

	assert.True(t, hasDupCameraTitles(cameras))
}

func TestHasDupCameraTitlesDoesFindDuplicateWithLargeGap(t *testing.T) {
	cameras := []Camera{
		{Title: "TestCam1"},
		{Title: "TestCam2"},
		{Title: "TestCam3"},
		{Title: "TestCam4"},
		{Title: "TestCam5"},
		{Title: "TestCam6"},
		{Title: "TestCam7"},
		{Title: "TestCam8"},
		{Title: "TestCam1"},
		{Title: "TestCam10"},
	}

	assert.True(t, hasDupCameraTitles(cameras))
}
