package config

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gopkg.in/dealancer/validate.v2"
)

var _ = Describe("Config", func() {
	var (
		mockValidConfigContent []byte
	)

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
	})

})
