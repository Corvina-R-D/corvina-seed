package dto

type TriggerDTO struct {
	ChangeMask        string `json:"changeMask"`
	MinIntervalMs     int    `json:"minIntervalMs"`
	SkipFirstNChanges int    `json:"skipFirstNChanges"`
	Type              string `json:"type"`
}

type SendPolicyDTO struct {
	Triggers []TriggerDTO `json:"triggers"`
}

type HistoryPolicyDTO struct {
	Enabled bool `json:"enabled"`
}

type DatalinkDTO struct {
	Source string `json:"source"`
}

type IoTDataPropertiesDTO struct {
	Type          string            `json:"type"`
	Mode          *string           `json:"mode,omitempty"`
	HistoryPolicy *HistoryPolicyDTO `json:"historyPolicy,omitempty"`
	Datalink      *DatalinkDTO      `json:"datalink,omitempty"`
	SendPolicy    *SendPolicyDTO    `json:"sendPolicy,omitempty"`
	Version       *string           `json:"version,omitempty"`
}

type IoTDataDTO struct {
	UUID        *string                         `json:"UUID"`
	Type        string                          `json:"type"`
	InstanceOf  string                          `json:"instanceOf"`
	Properties  map[string]IoTDataPropertiesDTO `json:"properties"`
	Label       string                          `json:"label"`
	Unit        string                          `json:"unit"`
	Description string                          `json:"description"`
	Tags        []string                        `json:"tags"`
}
