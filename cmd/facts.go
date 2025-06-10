package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

func NewFactsCmd() *cobra.Command {
	factsCmd := &cobra.Command{ //nolint
		Use:   "facts",
		Short: "Run facts related operations",
	}

	factsCmd.AddCommand(NewFactsGatherCmd())
	factsCmd.AddCommand(NewFactsListCmd())

	return factsCmd
}

func NewFactsGatherCmd() *cobra.Command {
	gatherCmd := &cobra.Command{ //nolint
		Use:   "gather",
		Short: "Gather the requested fact",
		Run:   gather,
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

	gatherCmd.Flags().String("gatherer", "", "The gatherer to use")
	gatherCmd.Flags().String("argument", "", "The used gatherer argument")
	err := gatherCmd.MarkFlagRequired("gatherer")
	if err != nil {
		panic(err)
	}

	return gatherCmd
}

func NewFactsListCmd() *cobra.Command {
	gatherCmd := &cobra.Command{ //nolint
		Use:   "list",
		Short: "List the available gatherers",
		Run:   list,
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

	return gatherCmd
}

func gather(cmd *cobra.Command, _ []string) {
	var gatherer = viper.GetString("gatherer")
	var argument = viper.GetString("argument")
	var pluginsFolder = viper.GetString("plugins-folder")
	var logger = utils.NewDefaultLogger(
		viper.GetString("log-level"),
	)

	slog.SetDefault(logger)

	gathererRegistry := gatherers.NewRegistry(gatherers.StandardGatherers())

	slog.Info("loading plugins")

	pluginLoaders := gatherers.PluginLoaders{
		"rpc": &gatherers.RPCPluginLoader{},
	}

	gatherersFromPlugins, err := gatherers.GetGatherersFromPlugins(
		pluginLoaders,
		pluginsFolder,
	)
	if err != nil {
		slog.Error("Error loading gatherers from plugins", "error", err)
		os.Exit(1)
	}

	gathererRegistry.AddGatherers(gatherersFromPlugins)

	defer gatherers.CleanupPlugins()

	g, err := gathererRegistry.GetGatherer(gatherer)
	if err != nil {
		cleanupAndFatal(err)
	}

	factRequest := []entities.FactRequest{
		{
			Name:     argument,
			Argument: argument,
		},
	}

	ctx, cancel := context.WithCancel(cmd.Context())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	cancelled := false
	go func() {
		<-signals
		slog.Info("Caught signal!")
		cancelled = true
		cancel()

	}()

	value, err := g.Gather(ctx, factRequest)

	if cancelled {
		slog.Info("Gathering cancelled")
		return
	}

	if err != nil {
		slog.Error("Error gathering fact", "gatherer", gatherer, "argument", argument)
		cleanupAndFatal(err)
	}

	if len(value) != 1 {
		slog.Info("No value gathered", "gatherer", gatherer, "argument", argument)
		return
	}

	if value[0].Error != nil {
		slog.Error("Error gathering fact", "gatherer", gatherer, "argument", argument)
		cleanupAndFatal(value[0].Error)
	}

	result, err := value[0].Prettify()
	if err != nil {
		cleanupAndFatal(err)
	}

	slog.Info("Gathered fact", "gatherer", gatherer, "argument", argument)
	slog.Info(result)
}

func cleanupAndFatal(err error) {
	gatherers.CleanupPlugins()
	slog.Error(err.Error())
	os.Exit(1)
}

func list(*cobra.Command, []string) {
	var pluginsFolder = viper.GetString("plugins-folder")
	var logger = utils.NewDefaultLogger(
		viper.GetString("log-level"),
	)

	slog.SetDefault(logger)

	gathererRegistry := gatherers.NewRegistry(gatherers.StandardGatherers())

	slog.Info("loading plugins")

	pluginLoaders := gatherers.PluginLoaders{
		"rpc": &gatherers.RPCPluginLoader{},
	}

	gatherersFromPlugins, err := gatherers.GetGatherersFromPlugins(
		pluginLoaders,
		pluginsFolder,
	)
	if err != nil {
		slog.Error("Error loading gatherers from plugins", "error", err)
	}

	gathererRegistry.AddGatherers(gatherersFromPlugins)

	defer gatherers.CleanupPlugins()

	gatherers := gathererRegistry.AvailableGatherers()

	slog.Info("Available gatherers:")

	for _, g := range gatherers {
		slog.Info(g)
	}
}
