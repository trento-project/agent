package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/trento-project/agent/internal"
	"github.com/trento-project/agent/internal/factsengine"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

func NewFactsCmd() *cobra.Command {
	factsCmd := &cobra.Command{
		Use:   "facts",
		Short: "Run facts related operations",
	}

	factsCmd.AddCommand(NewFactsGatherCmd())

	return factsCmd
}

func NewFactsGatherCmd() *cobra.Command {
	gatherCmd := &cobra.Command{
		Use:   "gather",
		Short: "Gather the requested fact",
		Run:   gather,
		PersistentPreRunE: func(agentCmd *cobra.Command, _ []string) error {
			agentCmd.Flags().VisitAll(func(f *pflag.Flag) {
				viper.BindPFlag(f.Name, f)
			})

			return internal.InitConfig("agent")
		},
	}

	gatherCmd.Flags().String("gatherer", "", "The gatherer to use")
	gatherCmd.Flags().String("argument", "", "The used gatherer argument")
	gatherCmd.MarkFlagRequired("gatherer")
	gatherCmd.MarkFlagRequired("argument")

	return gatherCmd
}

func gather(*cobra.Command, []string) {
	var gatherer = viper.GetString("gatherer")
	var argument = viper.GetString("argument")

	engine := factsengine.NewFactsEngine("", "")

	g, err := engine.GetGatherer(gatherer)
	if err != nil {
		log.Fatal(err)
	}

	factRequest := []gatherers.FactRequest{
		{
			Name:     argument,
			Argument: argument,
		},
	}

	value, err := g.Gather(factRequest)
	if err != nil {
		log.Fatal(err)
	}

	result, err := factsengine.PrettifyFactResult(value[0])
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Gathererd fact for \"%s\" with argument \"%s\":", gatherer, argument)
	log.Printf("%s", result)
}
