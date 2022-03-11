package pkg

type BundleFlags struct {
	IndexImage      string `json:"image"`
	OutputPath      string `json:"outputPath"`
	ContainerEngine string `json:"containerEngine"`
}

type CapabilitiesFlags struct {
}

// BindFlags define the flags used to generate the bundle report; inherited from audit-tool
type BindFlags struct {
	IndexImage        string `json:"image"`
	Limit             int32  `json:"limit"`
	HeadOnly          bool   `json:"headOnly"`
	DisableScorecard  bool   `json:"disableScorecard"`
	DisableValidators bool   `json:"disableValidators"`
	ServerMode        bool   `json:"serverMode"`
	Label             string `json:"label"`
	LabelValue        string `json:"labelValue"`
	Filter            string `json:"filter"`
	OutputPath        string `json:"outputPath"`
	OutputFormat      string `json:"outputFormat"`
	ContainerEngine   string `json:"containerEngine"`
}

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
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	PoolName  string `json:"poolName"`
}

type BundleList struct {
	Bundles []Bundle
}

type Bundle struct {
	Name           string   `json:"name"`
	PackageName    string   `json:"packageName"`
	DefaultChannel string   `json:"defaultChannel"`
	BundleImage    string   `json:"bundleImage"`
	Channels       []string `json:"channels"`
}
