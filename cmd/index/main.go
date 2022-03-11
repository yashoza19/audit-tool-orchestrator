package index

import (
	"audit-tool-orchestrator/cmd/index/bundles"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "index has subcommands specific to working with catalog indexes",
		Long:  "",
	}

	indexCmd.AddCommand(
		bundles.NewCmd(),
	)

	return indexCmd
}
