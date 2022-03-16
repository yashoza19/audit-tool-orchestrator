package claim

// create and delete ClusterClaim resource

import (
	"audit-tool-orchestrator/pkg/orchestrate"
	"context"
	"github.com/google/uuid"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var flags = orchestrate.ClaimFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Create a Hive ClusterClaim resource.",
		Long: "Create a ClusterClaim resource to get a cluster from a ClusterPool. If a cluster is not available " +
			"because all clusters have been claimed or none are available due to errors a cluster will be created in " +
			"to fulfill this claim.",
		PreRunE: validation,
		RunE:    run,
	}

	cmd.Flags().StringVar(&flags.Name, "name", "ato-cluster-claim",
		"Name for the ClusterClaim resource. Required when --delete is set.")
	cmd.Flags().StringVar(&flags.Namespace, "namespace", "hive",
		"OpenShift project (namespace) the ClusterClaim should be created in.")
	cmd.Flags().StringVar(&flags.PoolName, "pool-name", "ato-cluster-pool",
		"ClusterPool to claim cluster from. If cluster is not available one will be created; ~45 minutes to create.")
	cmd.Flags().StringVar(&flags.BundleName, "bundle-name", "",
		"name SQLite column value for bundle this ClusterClaim is for.")
	cmd.Flags().BoolVar(&flags.Delete, "delete", false,
		"Delete the ClusterClaim provided by the name flag. If you do not provide the name and set "+
			"the --delete flag command will fail.")

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {
	flags.Name = strings.TrimSpace(flags.Name)

	if flags.Delete && (flags.Name == "ato-cluster-claim" || flags.Name == "") {
		return &orchestrate.ClusterClaimDeleteFlagSetNameFlagEmptyError{}
	}

	if len(flags.Name) < 8 || len(flags.Name) > 64 {
		return &orchestrate.ClusterClaimNameLengthIncorrectError{}
	}

	// Need to validate the name conforms to k8s naming convention
	/*validateClusterClaimName := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if err := validateClusterClaimName(flags.Name); err != false {
		return &orchestrate.ClusterClaimNameHasInvalidCharactersError{}
	}*/

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	hvclient := orchestrate.GetHiveClient()

	uniqueId := uuid.New()
	claimName := flags.Name + "-" + strings.Split(uniqueId.String(), "-")[0]

	if flags.Delete {
		claimName = flags.Name
	}

	cc := hivev1.ClusterClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: flags.Namespace,
			Labels:    map[string]string{"bundle-name": flags.BundleName},
		},
		Spec: hivev1.ClusterClaimSpec{
			ClusterPoolName: flags.PoolName,
		},
	}

	if flags.Delete {
		err := hvclient.HiveV1().ClusterClaims(flags.Namespace).Delete(ctx, claimName, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Unable to delete ClusterClaim %s: %v\n", flags.Name, err)
			return err
		}

		log.Infof("ClusterClaim %s deleted.\n", claimName)

		return nil
	} else {
		_, err := hvclient.HiveV1().ClusterClaims(flags.Namespace).Create(ctx, &cc, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Unable to create ClusterClaim %s: %v\n", claimName, err)
			return err
		}

		log.Infof("ClusterClaim %s submitted. Waiting for Pending and ClusterRunning statuses", claimName)
	}

	// ClusterClaim was submitted, we need to wait for Pending (False) and ClusterRunning (True) statuses
	claim, err := hvclient.HiveV1().ClusterClaims(flags.Namespace).Get(ctx, claimName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to get ClusterClaim: %v\n", err)
	}

	clusterDeployment, _ := orchestrate.WaitForSuccessfulClusterClaim(hvclient, claim)
	log.Infof(clusterDeployment)

	// clusterDeployment == "" is an error and we need to handle
	// clusterDeployment
	return nil
}
