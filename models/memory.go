package models

type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	UsedPercent  float64 `json:"usedPercent"`
	Free         uint64  `json:"free"`
	Available    uint64  `json:"available"`
	Buffers      uint64  `json:"buffers"`
	Cached       uint64  `json:"cached"`
	SReclaimable uint64  `json:"sreclaimable"`
	Shared       uint64  `json:"shared"`
	GPUActive    uint64  `json:"gpuActive"`
	GPUReclaim   uint64  `json:"gpuReclaim"`
	SwapTotal    uint64  `json:"swaptotal"`
	SwapFree     uint64  `json:"swapfree"`

	// Apple Silicon unified-memory breakdown, in bytes.
	Wired          uint64 `json:"wired,omitempty"`
	Compressed     uint64 `json:"compressed,omitempty"`
	MemoryPressure string `json:"memoryPressure,omitempty" doc:"macOS memory pressure level: normal, warn or critical"`

	SwapDevices []SwapDevice `json:"swapDevices,omitempty"`
}

// SwapDevice is a macOS swap backing store.
type SwapDevice struct {
	Name string `json:"name"`
	Kind string `json:"kind" doc:"swap file (e.g. /private/var/vm/swapfile0), encrypted swap or partition"`
	Size uint64 `json:"size" doc:"KB"`
	Used uint64 `json:"used" doc:"KB"`
}
