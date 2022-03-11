package orchestrate

import (
	"audit-tool-orchestrator/cmd/orchestrate/pool"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	orchestrateCmd := &cobra.Command{
		Use:   "orchestrate",
		Short: "",
		Long:  "",
	}

	orchestrateCmd.AddCommand(
		pool.NewCmd(),
	)

	return orchestrateCmd
}
