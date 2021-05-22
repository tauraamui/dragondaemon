package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config/schedule"
)

var _ = Describe("Config", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel
	var (
		mockValidConfigContent       []byte
		mockInvalidJSONConfigContent []byte
	)

	testCfg := values{
		fs: afero.NewMemMapFs(),
	}

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
	})

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
	})

	Describe("Loading into struct", func() {
		BeforeEach(func() {
			mockValidConfigContent = []byte(`{
				"debug": true,
				"secret": "test-secret",
				"max_clip_age_in_days": 7,
				"cameras": [
					{
						"title": "Test Cam 1",
						"address": "camera-network-addr",
						"fps": 1,
						"seconds_per_clip": 2,
						"disabled": false,
						"schedule": {
							"monday": {
								"on": "08:00:00",
								"off": "19:00:00"
							}
						}
					}
				]
			}`)

			mockInvalidJSONConfigContent = []byte(`{
				"debug" true,
			}`)

			mockValidationMissingRequiredFPSField = []byte(`{
				"max_clip_age_in_days": 1,
				"cameras": [
					{
						"title": "Test Cam 2"
					}
				]
			}`)
		})

		Describe("Loading config", func() {
			// TODO: replace the following two tests with unit tests for resolveConfigPath
			// It("Passes the expected ENV value for config location into file reader", func() {
			// 	// set the ENV var to known value
			// 	os.Setenv("DRAGON_DAEMON_CONFIG", "test-config-path-root/tacusci/dragondaemon/config.json")
			// 	defer os.Unsetenv("DRAGON_DAEMON_CONFIG")

			// 	testCfg.r = func(path string) ([]byte, error) {
			// 		Expect(path).To(Equal("test-config-path-root/tacusci/dragondaemon/config.json"))
			// 		return []byte{}, nil
			// 	}

			// 	testCfg.Load()
			// })

			// It("Passes the path from user config location into file reader", func() {
			// 	testCfg.uc = func() (string, error) {
			// 		return "user-config-path-root", nil
			// 	}
			// 	testCfg.r = func(path string) ([]byte, error) {
			// 		Expect(path).To(Equal("user-config-path-root/tacusci/dragondaemon/config.json"))
			// 		return []byte{}, nil
			// 	}

			// 	testCfg.Load()
			// })

			Context("From valid config JSON", func() {
				It("Should load valid config values", func() {
					afero.WriteFile(testCfg.fs, "test/tacusci/dragondaemon/config.json", mockValidConfigContent, 0666)
					defer testCfg.fs.Remove("test/tacusci/dragondaemon/config.json")
					testCfg.uc = func() (string, error) {
						return "test", nil
					}

					err := testCfg.Load()
					Expect(err).To(BeNil())
					Expect(testCfg.Secret).To(Equal("test-secret"))
					Expect(testCfg.MaxClipAgeInDays).To(Equal(7))
					Expect(testCfg.Cameras).To(Equal([]Camera{
						{
							Title:          "Test Cam 1",
							Address:        "camera-network-addr",
							FPS:            1,
							DateTimeLabel:  false,
							DateTimeFormat: "2006/01/02 15:04:05.999999999",
							SecondsPerClip: 2,
							Disabled:       false,
							Schedule: schedule.Schedule{
								Monday: schedule.OnOffTimes{
									On: func() *schedule.Time {
										t, _ := schedule.ParseTime("08:00:00")
										return &t
									}(),
									Off: func() *schedule.Time {
										t, _ := schedule.ParseTime("19:00:00")
										return &t
									}(),
								},
							},
						},
					}))
				})
			})

			Context("From failure to read config data", func() {
				It("Should handle read error gracefully and return wrapped error", func() {
					testCfg.uc = func() (string, error) {
						return "test", nil
					}

					err := testCfg.Load()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("Unable to read from path test/tacusci/dragondaemon/config.json: open test/tacusci/dragondaemon/config.json: file does not exist"))
				})
			})

			Context("From JSON unmarshal failure", func() {
				It("Should handle unmarshal error gracefully and return wrapped error", func() {
					afero.WriteFile(
						testCfg.fs,
						"test/tacusci/dragondaemon/config.json",
						mockInvalidJSONConfigContent,
						0666,
					)
					defer testCfg.fs.Remove("test/tacusci/dragondaemon/config.json")
					testCfg.uc = func() (string, error) {
						return "test", nil
					}

					err := testCfg.Load()
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("Parsing configuration file error: invalid character 't' after object key"))
				})
			})
		})
	})
})
