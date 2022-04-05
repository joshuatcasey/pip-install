package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	pipinstall "github.com/paketo-buildpacks/pip-install"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout)

	packit.Run(
		pipinstall.Detect(),
		pipinstall.Build(
			draft.NewPlanner(),
			pipinstall.NewPipInstallProcess(pexec.NewExecutable("pip"), logger, fs.NewChecksumCalculator()),
			pipinstall.NewSiteProcess(pexec.NewExecutable("python")),
			logger,
		),
	)
}
