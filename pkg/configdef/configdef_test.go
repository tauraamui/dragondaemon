package configdef

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmptyConfigPasses(t *testing.T) {
	// TODO(tauraamui): return this to actually be empty again
	// once this root field has been removed
	// body := `{}`
	body := `{"max_clip_age_in_days": 1}`
	config := Values{}
	json.Unmarshal([]byte(body), &config)

	assert.NoError(t, config.runValidate())
}

func TestValidatePopulatedConfigPassesValidation(t *testing.T) {
	// TODO(tauraamui): return this to actually be empty again
	// once this root field has been removed
	// body := `{}`
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"max_clip_age_days": 15,
					"fps": 11,
					"seconds_per_clip": 1
				}
			]
		}`
	config := Values{}
	json.Unmarshal([]byte(body), &config)

	assert.NoError(t, config.runValidate())
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
