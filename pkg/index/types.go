package index

type BundleFlags struct {
	IndexImage      string `json:"image"`
	OutputPath      string `json:"outputPath"`
	ContainerEngine string `json:"containerEngine"`
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
