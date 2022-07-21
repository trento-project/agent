package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCmd()
	cobra.CheckErr(rootCmd.Execute())
}

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{ //nolint
		Use:   "trento-agent",
		Short: "An open cloud-native web console improving on the life of SAP Applications administrators.",
		Long: `Trento is a web-based graphical user interface
that can help you deploy, provision and operate infrastructure for SAP Applications`,
	}

	rootCmd.PersistentFlags().
		String("config", "", "config file (default is $HOME/.trento.yaml)")
	rootCmd.PersistentFlags().
		String("log-level", "info", "then minimum severity (error, warn, info, debug) of logs to output")

	// Make global flags available in the children commands
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		err := viper.BindPFlag(f.Name, f)
		if err != nil {
			panic(errors.Wrap(err, "error during cli init"))
		}
	})

	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}
