package gops

import (
	"cmp"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/JonasNowak/dms-gop/models"
)

type gpuSample struct {
	Driver         string
	Temperature    float64
	Usage          *float64
	MemoryUsed     *uint64
	MemoryTotal    *uint64
	Power          *float64
	CoreClock      *float64
	MemoryClock    *float64
	CoreClockMax   *float64
	MemoryClockMax *float64
}

func (s gpuSample) applyTo(gpu *models.GPU) {
	gpu.Driver = cmp.Or(s.Driver, gpu.Driver)
	gpu.Temperature = s.Temperature
	gpu.Usage = s.Usage
	gpu.MemoryUsed = s.MemoryUsed
	gpu.MemoryTotal = s.MemoryTotal
	gpu.Power = s.Power
	gpu.CoreClock = s.CoreClock
	gpu.MemoryClock = s.MemoryClock
	gpu.CoreClockMax = s.CoreClockMax
	gpu.MemoryClockMax = s.MemoryClockMax
}

// gpuCursor carries monotonic counters between samples.
type gpuCursor struct {
	Timestamp int64                        `json:"timestamp"`
	Counters  map[string]map[string]uint64 `json:"counters,omitempty"`
}

func newGPUCursor(now time.Time) gpuCursor {
	return gpuCursor{Timestamp: now.UnixNano(), Counters: map[string]map[string]uint64{}}
}

func encodeGPUCursor(c gpuCursor) string {
	if len(c.Counters) == 0 {
		return ""
	}
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeGPUCursor(s string) gpuCursor {
	var c gpuCursor
	if s == "" {
		return c
	}
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return gpuCursor{}
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return gpuCursor{}
	}
	return c
}
