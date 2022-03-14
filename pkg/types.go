package pkg

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
