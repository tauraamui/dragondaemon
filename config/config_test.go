package config

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config/schedule"
)

var _ = Describe("Config", func() {
	existingLoggingLevel := logging.CurrentLoggingLevel
	var (
		mockValidConfigContent                []byte
		mockInvalidJSONConfigContent          []byte
		mockValidationMissingRequiredFPSField []byte
	)

	var testCfg values

	BeforeEach(func() {
		logging.CurrentLoggingLevel = logging.SilentLevel
		testCfg = values{
			fs: afero.NewMemMapFs(),
		}
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
			Context("resolveConfigPath", func() {
				It("Resolves the config path from ENV variable", func() {
					os.Setenv("DRAGON_DAEMON_CONFIG", "test/tacusci/dragondaemon/config.json")
					defer os.Unsetenv("DRAGON_DAEMON_CONFIG")
					Expect(resolveConfigPath(func() (string, error) {
						return "unused-test-config-path-root", nil
					})).To(Equal("test/tacusci/dragondaemon/config.json"))
				})

				It("Resolves the config path from user config location", func() {
					Expect(resolveConfigPath(func() (string, error) {
						return "test", nil
					})).To(Equal("test/tacusci/dragondaemon/config.json"))
				})
			})

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

			Context("Loading defaults", func() {
				It("Should assign values fields to default values", func() {
					testCfg.ResetToDefaults()
					Expect(testCfg.MaxClipAgeInDays).To(Equal(30))
					Expect(testCfg.Cameras).To(BeEmpty())
				})
			})

			Context("From failure to resolve config path", func() {
				It("Should handle path resolve error gracefully and return wrapped error", func() {
					testCfg.uc = func() (string, error) {
						return "", errors.New("error resolving user config dir")
					}
					err := testCfg.Load()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("unable to resolve config.json config file location: error resolving user config dir"))
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
					err := afero.WriteFile(
						testCfg.fs,
						"test/tacusci/dragondaemon/config.json",
						mockInvalidJSONConfigContent,
						0666,
					)
					Expect(err).To(BeNil())

					defer testCfg.fs.Remove("test/tacusci/dragondaemon/config.json")
					testCfg.uc = func() (string, error) {
						return "test", nil
					}

					err = testCfg.Load()
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("Parsing configuration file error: invalid character 't' after object key"))
				})
			})

			Context("From config validation failure", func() {
				It("Should handle validation error gracefully and return wrapped error", func() {
					err := afero.WriteFile(
						testCfg.fs,
						"test/tacusci/dragondaemon/config.json",
						mockValidationMissingRequiredFPSField,
						0666,
					)
					Expect(err).To(BeNil())

					defer testCfg.fs.Remove("test/tacusci/dragondaemon/config.json")
					testCfg.uc = func() (string, error) {
						return "test", nil
					}

					err = testCfg.Load()
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError(
						"Unable to validate configuration: Validation error in field \"FPS\" of type \"int\" using validator \"gte=1\"",
					))
				})
			})
		})
	})

	Describe("Writing struct to disk", func() {
		Describe("Saving config", func() {

			var testCfg values
			var filesystem afero.Fs
			testUserConfigResolver := func() (string, error) {
				return "test", nil
			}

			BeforeEach(func() {
				filesystem = afero.NewMemMapFs()
				testCfg = values{
					fs: filesystem,
					uc: testUserConfigResolver,
				}
			})

			It("Should write to file", func() {
				testCfg.MaxClipAgeInDays = 9
				testCfg.Cameras = []Camera{}
				path, err := testCfg.Save(false)

				Expect(path).To(Equal("test/tacusci/dragondaemon/config.json"))
				Expect(err).To(BeNil())

				configFile, err := testCfg.fs.Open("test/tacusci/dragondaemon/config.json")
				Expect(err).To(BeNil())
				defer configFile.Close()

				data, err := afero.ReadAll(configFile)
				Expect(err).To(BeNil())
				Expect(data).ToNot(BeEmpty())

				newConfig := values{
					fs: filesystem,
					uc: testUserConfigResolver,
				}
				err = newConfig.Load()
				Expect(err).To(BeNil())

				Expect(newConfig.MaxClipAgeInDays).To(Equal(9))
				Expect(newConfig.Cameras).To(BeEmpty())
			})

			It("Should handle path resolve error gracefully and return wrapped error", func() {
				testCfg.uc = func() (string, error) {
					return "", errors.New("error resolving user config dir")
				}
				path, err := testCfg.Save(true)
				Expect(path).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unable to resolve config.json config file location: error resolving user config dir"))
			})
		})
	})
})
