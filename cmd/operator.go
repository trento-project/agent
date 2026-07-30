// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/trento-project/agent/v3/internal/agent"
	"github.com/trento-project/agent/v3/internal/operations/operator"
	"github.com/trento-project/agent/v3/pkg/utils"
)

func NewOperatorCmd() *cobra.Command {
	operatorCmd := &cobra.Command{
		Use:    "operator",
		Short:  "Run operator related commands",
		Hidden: true,
	}

	operatorCmd.AddCommand(NewOperatorRunCmd())
	operatorCmd.AddCommand(NewOperatorListCmd())

	return operatorCmd
}

func NewOperatorRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run operator",
		Run:   runOperator,
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

	runCmd.Flags().StringP("operator", "o", "", "The operator to use")
	runCmd.Flags().StringP("arguments", "a", "", "The used operator arguments")

	err := runCmd.MarkFlagRequired("operator")
	if err != nil {
		panic(err)
	}

	return runCmd
}

func NewOperatorListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the available operators",
		Run:   listOperators,
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

	return listCmd
}

func runOperator(cmd *cobra.Command, _ []string) {
	var (
		operatorName = viper.GetString("operator")
		arguments    = viper.GetString("arguments")
		logger       = utils.NewDefaultLogger(
			viper.GetString("log-level"),
		)
	)

	slog.SetDefault(logger)
	slog.Info("Operation", "operator", operatorName, "arguments", arguments)

	opArgs := make(operator.Arguments)

	err := json.Unmarshal([]byte(arguments), &opArgs)
	if err != nil {
		logger.Error("error unmarshalling arguments", "err", err)
		os.Exit(1)
	}

	registry := operator.StandardRegistry()

	operatorBuilder, err := registry.GetOperatorBuilder(operatorName)
	if err != nil {
		logger.Error("error building operator", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(cmd.Context())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		slog.Info("Caught signal!")
		cancel()
	}()

	op := operatorBuilder("", opArgs)
	report := op.Run(ctx)

	if ctx.Err() != nil {
		slog.Info("Operation cancelled")

		return
	}

	if report.Error != nil {
		logger.Error(report.Error.Error())
		os.Exit(0)
	}

	diff, err := json.Marshal(report.Success.Diff)
	if err != nil {
		logger.Error("error marshalling diff output", "err", err)
		os.Exit(1)
	}

	logger.Info("Operation succeeded",
		"phase", report.Success.LastPhase,
		"diff", string(diff),
	)
}

func listOperators(*cobra.Command, []string) {
	var logger = utils.NewDefaultLogger(
		viper.GetString("log-level"),
	)

	slog.SetDefault(logger)

	registry := operator.StandardRegistry()
	operators := registry.AvailableOperators()

	slog.Info("Available operators:")

	for _, o := range operators {
		slog.Info(o)
	}
}
