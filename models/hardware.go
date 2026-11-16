package models

type BIOSInfo struct {
	Vendor      string `json:"vendor"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Motherboard string `json:"motherboard"`
}

type SystemHardware struct {
	Kernel   string   `json:"kernel"`
	Distro   string   `json:"distro"`
	Hostname string   `json:"hostname"`
	Arch     string   `json:"arch"`
	CPU      CPUBasic `json:"cpu"`
	BIOS     BIOSInfo `json:"bios"`
	Init     string   `json:"initSystem,omitempty" doc:"Name of PID 1, e.g. launchd"`
}

type CPUBasic struct {
	Count int    `json:"count"`
	Model string `json:"model"`
}

type GPU struct {
	Driver      string  `json:"driver,omitempty"`
	Vendor      string  `json:"vendor"`
	DisplayName string  `json:"displayName"`
	FullName    string  `json:"fullName"`
	Temperature float64 `json:"temperature"`

	// Live metrics are only filled for requested GPUs; nil means the driver does
	// not expose the value. Units: percent, KB, watts, MHz.
	Usage          *float64 `json:"usage,omitempty"`
	MemoryUsed     *uint64  `json:"memoryUsed,omitempty"`
	MemoryTotal    *uint64  `json:"memoryTotal,omitempty"`
	Power          *float64 `json:"power,omitempty"`
	CoreClock      *float64 `json:"coreClock,omitempty"`
	MemoryClock    *float64 `json:"memoryClock,omitempty"`
	CoreClockMax   *float64 `json:"coreClockMax,omitempty" doc:"MHz"`
	MemoryClockMax *float64 `json:"memoryClockMax,omitempty" doc:"MHz"`
}

type GPUInfo struct {
	GPUs   []GPU  `json:"gpus"`
	Cursor string `json:"cursor,omitempty"`
}
