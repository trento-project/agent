package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/trento-project/agent/internal"
)

// rootCmd represents the base command when called without any subcommands

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCmd()
	cobra.CheckErr(rootCmd.Execute())
}

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "trento-agent",
		Short: "An open cloud-native web console improving on the life of SAP Applications administrators.",
		Long: `Trento is a web-based graphical user interface
that can help you deploy, provision and operate infrastructure for SAP Applications`,
	}
	var cfgFile string
	var logLevel string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.trento.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "then minimum severity (error, warn, info, debug) of logs to output")

	// Make global flags available in the children commands
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		//viper.BindPFlag(f.Name, f)
	})

	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}

func start(*cobra.Command, []string) {
	var err error

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to create the agent configuration: ", err)
	}

	a, err := internal.NewAgent(config)
	if err != nil {
		log.Fatal("Failed to create the agent: ", err)
	}

	go func() {
		quit := <-signals
		log.Printf("Caught %s signal!", quit)

		log.Println("Stopping the agent...")
		a.Stop()
	}()

	log.Println("Starting the Console Agent...")
	err = a.Start()
	if err != nil {
		log.Fatal("Failed to start the agent: ", err)
	}
}
