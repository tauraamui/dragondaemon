package configdef_test

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
)

func TestValidateEmptyConfigPasses(t *testing.T) {
	is := is.New(t)
	// TODO(tauraamui): return this to actually be empty again
	// once this root field has been removed
	body := `{}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.NoErr(config.RunValidate())
}

func TestValidatePopulatedConfigPassesValidation(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 15,
					"fps": 11,
					"seconds_per_clip": 1
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.NoErr(config.RunValidate())
}

func TestValidatePopulatedConfigFailsValidationForMissingPersistLocation(t *testing.T) {
	is := is.New(t)
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
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "PersistLoc" of type "string" using validator "empty=false"`)
}

func TestValidatePopulatedConfigFailsValiationForNonUniqueCameraTitles(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "TheSameNotUnique",
					"persist_location": "Nowhere",
					"max_clip_age_days": 15,
					"fps": 11,
					"seconds_per_clip": 1
				},
				{
					"title": "TheSameNotUnique",
					"persist_location": "Nowhere",
					"max_clip_age_days": 15,
					"fps": 11,
					"seconds_per_clip": 1
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), "validation failed: camera titles must be unique")
}

func TestValidatePopulatedConfigFailsValiationForMaxClipAgeDays(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 85,
					"fps": 11,
					"seconds_per_clip": 1
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "MaxClipAgeDays" of type "int" using validator "lte=30"`)
}

func TestValidatePopulatedConfigFailsValiationForFPSLessThan1(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 30,
					"fps": -4,
					"seconds_per_clip": 1
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "FPS" of type "int" using validator "gte=1"`)
}

func TestValidatePopulatedConfigFailsValiationForFPSMoreThan30(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 30,
					"fps": 39,
					"seconds_per_clip": 1
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "FPS" of type "int" using validator "lte=30"`)
}

func TestValidatePopulatedConfigFailsValiationForSPCLessThan1(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 30,
					"fps": 30,
					"seconds_per_clip":-5
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "SecondsPerClip" of type "int" using validator "gte=1"`)
}

func TestValidatePopulatedConfigFailsValiationForSPCMoreThan3(t *testing.T) {
	is := is.New(t)
	body := `{
			"max_clip_age_in_days": 1,
			"cameras": [
				{
					"title": "NotBlank",
					"persist_location": "Nowhere",
					"max_clip_age_days": 30,
					"fps": 30,
					"seconds_per_clip":12
				}
			]
		}`
	config := configdef.Values{}
	is.NoErr(json.Unmarshal([]byte(body), &config))
	is.Equal(config.RunValidate().Error(), `Validation error in field "SecondsPerClip" of type "int" using validator "lte=3"`)
}

func TestHasDupCameraTitlesDoesNotFindDuplicates(t *testing.T) {
	is := is.New(t)
	cameras := []configdef.Camera{}
	is.True(configdef.HasDupCameraTitles(cameras) == false)

	cameras = []configdef.Camera{
		{Title: "TestCam1"},
		{Title: "TestCam2"},
		{Title: "TestCam3"},
	}

	is.True(configdef.HasDupCameraTitles(cameras) == false)
}

func TestHasDupCameraTitlesDoesFindDuplicates(t *testing.T) {
	is := is.New(t)
	cameras := []configdef.Camera{
		{Title: "TestCam1"},
		{Title: "TestCam2"},
		{Title: "TestCam3"},
		{Title: "TestCam3"},
		{Title: "TestCam4"},
	}

	is.True(configdef.HasDupCameraTitles(cameras))
}

func TestHasDupCameraTitlesDoesFindDuplicateWithLargeGap(t *testing.T) {
	is := is.New(t)
	cameras := []configdef.Camera{
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

	is.True(configdef.HasDupCameraTitles(cameras))
}
