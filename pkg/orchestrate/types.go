package orchestrate

type PoolFlags struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	BaseDomain    string `json:"baseDomain"`
	OpenShift     string `json:"openshift"`
	InstallConfig string `json:"installConfig"`
	Platform      string `json:"platform"`
	Credentials   string `json:"credentials"`
	Region        string `json:"region"`
	Running       int    `json:"running"`
	Size          int    `json:"size"`
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
