package pool

// create ClusterPool resource

import (
	"audit-tool-orchestrator/pkg/orchestrate"
	"context"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/aws"
	"github.com/openshift/hive/apis/hive/v1/azure"
	"github.com/openshift/hive/apis/hive/v1/gcp"
	"github.com/openshift/hive/apis/hive/v1/ibmcloud"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var flags = orchestrate.PoolFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pool",
		Short:   "Create a Hive ClusterPool resource.",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	cmd.Flags().StringVar(&flags.Name, "name", "ato-cluster-pool",
		"")
	cmd.Flags().StringVar(&flags.Namespace, "namespace", "hive",
		"")
	cmd.Flags().StringVar(&flags.BaseDomain, "basedomain", "coreostrain.me",
		"")
	cmd.Flags().StringVar(&flags.OpenShift, "openshift", "",
		"")
	cmd.Flags().StringVar(&flags.InstallConfig, "install-config", "ato-install-config",
		"")
	cmd.Flags().StringVar(&flags.ImagePullSecret, "image-pull-secret", "hive-install-config-global-pullsecret",
		"")
	cmd.Flags().StringVar(&flags.Platform, "platform", "",
		"")
	cmd.Flags().StringVar(&flags.Credentials, "credentials", "",
		"")
	cmd.Flags().StringVar(&flags.Region, "region", "",
		"")
	cmd.Flags().Int32Var(&flags.Running, "running", 0,
		"")
	cmd.Flags().Int32Var(&flags.Size, "size", 0,
		"")
	cmd.Flags().StringVar(&flags.IBMAccountID, "ibmaccountid", "",
		"")
	cmd.Flags().StringVar(&flags.IBMCISInstanceCRN, "ibmcisinstancecrn", "",
		"")

	return cmd
}

func setPlatform(platform string, flags orchestrate.PoolFlags) hivev1.Platform {
	switch platform {
	case "aws":
		aws := &aws.Platform{
			CredentialsSecretRef: corev1.LocalObjectReference{Name: flags.Credentials},
			Region:               flags.Region,
		}

		return hivev1.Platform{AWS: aws}
	case "azure":
		azure := &azure.Platform{
			CredentialsSecretRef:        corev1.LocalObjectReference{Name: flags.Credentials},
			Region:                      flags.Region,
			BaseDomainResourceGroupName: flags.AzureBaseDomainResourceGroupName,
			CloudName:                   flags.AzureCloudName,
		}

		return hivev1.Platform{Azure: azure}
	case "gcp":
		gcp := &gcp.Platform{
			CredentialsSecretRef: corev1.LocalObjectReference{Name: flags.Credentials},
			Region:               flags.Region,
		}

		return hivev1.Platform{GCP: gcp}
	case "ibm":
		ibm := &ibmcloud.Platform{
			CredentialsSecretRef: corev1.LocalObjectReference{Name: flags.Credentials},
			AccountID:            flags.IBMAccountID,
			CISInstanceCRN:       flags.IBMCISInstanceCRN,
			Region:               flags.Region,
		}

		return hivev1.Platform{IBMCloud: ibm}
	}

	return hivev1.Platform{
		AWS: &aws.Platform{
			CredentialsSecretRef: corev1.LocalObjectReference{Name: flags.Credentials},
			Region:               flags.Region,
		},
	}
}

func validation(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	hvclient := orchestrate.GetHiveClient()
	osversion := "ocp-" + orchestrate.GetOpenShiftVersions(flags)

	cp := hivev1.ClusterPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flags.Name,
			Namespace: flags.Namespace,
		},
		Spec: hivev1.ClusterPoolSpec{
			Platform:                       setPlatform(flags.Platform, flags),
			PullSecretRef:                  &corev1.LocalObjectReference{Name: flags.ImagePullSecret},
			Size:                           flags.Size,
			RunningCount:                   flags.Running,
			BaseDomain:                     flags.BaseDomain,
			ImageSetRef:                    hivev1.ClusterImageSetReference{Name: osversion},
			InstallConfigSecretTemplateRef: &corev1.LocalObjectReference{Name: flags.InstallConfig},
			SkipMachinePools:               true,
		},
	}

	if _, err := hvclient.HiveV1().ClusterPools(flags.Namespace).Create(ctx, &cp, metav1.CreateOptions{}); err != nil {
		log.Errorf("Unable to create ClusterPool: %v\n", err)
		return err
	}

	return nil
}
