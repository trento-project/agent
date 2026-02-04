package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/pkg/utils"
)

func NewStartCmd() *cobra.Command {
	var clusterDiscoveryPeriod time.Duration
	var sapSystemDiscoveryPeriod time.Duration
	var cloudDiscoveryPeriod time.Duration
	var hostDiscoveryPeriod time.Duration
	var subscriptionDiscoveryPeriod time.Duration
	var saptuneDiscoveryPeriod time.Duration
	var logger = utils.NewDefaultLogger(
		viper.GetString("log-level"),
	)

	slog.SetDefault(logger)

	startCmd := &cobra.Command{ //nolint
		Use:   "start",
		Short: "Start the agent",
		Run:   start,
		PersistentPreRunE: func(agentCmd *cobra.Command, _ []string) error {
			agentCmd.Flags().VisitAll(func(f *pflag.Flag) {
				err := viper.BindPFlag(f.Name, f)
				if err != nil {
					panic(fmt.Errorf("error during cli init: %w", err))
				}
			})

			return agent.InitConfig("agent")
		},
	}

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
		DurationVarP(
			&saptuneDiscoveryPeriod,
			"saptune-discovery-period",
			"",
			900*time.Second,
			"Saptune discovery mechanism loop period in seconds",
		)

	startCmd.Flags().
		String("force-agent-id", "", "Agent ID. Used to mock the real ID for development purposes")
	err = startCmd.Flags().
		MarkHidden("force-agent-id")
	if err != nil {
		panic(err)
	}

	startCmd.Flags().String("facts-service-url", "amqp://guest:guest@localhost:5672", "Facts service queue url")

	startCmd.Flags().
		String(
			"node-exporter-target",
			"",
			"Node exporter target address in ip:port notation. If not given the lowest "+
				"ipv4 address with the default 9100 port is used",
		)

	startCmd.Flags().
		String(
			"prometheus-url",
			"",
			"Prometheus URL for push mode. If provided, the agent operates in push mode",
		)

	return startCmd
}

func start(*cobra.Command, []string) {
	var err error

	ctx, ctxCancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	config, err := LoadConfig(afero.NewOsFs())
	if err != nil {
		slog.Error("Failed to create the agent configuration", "error", err)
		os.Exit(1)
	}

	a, err := agent.NewAgent(config)
	if err != nil {
		slog.Error("Failed to create the agent", "error", err)
		os.Exit(1)
	}

	go func() {
		quit := <-signals
		slog.Info("Caught signal!", "signal", quit)

		slog.Info("Stopping the agent...")
		a.Stop(ctxCancel)
	}()

	slog.Info("Starting the Console Agent...")
	err = a.Start(ctx)
	if err != nil {
		slog.Error("Failed to start the agent", "error", err)
	}
}
