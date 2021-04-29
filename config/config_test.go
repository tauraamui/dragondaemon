package config

import (
	"encoding/json"
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

var _ = Describe("Config", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel
	var (
		mockValidConfigContent                []byte
		mockInvalidJSONConfigContent          []byte
		mockValidationMissingRequiredFPSField []byte
	)

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel

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

	AfterEach(func() {
		logging.CurrentLoggingLevel = existingLoggingLevel
	})

	Describe("Loading config", func() {
		It("Passes the expected ENV value for config location into file reader", func() {
			// set the ENV var to known value
			os.Setenv("DRAGON_DAEMON_CONFIG", "test-config-path")

			cfg := values{
				r: func(path string) ([]byte, error) {
					Expect(path).To(Equal("test-config-path"))
					return []byte{}, nil
				},
				um: json.Unmarshal,
				v:  validate.Validate,
			}

			cfg.Load()
		})

		Context("From valid config JSON", func() {
			It("Should load valid config values", func() {
				cfg := values{
					r: func(string) ([]byte, error) {
						return mockValidConfigContent, nil
					},
					um: json.Unmarshal,
					v:  validate.Validate,
				}

				err := cfg.Load()
				Expect(err).To(BeNil())
				Expect(cfg.Secret).To(Equal("test-secret"))
				Expect(cfg.MaxClipAgeInDays).To(Equal(7))
				Expect(cfg.Cameras).To(Equal([]Camera{
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
				cfg := values{
					r: func(string) ([]byte, error) {
						return nil, errors.New("read failure")
					},
				}

				err := cfg.Load()
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError("Unable to read from path test-config-path: read failure"))
			})
		})

		Context("From JSON unmarshal failure", func() {
			It("Should handle unmarshal error gracefully and return wrapped error", func() {
				cfg := values{
					r: func(string) ([]byte, error) {
						return mockInvalidJSONConfigContent, nil
					},
					um: json.Unmarshal,
				}

				err := cfg.Load()
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError("Parsing configuration file error: invalid character 't' after object key"))
			})
		})

		Context("From config validation failure", func() {
			It("Should handle validation error gracefully and return wrapped error", func() {
				cfg := values{
					r: func(string) ([]byte, error) {
						return mockValidationMissingRequiredFPSField, nil
					},
					um: json.Unmarshal,
					v:  validate.Validate,
				}

				err := cfg.Load()
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError(
					"Unable to validate configuration: Validation error in field \"FPS\" of type \"int\" using validator \"gte=1\"",
				))
			})
		})
	})

})
