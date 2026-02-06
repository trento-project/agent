package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/pkg/utils"
)

func NewGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{ //nolint
		Use:   "generate",
		Short: "Generate configuration files",
	}

	generateCmd.AddCommand(NewGenerateAlloyCmd())

	return generateCmd
}

func NewGenerateAlloyCmd() *cobra.Command {
	alloyCmd := &cobra.Command{ //nolint
		Use:   "alloy",
		Short: "Generate Grafana Alloy configuration for Trento metrics",
		Long: `Generate Grafana Alloy configuration for pushing system metrics to Prometheus.

The generated configuration is printed to stdout, allowing users to review
and deploy it as appropriate for their environment.

Example usage:
  # Generate and review configuration
  trento-agent generate alloy \
  	--prometheus-url https://prometheus.example.com/api/v1/write \
  	--prometheus-mode push \
  	--prometheus-auth basic \
  	--prometheus-auth-username myuser \
  	--prometheus-auth-password mypassword

  # Generate and deploy to Alloy config directory
  trento-agent generate alloy | sudo tee /etc/alloy/config.d/trento.alloy > /dev/null`,
		RunE: generateAlloy,
		PersistentPreRunE: func(agentCmd *cobra.Command, _ []string) error {
			// Use stderr logger to keep stdout clean for piping
			logger := utils.NewStderrLogger(viper.GetString("log-level"))
			slog.SetDefault(logger)

			agentCmd.Flags().VisitAll(func(f *pflag.Flag) {
				err := viper.BindPFlag(f.Name, f)
				if err != nil {
					panic(fmt.Errorf("error during cli init: %w", err))
				}
			})

			return agent.InitConfig("agent")
		},
		SilenceUsage: true,
	}

	alloyCmd.Flags().
		String(
			"prometheus-mode",
			"",
			"Prometheus mode (must be 'push' for alloy configuration)",
		)

	alloyCmd.Flags().
		String(
			"prometheus-url",
			"",
			"The Prometheus remote write endpoint URL",
		)

	alloyCmd.Flags().
		String(
			"prometheus-auth",
			agent.DefaultAuthMethod,
			"Authentication method: none, basic, bearer, mtls",
		)

	alloyCmd.Flags().
		String(
			"prometheus-auth-username",
			"",
			"Username for basic authentication",
		)

	alloyCmd.Flags().
		String(
			"prometheus-auth-password",
			"",
			"Password for basic authentication",
		)

	alloyCmd.Flags().
		String(
			"prometheus-auth-bearer-token",
			"",
			"Bearer token for bearer authentication",
		)

	alloyCmd.Flags().
		String(
			"prometheus-tls-ca-cert",
			"",
			"Path to CA certificate file for TLS verification",
		)

	alloyCmd.Flags().
		String(
			"prometheus-tls-client-cert",
			"",
			"Path to client certificate file for mTLS authentication",
		)

	alloyCmd.Flags().
		String(
			"prometheus-tls-client-key",
			"",
			"Path to client private key file for mTLS authentication",
		)

	alloyCmd.Flags().
		Duration(
			"prometheus-scrape-interval",
			agent.DefaultScrapeInterval,
			"Metrics scrape interval (e.g., 15s, 1m)",
		)

	alloyCmd.Flags().
		String(
			"prometheus-exporter-name",
			agent.DefaultExporterName,
			"Name used as the exporter_name label in metrics",
		)

	alloyCmd.Flags().
		String(
			"force-agent-id",
			"",
			"Override the automatically determined agent ID (use only for development/testing)",
		)
	err := alloyCmd.Flags().MarkHidden("force-agent-id")
	if err != nil {
		panic(err)
	}

	return alloyCmd
}

func generateAlloy(_ *cobra.Command, _ []string) error {
	prometheusMode := viper.GetString("prometheus-mode")
	if prometheusMode != "push" {
		return fmt.Errorf("prometheus-mode must be 'push' for alloy configuration, got '%s'", prometheusMode)
	}

	agentID := viper.GetString("force-agent-id")
	if agentID == "" {
		id, err := agent.GetAgentID(afero.NewOsFs())
		if err != nil {
			return fmt.Errorf("could not get the agent ID: %w", err)
		}
		agentID = id
	}

	config := &agent.AlloyConfig{
		AgentID:         agentID,
		PrometheusURL:   viper.GetString("prometheus-url"),
		ScrapeInterval:  viper.GetDuration("prometheus-scrape-interval"),
		ExporterName:    viper.GetString("prometheus-exporter-name"),
		AuthMethod:      viper.GetString("prometheus-auth"),
		AuthUsername:    viper.GetString("prometheus-auth-username"),
		AuthPassword:    viper.GetString("prometheus-auth-password"),
		AuthBearerToken: viper.GetString("prometheus-auth-bearer-token"),
		TLSCACert:       viper.GetString("prometheus-tls-ca-cert"),
		TLSClientCert:   viper.GetString("prometheus-tls-client-cert"),
		TLSClientKey:    viper.GetString("prometheus-tls-client-key"),
	}

	return agent.GenerateAlloyConfig(os.Stdout, config)
}
