package job

import (
	"audit-tool-orchestrator/pkg/orchestrate"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

var flags = orchestrate.JobFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Short:   "Create a K8s Job resource.",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	cmd.Flags().StringVar(&flags.Name, "name", "ato-cluster-claim",
		"Name for the Job resource.")
	cmd.Flags().StringVar(&flags.BundleImage, "bundle-image", "",
		"Bundle this job will run operator-sdk run bundle against.")
	cmd.Flags().StringVar(&flags.BundleName, "bundle-name", "",
		"Bundle this job will run operator-sdk run bundle against.")
	cmd.Flags().StringVar(&flags.BucketName, "bucket-name", "",
		"S3 (minio) compatible bucket to store logs.")
	cmd.Flags().StringVar(&flags.ClaimName, "claim-name", "",
		"ClusterClaim resource to be used for audit job.")
	cmd.Flags().StringVar(&flags.Kubeconfig, "kubeconfig", "",
		"Kubeconfig to use for creating Job resource.")

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	kubeconfig, err := os.ReadFile(flags.Kubeconfig)
	if err != nil {
		log.Fatalf("Kubeconfig required to create Job resource: %v\n", err)
	}

	auditClient := orchestrate.K8sClientForAudit(kubeconfig)

	jobBackoffLimit := int32(1)
	jobPrivileged := true
	logEndpoint := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "ENDPOINT",
		},
	}
	logAccessKeyID := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "ACCESS_KEY_ID",
		},
	}
	logSecretAccessKey := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "SECRET_ACCESS_KEY",
		},
	}
	auditJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flags.Name,
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &jobBackoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "audit-tool-job-pod",
					Labels: map[string]string{"operator": flags.Name},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "docker-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "registry-pull-secret",
									Items: []corev1.KeyToPath{
										{Key: ".dockerconfigjson", Path: "config.json"},
									},
								},
							},
						},
						{
							Name: "kube-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "kubeconfig",
									Items: []corev1.KeyToPath{
										{Key: "config", Path: "config"},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "audit-tool",
							Image: "quay.io/opdev/capabilities-tool:v0.2.6",
							Args: []string{
								"index",
								"capabilities",
								"--container-engine",
								"podman",
								"--output-path",
								"/opt/capabilities-tool",
								"--bundle-image",
								flags.BundleImage,
								"--bucket-name",
								flags.BucketName,
								"--bundle-name",
								flags.BundleName,
							},
							Env: []corev1.EnvVar{
								{Name: "ENDPOINT", ValueFrom: logEndpoint},
								{Name: "ACCESS_KEY_ID", ValueFrom: logAccessKeyID},
								{Name: "SECRET_ACCESS_KEY", ValueFrom: logSecretAccessKey},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "docker-config", MountPath: "/opt/capabilities-tool/.docker/"},
								{Name: "kube-config", MountPath: "/opt/capabilities-tool/.kube/"},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &jobPrivileged,
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	job, err := auditClient.BatchV1().Jobs("default").Create(ctx, &auditJob, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_ = orchestrate.WaitForAuditJob(auditClient, job)

	return nil
}
