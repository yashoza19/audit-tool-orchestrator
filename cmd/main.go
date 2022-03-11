package main

import (
	"audit-tool-orchestrator/cmd/index"
	"audit-tool-orchestrator/cmd/orchestrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "audit-tool-orchestrator",
		Short: "orchestrate running audit-tool against openshift",
		Long:  "",
	}

	rootCmd.AddCommand(index.NewCmd())
	rootCmd.AddCommand(orchestrate.NewCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
