package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal"
)

func NewStartCmd() *cobra.Command {
	var sshAddress string

	var clusterDiscoveryPeriod time.Duration
	var sapSystemDiscoveryPeriod time.Duration
	var cloudDiscoveryPeriod time.Duration
	var hostDiscoveryPeriod time.Duration
	var subscriptionDiscoveryPeriod time.Duration

	var serverURL string
	var apiKey string

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

	startCmd.Flags().StringVar(&serverURL, "server-url", "http://localhost", "Trento server URL")
	startCmd.Flags().StringVar(&apiKey, "api-key", "", "API key provided by trento control plane. Allows communication")

	startCmd.Flags().DurationVarP(&clusterDiscoveryPeriod, "cluster-discovery-period", "", 10*time.Second, "Cluster discovery mechanism loop period in seconds")
	startCmd.Flags().DurationVarP(&sapSystemDiscoveryPeriod, "sapsystem-discovery-period", "", 10*time.Second, "SAP systems discovery mechanism loop period in seconds")
	startCmd.Flags().DurationVarP(&cloudDiscoveryPeriod, "cloud-discovery-period", "", 10*time.Second, "Cloud discovery mechanism loop period in seconds")
	startCmd.Flags().DurationVarP(&hostDiscoveryPeriod, "host-discovery-period", "", 10*time.Second, "Host discovery mechanism loop period in seconds")
	startCmd.Flags().DurationVarP(&subscriptionDiscoveryPeriod, "subscription-discovery-period", "", 900*time.Second, "Subscription discovery mechanism loop period in seconds")

	startCmd.Flags().MarkHidden("subscription-discovery-period")

	startCmd.Flags().Bool("checksengine", false, "Enable the checksengine")
	startCmd.Flags().MarkHidden("checksengine")

	return startCmd
}
