package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/trento-project/agent/version"
)

func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{ //nolint
		Use:   "version",
		Short: "Print the version number of Trento",
		Long:  `All software has versions. This is Trento's`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Trento installed from %s version %s\nbuilt with %s %s/%s\n", version.InstallationSource, version.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH) //nolint
		},
	}

	return versionCmd
}
