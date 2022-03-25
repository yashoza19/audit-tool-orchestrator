package orchestrate

import (
	"audit-tool-orchestrator/cmd/orchestrate/claim"
	"audit-tool-orchestrator/cmd/orchestrate/job"
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
		claim.NewCmd(),
		job.NewCmd(),
	)

	return orchestrateCmd
}
