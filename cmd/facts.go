package cmd

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/trento-project/agent/internal"
	"github.com/trento-project/agent/internal/factsengine"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/utils"
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

			return internal.InitConfig("agent")
		},
	}

	gatherCmd.Flags().String("gatherer", "", "The gatherer to use")
	gatherCmd.Flags().String("argument", "", "The used gatherer argument")
	err := gatherCmd.MarkFlagRequired("gatherer")
	if err != nil {
		panic(err)
	}
	err = gatherCmd.MarkFlagRequired("argument")
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

			return internal.InitConfig("agent")
		},
	}

	return gatherCmd
}

func gather(*cobra.Command, []string) {
	var gatherer = viper.GetString("gatherer")
	var argument = viper.GetString("argument")
	var pluginsFolder = viper.GetString("plugins-folder")

	gathererManager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	log.Info("loading plugins")

	pluginLoaders := factsengine.NewPluginLoaders()

	gatherersFromPlugins, err := factsengine.GetGatherersFromPlugins(
		pluginLoaders,
		pluginsFolder,
	)
	if err != nil {
		log.Fatalf("Error loading gatherers from plugins: %s", err)
	}

	gathererManager.AddGatherers(gatherersFromPlugins)

	engine := factsengine.NewFactsEngine("", "", *gathererManager)

	defer factsengine.CleanupPlugins()

	g, err := gathererManager.GetGatherer(gatherer)
	if err != nil {
		cleanupAndFatal(engine, err)
	}

	factRequest := []entities.FactRequest{
		{
			Name:     argument,
			Argument: argument,
		},
	}

	value, err := g.Gather(factRequest)
	if err != nil {
		cleanupAndFatal(engine, err)
	}

	result, err := utils.PrettifyInterfaceToJSON(value[0])
	if err != nil {
		cleanupAndFatal(engine, err)
	}

	log.Printf("Gathered fact for \"%s\" with argument \"%s\":", gatherer, argument)
	log.Printf("%s", result)
}

func cleanupAndFatal(engine *factsengine.FactsEngine, err error) {
	factsengine.CleanupPlugins()
	log.Fatal(err)
}

func list(*cobra.Command, []string) {
	var pluginsFolder = viper.GetString("plugins-folder")

	gathererManager := gatherers.NewManager(map[string]gatherers.FactGatherer{
		gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
	})

	log.Info("loading plugins")

	pluginLoaders := factsengine.NewPluginLoaders()

	gatherersFromPlugins, err := factsengine.GetGatherersFromPlugins(
		pluginLoaders,
		pluginsFolder,
	)
	if err != nil {
		log.Fatalf("Error loading gatherers from plugins: %s", err)
	}

	gathererManager.AddGatherers(gatherersFromPlugins)

	defer factsengine.CleanupPlugins()

	gatherers := gathererManager.AvailableGatherers()

	log.Printf("Available gatherers:")

	for _, g := range gatherers {
		log.Printf(g)
	}
}
