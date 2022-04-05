package pipinstall_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/scribe"
	pipinstall "github.com/paketo-buildpacks/pip-install"
	"github.com/paketo-buildpacks/pip-install/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func testInstallProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		packagesLayerPath string
		cacheLayerPath    string
		workingDir        string
		executable        *fakes.Executable
		summer            *fakes.Summer

		pipInstallProcess pipinstall.PipInstallProcess
	)

	it.Before(func() {
		var err error
		packagesLayerPath, err = ioutil.TempDir("", "packages")
		Expect(err).NotTo(HaveOccurred())

		cacheLayerPath, err = ioutil.TempDir("", "cache")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "workingdir")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}
		summer = &fakes.Summer{}

		summer.SumCall.Returns.String = "fake-checksum"

		pipInstallProcess = pipinstall.NewPipInstallProcess(executable, scribe.NewEmitter(bytes.NewBuffer(nil)), summer)
	})

	context("Execute", func() {
		it("runs installation", func() {
			checksum, err := pipInstallProcess.Execute(workingDir, packagesLayerPath, cacheLayerPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(checksum).To(Equal("fake-checksum"))

			Expect(executable.ExecuteCall.Receives.Execution).To(MatchFields(IgnoreExtras, Fields{
				"Args": Equal([]string{
					"install",
					"--requirement",
					"requirements.txt",
					"--exists-action=w",
					fmt.Sprintf("--cache-dir=%s", cacheLayerPath),
					"--compile",
					"--user",
					"--disable-pip-version-check",
				}),
				"Dir": Equal(workingDir),
				"Env": ContainElement(fmt.Sprintf("PYTHONUSERBASE=%s", packagesLayerPath)),
			}))

			Expect(summer.SumCall.Receives.Paths).To(Equal([]string{
				packagesLayerPath,
			}))
		})

		context("when vendor directory exists", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(workingDir, "vendor"), os.ModeDir)).To(Succeed())
			})

			it("runs installation", func() {
				checksum, err := pipInstallProcess.Execute(workingDir, packagesLayerPath, cacheLayerPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(checksum).To(Equal("fake-checksum"))

				Expect(executable.ExecuteCall.Receives.Execution).To(MatchFields(IgnoreExtras, Fields{
					"Args": Equal([]string{
						"install",
						"--requirement",
						"requirements.txt",
						"--ignore-installed",
						"--exists-action=w",
						"--no-index",
						fmt.Sprintf("--find-links=%s", filepath.Join(workingDir, "vendor")),
						"--compile",
						"--user",
						"--disable-pip-version-check",
					}),
					"Dir": Equal(workingDir),
					"Env": ContainElement(fmt.Sprintf("PYTHONUSERBASE=%s", packagesLayerPath)),
				}))

				Expect(summer.SumCall.Receives.Paths).To(Equal([]string{
					packagesLayerPath,
				}))
			})
		})

		context("failure cases", func() {
			context("when vendor stat fails", func() {
				it.Before(func() {
					Expect(os.Chmod(workingDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					checksum, err := pipInstallProcess.Execute(workingDir, packagesLayerPath, cacheLayerPath)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
					Expect(checksum).To(Equal(""))
				})
			})

			context("when checksum fails", func() {
				it.Before(func() {
					summer.SumCall.Returns.String = "checksum-fail"
					summer.SumCall.Returns.Error = errors.New("fake error")
				})

				it("returns an error", func() {
					checksum, err := pipInstallProcess.Execute(workingDir, packagesLayerPath, cacheLayerPath)
					Expect(err).To(Equal(errors.New("fake error")))
					Expect(checksum).To(Equal("checksum-fail"))
				})
			})
		})
	})
}
