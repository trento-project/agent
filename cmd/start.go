package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal/agent"
)

func NewStartCmd() *cobra.Command {
	var sshAddress string

	var clusterDiscoveryPeriod time.Duration
	var sapSystemDiscoveryPeriod time.Duration
	var cloudDiscoveryPeriod time.Duration
	var hostDiscoveryPeriod time.Duration
	var subscriptionDiscoveryPeriod time.Duration

	startCmd := &cobra.Command{ //nolint
		Use:   "start",
		Short: "Start the agent",
		Run:   start,
		PersistentPreRunE: func(agentCmd *cobra.Command, _ []string) error {
			agentCmd.Flags().VisitAll(func(f *pflag.Flag) {
				err := viper.BindPFlag(f.Name, f)
				if err != nil {
					panic(errors.Wrap(err, "error during cli init"))
				}
			})

			return agent.InitConfig("agent")
		},
	}

	startCmd.Flags().
		StringVar(
			&sshAddress,
			"ssh-address",
			"",
			"The address to which the trento-agent should be reachable for ssh connection by the runner for check execution.",
		)

	startCmd.Flags().
		String(
			"server-url",
			"http://localhost",
			"Trento server URL",
		)

	startCmd.Flags().
		String(
			"api-key",
			"",
			"API key provided by trento control plane. Allows communication",
		)

	startCmd.Flags().
		DurationVarP(
			&clusterDiscoveryPeriod,
			"cluster-discovery-period",
			"",
			10*time.Second,
			"Cluster discovery mechanism loop period in seconds",
		)

	startCmd.Flags().
		DurationVarP(
			&sapSystemDiscoveryPeriod,
			"sapsystem-discovery-period",
			"",
			10*time.Second,
			"SAP systems discovery mechanism loop period in seconds",
		)

	startCmd.Flags().
		DurationVarP(
			&cloudDiscoveryPeriod,
			"cloud-discovery-period",
			"",
			10*time.Second,
			"Cloud discovery mechanism loop period in seconds",
		)

	startCmd.Flags().
		DurationVarP(
			&hostDiscoveryPeriod,
			"host-discovery-period",
			"",
			10*time.Second,
			"Host discovery mechanism loop period in seconds",
		)

	startCmd.Flags().
		DurationVarP(
			&subscriptionDiscoveryPeriod,
			"subscription-discovery-period",
			"",
			900*time.Second,
			"Subscription discovery mechanism loop period in seconds",
		)
	err := startCmd.Flags().
		MarkHidden("subscription-discovery-period")
	if err != nil {
		panic(err)
	}

	startCmd.Flags().
		String("force-agent-id", "", "Agent ID. Used to mock the real ID for development purposes")
	err = startCmd.Flags().
		MarkHidden("force-agent-id")
	if err != nil {
		panic(err)
	}

	startCmd.Flags().String("facts-service-url", "amqp://guest:guest@localhost:5672", "Facts service queue url")

	return startCmd
}

func start(*cobra.Command, []string) {
	var err error

	ctx, ctxCancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	config, err := LoadConfig(afero.NewOsFs())
	if err != nil {
		log.Fatal("Failed to create the agent configuration: ", err)
	}

	a, err := agent.NewAgent(config)
	if err != nil {
		log.Fatal("Failed to create the agent: ", err)
	}

	go func() {
		quit := <-signals
		log.Infof("Caught %s signal!", quit)

		log.Info("Stopping the agent...")
		a.Stop(ctxCancel)
	}()

	log.Info("Starting the Console Agent...")
	err = a.Start(ctx)
	if err != nil {
		log.Fatal("Failed to start the agent: ", err)
	}
}
