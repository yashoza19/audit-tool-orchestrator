package orchestrate

import "github.com/openshift/hive/apis/hive/v1/azure"

type PoolFlags struct {
	Name                             string                 `json:"name"`
	Namespace                        string                 `json:"namespace"`
	BaseDomain                       string                 `json:"baseDomain"`
	OpenShift                        string                 `json:"openshift"`
	InstallConfig                    string                 `json:"installConfig"`
	ImagePullSecret                  string                 `json:"image-pull-secret"`
	Platform                         string                 `json:"platform"`
	Credentials                      string                 `json:"credentials"`
	Region                           string                 `json:"region"`
	Running                          int32                  `json:"running"`
	Size                             int32                  `json:"size"`
	AzureBaseDomainResourceGroupName string                 `json:"azurebasedomainresourcegroupname"`
	AzureCloudName                   azure.CloudEnvironment `json:"azurecloudname"`
	IBMAccountID                     string                 `json:"ibmaccountid"`
	IBMCISInstanceCRN                string                 `json:"ibmcisinstancecrn"`
}

type ClaimFlags struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	PoolName   string `json:"poolName"`
	BundleName string `json:"bundleName"`
	Delete     bool   `json:"delete"`
}

type ClusterClaimDeleteFlagSetNameFlagEmptyError struct{}
type ClusterClaimNameLengthIncorrectError struct{}
type ClusterClaimNameHasInvalidCharactersError struct{}

type JobFlags struct {
	Name        string `json:"name"`
	BundleImage string `json:"bundleImage"`
	BundleName  string `json:"bundleName"`
	BucketName  string `json:"bucket-name"`
	ClaimName   string `json:"claim-name"`
	Kubeconfig  string `json:"kubeconfig"`
	PackageName string `json:"package-name"`
	InstallMode bool   `json:"installMode"`
}

type ClusterServiceVersion struct {
	Spec ClusterServiceVersionSpec `yaml:"spec"`
}

type ClusterServiceVersionSpec struct {
	// InstallModes specify supported installation types
	InstallModes []InstallMode `yaml:"installModes,omitempty"`
}

type InstallMode struct {
	Type      string `yaml:"type"`
	Supported bool   `yaml:"supported"`
}
