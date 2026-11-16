package models

type Sensor struct {
	Chip     string   `json:"chip" doc:"Sensor source, e.g. an IOHID/SMC device name."`
	Device   string   `json:"device" doc:"Basename of the sensor device; empty if none."`
	Label    string   `json:"label"`
	Kind     string   `json:"kind" doc:"temperature, fan or power."`
	Value    float64  `json:"value" doc:"Degrees Celsius, rpm or watts depending on kind."`
	Max      *float64 `json:"max,omitempty"`
	Critical *float64 `json:"critical,omitempty"`
}

type SensorsInfo struct {
	Sensors []Sensor `json:"sensors"`
}
