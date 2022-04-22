package job

import (
	"audit-tool-orchestrator/pkg"
	"audit-tool-orchestrator/pkg/orchestrate"
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
)

var flags = orchestrate.JobFlags{}

var prioritizedInstallModes = []string{
	string(operatorv1alpha1.InstallModeTypeOwnNamespace),
	string(operatorv1alpha1.InstallModeTypeSingleNamespace),
	string(operatorv1alpha1.InstallModeTypeMultiNamespace),
	string(operatorv1alpha1.InstallModeTypeAllNamespaces),
}

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
	cmd.Flags().StringVar(&flags.PackageName, "package-name", "",
		"Bundle Package Name for Install Modes")
	cmd.Flags().BoolVar(&flags.InstallMode, "install-mode", false,
		"InstallMode Flag to run against provided bundle")
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

	targetNamespaces := []string{"default"}
	auditClient := orchestrate.K8sClientForAudit(kubeconfig)

	if flags.InstallMode {
		targetNamespaces, err = RunInstallMode()
		if err != nil {
			log.Errorf("Error running installModes: ", err)
		}
		log.Info("Creating namespace: ", targetNamespaces)
		nsName := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   targetNamespaces[0],
				Labels: map[string]string{"openshift.io/run-level": "0"},
			},
		}
		_, err = auditClient.CoreV1().Namespaces().Create(ctx, nsName, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Unable to create namespace: ", err)
		}
	}

	jobBackoffLimit := int32(1)
	jobPrivileged := true
	logEndpoint := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "MINIO_ENDPOINT",
		},
	}
	logAccessKeyID := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "MINIO_ACCESS_KEY_ID",
		},
	}
	logSecretAccessKey := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "env-var",
			},
			Key: "MINIO_SECRET_ACCESS_KEY",
		},
	}
	auditJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flags.Name,
			Namespace: targetNamespaces[0],
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
							Image: "quay.io/opdev/capabilities-tool:v1.0.1-test",
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
								"--namespace",
								targetNamespaces[0],
							},
							Env: []corev1.EnvVar{
								{Name: "MINIO_ENDPOINT", ValueFrom: logEndpoint},
								{Name: "MINIO_ACCESS_KEY_ID", ValueFrom: logAccessKeyID},
								{Name: "MINIO_SECRET_ACCESS_KEY", ValueFrom: logSecretAccessKey},
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

	job, err := auditClient.BatchV1().Jobs(targetNamespaces[0]).Create(ctx, &auditJob, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_ = orchestrate.WaitForAuditJob(auditClient, job)

	return nil
}

func RunInstallMode() ([]string, error) {
	log.Info("Pulling image: ", flags.BundleImage)

	options := make([]crane.Option, 0)
	img, err := crane.Pull(flags.BundleImage, options...)
	if err != nil {
		return nil, fmt.Errorf("Unable to pull image: %s\n", err)
	}

	os.Mkdir("tmp", 0o755)
	containerFSPath := path.Join("tmp", "bundle")
	if err := os.Mkdir(containerFSPath, 0o755); err != nil {
		return nil, fmt.Errorf("%s: %s", containerFSPath, err)
	}

	// export/flatten, and extract
	log.Info("exporting and flattening image")
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		log.Debugf("writing container filesystem to output dir: %s", containerFSPath)
		err = crane.Export(img, w)
		if err != nil {
			log.Error("unable to export and flatten container filesystem:", err)
		}
	}()

	log.Info("extracting container filesystem to ", containerFSPath)
	if err := pkg.Untar(containerFSPath, r); err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	installedModes, err := orchestrate.GetSupportedInstalledModes("tmp/bundle")

	var installMode string
	for i := 0; i < len(prioritizedInstallModes); i++ {
		if _, ok := installedModes[prioritizedInstallModes[i]]; ok {
			installMode = prioritizedInstallModes[i]
			break
		}
	}

	log.Debugf("The operator install mode is %s", installMode)
	targetNamespaces := make([]string, 2)

	switch installMode {
	case string(operatorv1alpha1.InstallModeTypeOwnNamespace):
		targetNamespaces = []string{flags.PackageName}
	case string(operatorv1alpha1.InstallModeTypeSingleNamespace):
		targetNamespaces = []string{flags.PackageName + "-target"}
	case string(operatorv1alpha1.InstallModeTypeMultiNamespace):
		targetNamespaces = []string{flags.PackageName, flags.PackageName + "-target"}
	case string(operatorv1alpha1.InstallModeTypeAllNamespaces):
		targetNamespaces = []string{}

	}

	return targetNamespaces, nil
}
