package claim

// create and delete ClusterClaim resource

import (
	"audit-tool-orchestrator/pkg"
	"context"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claim",
		Short:   "",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	// TODO: check if delete flag is set
	// NOTE: there will be a claim per bundle run with name being set to the bundle.PackageName
	// TODO: add a label that is the bundle.name
	// TODO: ClusterClaim name needs unique ID; split UUID and add UUID[0]
	ctx := context.Background()

	cc := hivev1.ClusterClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ato-cluster-claim",
			Namespace: "hive",
		},
		Spec: hivev1.ClusterClaimSpec{
			ClusterPoolName: "ato-cluster-pool",
		},
	}

	hvclient := pkg.GetHiveClient()
	if err := hvclient.Create(ctx, &cc); err != nil {
		log.Errorf("Unable to create ClusterClaim: %v\n", err)
		return err
	}

	// Delete ClusterClaim resource
	// pass --delete and --claim
	return nil
}
