package pool

// create ClusterPool resource

import (
	"audit-tool-orchestrator/pkg"
	"context"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pool",
		Short:   "",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	/*
		flags: name, namespace, basedomain, openshift, installconfig, platform, creds, region, running, size
	*/

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cp := hivev1.ClusterPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ato-cluster-pool",
			Namespace: "hive",
		},
		Spec: hivev1.ClusterPoolSpec{
			Platform: hivev1.Platform{
				AWS: &aws.Platform{
					Region:               "us-east-1",
					CredentialsSecretRef: corev1.LocalObjectReference{Name: "hive-aws-creds"},
				},
			},
			PullSecretRef:                  &corev1.LocalObjectReference{Name: "hive-install-config-global-pullsecret"},
			Size:                           3,
			RunningCount:                   1,
			BaseDomain:                     "coreostrain.me",
			ImageSetRef:                    hivev1.ClusterImageSetReference{Name: "ocp-4.9.23"},
			InstallConfigSecretTemplateRef: &corev1.LocalObjectReference{Name: "sno-install-config"},
			SkipMachinePools:               true,
		},
	}

	hvclient := pkg.GetHiveClient()
	if err := hvclient.Create(ctx, &cp); err != nil {
		log.Errorf("Unable to create ClusterPool: %v\n", err)
		return err
	}

	return nil
}
