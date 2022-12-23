package cmd

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/trento-project/agent/internal/agent"
)

func NewAgentIDCmd() *cobra.Command {
	idCmd := &cobra.Command{ //nolint
		Use:   "id",
		Short: "Print the agent identifier",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, err := agent.GetAgentID(afero.NewOsFs())
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(agentID)

			return err
		},
	}

	return idCmd
}
