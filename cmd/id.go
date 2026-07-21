// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/trento-project/agent/internal/identity"
)

func NewAgentIDCmd() *cobra.Command {
	idCmd := &cobra.Command{
		Use:   "id",
		Short: "Print the agent identifier",
		RunE: func(_ *cobra.Command, _ []string) error {
			agentID, err := identity.GetAgentID(afero.NewOsFs())
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(agentID)

			return err
		},
	}

	return idCmd
}
