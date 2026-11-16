package gops

import (
	"strings"
	"time"

	"github.com/JonasNowak/dms-gop/models"
	"github.com/shirou/gopsutil/v4/host"
)

func (self *GopsUtil) GetSystemHardware() (*models.SystemHardware, error) {
	info := &models.SystemHardware{}

	cpuTracker.mu.Lock()
	self.loadCPUStatic()
	info.CPU = models.CPUBasic{Count: cpuTracker.cpuCount, Model: cpuTracker.cpuModel}
	cpuTracker.mu.Unlock()

	biosInfo := getBIOSInfo()
	info.BIOS = biosInfo

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	info.Kernel = hostInfo.KernelVersion
	info.Hostname = hostInfo.Hostname
	info.Arch = hostInfo.KernelArch
	info.Distro = getDistroName()
	info.Init = "launchd"

	return info, nil
}

func (self *GopsUtil) GetGPUInfo() (*models.GPUInfo, error) {
	entries, err := self.detectGPUEntries()
	if err != nil {
		return nil, err
	}

	return &models.GPUInfo{GPUs: gpusFromEntries(entries)}, nil
}

// GetGPUInfoWithStats samples live metrics for all GPUs.
// The cursor from the previous call enables metrics derived from counter deltas.
func (self *GopsUtil) GetGPUInfoWithStats(cursor string) (*models.GPUInfo, error) {
	entries, err := self.detectGPUEntries()
	if err != nil {
		return nil, err
	}

	gpus := gpusFromEntries(entries)
	info := &models.GPUInfo{GPUs: gpus}

	samples, next := self.sampleGPUs(entries, decodeGPUCursor(cursor), time.Now())
	for i, sample := range samples {
		sample.applyTo(&gpus[i])
	}
	info.Cursor = encodeGPUCursor(next)
	return info, nil
}

type gpuEntry struct {
	Driver string
	Vendor string
	Model  string
}

func gpusFromEntries(entries []gpuEntry) []models.GPU {
	gpus := make([]models.GPU, 0, len(entries))
	for _, entry := range entries {
		gpus = append(gpus, models.GPU{
			Driver:      entry.Driver,
			Vendor:      entry.Vendor,
			DisplayName: entry.Model,
			FullName:    buildFullName(entry.Vendor, entry.Model),
		})
	}
	return gpus
}

func inferVendor(_ string, line string) string {
	lineLower := strings.ToLower(line)
	switch {
	case strings.Contains(lineLower, "apple"):
		return "Apple"
	case strings.Contains(lineLower, "amd"), strings.Contains(lineLower, "ati"):
		return "AMD"
	case strings.Contains(lineLower, "intel"):
		return "Intel"
	}

	return "Unknown"
}

func removeVendorPrefixes(name string) string {
	prefixes := []string{
		"Apple ",
		"Advanced Micro Devices, Inc. ",
		"AMD/ATI ",
		"AMD ",
		"ATI ",
		"Intel Corporation ",
		"Intel ",
	}

	result := name
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(result), strings.ToLower(prefix)) {
			result = result[len(prefix):]
			break
		}
	}

	return strings.TrimSpace(result)
}

func buildFullName(vendor, displayName string) string {
	if displayName == "Unknown" {
		return displayName
	}

	switch vendor {
	case "Apple", "AMD", "Intel":
		if strings.HasPrefix(strings.ToLower(displayName), strings.ToLower(vendor)+" ") {
			return displayName
		}
		return vendor + " " + displayName
	default:
		return displayName
	}
}
