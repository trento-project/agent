package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal"
)

func NewStartCmd() *cobra.Command {
	var sshAddress string
	var discoveryPeriod int

	var collectorHost string
	var collectorPort int

	var enablemTLS bool
	var cert string
	var key string
	var ca string

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the agent",
		Run:   start,
		PersistentPreRunE: func(agentCmd *cobra.Command, _ []string) error {
			agentCmd.Flags().VisitAll(func(f *pflag.Flag) {
				viper.BindPFlag(f.Name, f)
			})

			return internal.InitConfig("agent")
		},
	}

	startCmd.Flags().StringVar(&sshAddress, "ssh-address", "", "The address to which the trento-agent should be reachable for ssh connection by the runner for check execution.")

	startCmd.Flags().IntVarP(&discoveryPeriod, "discovery-period", "", 10, "Discovery mechanism loop period in seconds")

	startCmd.Flags().StringVar(&collectorHost, "collector-host", "localhost", "Data Collector host")
	startCmd.Flags().IntVar(&collectorPort, "collector-port", 8081, "Data Collector port")

	startCmd.Flags().BoolVar(&enablemTLS, "enable-mtls", false, "Enable mTLS authentication between server and agent")
	startCmd.Flags().StringVar(&cert, "cert", "", "mTLS client certificate")
	startCmd.Flags().StringVar(&key, "key", "", "mTLS client key")
	startCmd.Flags().StringVar(&ca, "ca", "", "mTLS Certificate Authority")

	return startCmd
}
