package manifest_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/types"
	. "code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	var manifest string

	Describe("ReadAndInterpolateManifest", func() {
		var (
			pathToManifest string
			pathToVarsFile string
			apps           []Application
			executeErr     error
		)

		BeforeEach(func() {
			pathToVarsFile = ""
			manifest = `---
applications:
- name: app-1
  buildpack: "some-buildpack"
  command: "some-command"
  health-check-http-endpoint: "\\some-endpoint"
  health-check-type: "http"
  instances: 10
  disk_quota: 100M
  docker:
    image: "some-docker-image"
    username: "some-docker-username"
  memory: 200M
  random-route: true
  stack: "some-stack"
  timeout: 120
- name: "app-2"
  buildpack: default
  disk_quota: 1G
  instances: 0
  memory: 2G
  routes:
  - route: foo.bar.com
  - route: baz.qux.com
  - route: blep.blah.com/boop
  services:
  - service_1
  - service_2
- name: "app-3"
  no-route: true
  env:
    env_1: 'foo'
    env_2: 182837403930483038
    env_3: true
    env_4: 1.00001
- name: "app-4"
  buildpack: null
  command: null
- name: "app-5"
  domain: "some-domain"
  domains:
  - domain_1
  - domain_2
- name: "app-6"
  host: "some-hostname"
  hosts:
  - hostname_1
  - hostname_2
  no-hostname: true
- name: "app-7"
  routes:
  - route: hello.com
  - route: bleep.blah.com
  random-route: true
`

			tempFile, err := ioutil.TempFile("", "manifest-test-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()

			err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
			Expect(err).ToNot(HaveOccurred())

		})

		AfterEach(func() {
			Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
			Expect(os.RemoveAll(pathToVarsFile)).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			apps, executeErr = ReadAndInterpolateManifest(pathToManifest, pathToVarsFile)
		})

		Context("when the manifest does not contain deprecated fields", func() {
			It("returns a merged set of applications", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(apps).To(HaveLen(7))

				Expect(apps[0]).To(Equal(Application{
					Name: "app-1",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "some-buildpack",
					},
					Command: types.FilteredString{
						IsSet: true,
						Value: "some-command",
					},
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances: types.NullInt{
						Value: 10,
						IsSet: true,
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 100,
						IsSet: true,
					},
					DockerImage:    "some-docker-image",
					DockerUsername: "some-docker-username",
					Memory: types.NullByteSizeInMb{
						Value: 200,
						IsSet: true,
					},
					RandomRoute:        true,
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				}))

				Expect(apps[1]).To(Equal(Application{
					Name: "app-2",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "",
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 1024,
						IsSet: true,
					},
					Instances: types.NullInt{
						IsSet: true,
						Value: 0,
					},
					Memory: types.NullByteSizeInMb{
						Value: 2048,
						IsSet: true,
					},
					Routes:   []string{"foo.bar.com", "baz.qux.com", "blep.blah.com/boop"},
					Services: []string{"service_1", "service_2"},
				}))

				Expect(apps[2]).To(Equal(Application{
					Name: "app-3",
					EnvironmentVariables: map[string]string{
						"env_1": "foo",
						"env_2": "182837403930483038",
						"env_3": "true",
						"env_4": "1.00001",
					},
					NoRoute: true,
				}))

				Expect(apps[3]).To(Equal(Application{
					Name: "app-4",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "",
					},
					Command: types.FilteredString{
						IsSet: true,
						Value: "",
					},
				}))

				Expect(apps[4].Name).To(Equal("app-5"))
				Expect(apps[4].DeprecatedDomain).ToNot(BeNil())
				Expect(apps[4].DeprecatedDomains).ToNot(BeNil())

				Expect(apps[5].Name).To(Equal("app-6"))
				Expect(apps[5].DeprecatedHost).ToNot(BeNil())
				Expect(apps[5].DeprecatedHosts).ToNot(BeNil())
				Expect(apps[5].DeprecatedNoHostname).ToNot(BeNil())

				Expect(apps[6]).To(Equal(Application{
					Name:        "app-7",
					Routes:      []string{"hello.com", "bleep.blah.com"},
					RandomRoute: true,
				}))
			})
		})

		Context("when provided deprecated fields", func() {

			Context("when global fields are provided", func() {
				DescribeTable("raises a GlobalFieldsError",
					func(manifestProperty string, numberOfValues int) {
						tempFile, err := ioutil.TempFile("", "manifest-test-")
						Expect(err).ToNot(HaveOccurred())
						defer os.Remove(tempFile.Name())
						Expect(tempFile.Close()).ToNot(HaveOccurred())
						pathToManifest = tempFile.Name()

						if numberOfValues == 1 {
							manifest = fmt.Sprintf("---\n%s: value", manifestProperty)
						} else {
							values := []string{"A", "B"}
							manifest = fmt.Sprintf("---\n%s: [%s]", manifestProperty, strings.Join(values, ","))
						}
						err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
						Expect(err).ToNot(HaveOccurred())

						_, err = ReadAndInterpolateManifest(pathToManifest, pathToVarsFile)
						Expect(err).To(MatchError(GlobalFieldsError{Fields: []string{manifestProperty}}))
					},

					Entry("global buildpack", "buildpack", 1),
					Entry("global command", "command", 1),
					Entry("global disk quota", "disk_quota", 1),
					Entry("global docker", "docker", 1),
					Entry("global domain", "domain", 1),
					Entry("global domains", "domains", 2),
					Entry("global environment variables", "env", 2),
					Entry("global health check HTTP endpoint", "health-check-http-endpoint", 1),
					Entry("global health check timeout", "timeout", 1),
					Entry("global health check type", "health-check-type", 1),
					Entry("global host", "host", 1),
					Entry("global hosts", "hosts", 2),
					Entry("global instances", "instances", 1),
					Entry("global memory", "memory", 1),
					Entry("global name", "name", 1),
					Entry("global no hostname", "no-hostname", 1),
					Entry("global no route", "no-route", 1),
					Entry("global path", "path", 1),
					Entry("global random-route", "random-route", 1),
					Entry("global routes", "routes", 2),
					Entry("global services", "services", 2),
					Entry("global stack", "stack", 1),
				)
			})

		})

		Context("when inheritance is provided", func() {
			BeforeEach(func() {
				manifest = `---
inherit: "./some-inheritance-file"
applications:
- name: "app-1"
`
				tempFile, err := ioutil.TempFile("", "manifest-test-")
				Expect(err).ToNot(HaveOccurred())
				Expect(tempFile.Close()).ToNot(HaveOccurred())
				pathToManifest = tempFile.Name()

				err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("raises an InheritanceFieldError", func() {
				Expect(executeErr).To(MatchError(InheritanceFieldError{}))
			})
		})

		Context("when a vars file is provided", func() {
			BeforeEach(func() {
				manifest = `---
applications:
- name: ((var1))
`
				err := ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
				Expect(err).ToNot(HaveOccurred())

				vars := `var1: app-1`
				varFile, err := ioutil.TempFile("", "vars-test-")
				Expect(err).ToNot(HaveOccurred())
				Expect(varFile.Close()).ToNot(HaveOccurred())
				pathToVarsFile = varFile.Name()

				err = ioutil.WriteFile(pathToVarsFile, []byte(vars), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the provided file exists and contains valid yaml", func() {
				It("interpolates the placeholder values", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps[0].Name).To(Equal("app-1"))
				})
			})

			Context("when the provided file path does not exist", func() {
				BeforeEach(func() {
					pathToVarsFile = "garbagepath"
				})

				It("returns an error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(os.IsNotExist(executeErr)).To(BeTrue())
				})
			})

			Context("when the provided file is not a valid yaml file", func() {
				BeforeEach(func() {
					vars := `: bad`
					err := ioutil.WriteFile(pathToVarsFile, []byte(vars), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(InvalidYAMLError{
						Err: errors.New("yaml: did not find expected key"),
					}))
				})
			})
		})

		Context("when no vars file is provided", func() {
			BeforeEach(func() {
				manifest = `---
applications:
- name: ((var1))
`
				err := ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the merged, uninterpolated applications", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(apps[0].Name).To(Equal("((var1))"))
			})
		})
	})

	Describe("WriteApplicationManifest", func() {
		var (
			application Application
			tmpDir      string
			filePath    string

			executeErr error
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "manifest-test-")
			Expect(err).NotTo(HaveOccurred())
			filePath = filepath.Join(tmpDir, "manifest.yml")
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		JustBeforeEach(func() {
			executeErr = WriteApplicationManifest(application, filePath)
		})

		Context("when all app properties are provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "some-buildpack",
					},
					Command: types.FilteredString{
						IsSet: true,
						Value: "some-command",
					},
					DockerImage:    "some-docker-image",
					DockerUsername: "some-docker-username",
					DockerPassword: "",
					EnvironmentVariables: map[string]string{
						"env_1": "foo",
						"env_2": "182837403930483038",
						"env_3": "true",
						"env_4": "1.00001",
					},
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances: types.NullInt{
						Value: 10,
						IsSet: true,
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 1024,
						IsSet: true,
					},
					Memory: types.NullByteSizeInMb{
						Value: 200,
						IsSet: true,
					},
					NoRoute:            true,
					Routes:             []string{"foo.bar.com", "baz.qux.com", "blep.blah.com/boop"},
					Services:           []string{"service_1", "service_2"},
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				}
			})

			It("creates and writes the manifest to the specified filepath", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  no-route: true
  routes:
  - route: foo.bar.com
  - route: baz.qux.com
  - route: blep.blah.com/boop
  services:
  - service_1
  - service_2
  stack: some-stack
  timeout: 120
`))
			})
		})

		Context("when some properties are not provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
				}
			})

			It("does not save them in manifest", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		Context("when the file is a relative path", func() {
			var pwd string

			BeforeEach(func() {
				var err error
				pwd, err = os.Getwd()
				Expect(err).ToNot(HaveOccurred())

				filePath = "./manifest.yml"
				Expect(os.Chdir(tmpDir)).To(Succeed())

				application = Application{
					Name: "app-1",
				}
			})

			AfterEach(func() {
				Expect(os.Chdir(pwd)).To(Succeed())
			})

			It("writes the file in an expanded path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filepath.Join(tmpDir, "manifest.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		Context("when the file already exists", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filePath, []byte(`{}`), 0644)
				Expect(err).ToNot(HaveOccurred())
				application = Application{
					Name: "app-1",
				}
			})

			Context("writes the file", func() {
				It("truncates and writes the manifest to specified filepath", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					manifestBytes, err := ioutil.ReadFile(filePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
				})
			})
		})
	})
})
