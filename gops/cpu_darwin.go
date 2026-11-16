//go:build darwin

package gops

import (
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/sensors"
)

// getCPUTemperatureCached returns the average CPU temperature from Apple Silicon sensors.
// It filters for valid die temperatures (10-125°C) matching "tdie", "cpu", or "soc" in the sensor key.
func getCPUTemperatureCached() float64 {
	sensorsTemps, err := sensors.SensorsTemperatures()
	if err != nil {
		return 0
	}

	var validTemps []float64
	for _, temp := range sensorsTemps {
		// Filter for reasonable temperature range and Apple Silicon die sensors
		if temp.Temperature > 10 && temp.Temperature < 125 {
			switch {
			case containsSensorKey(temp.SensorKey, "tdie", "cpu", "soc"):
				validTemps = append(validTemps, temp.Temperature)
			}
		}
	}

	// If we have valid die temps, average them
	if len(validTemps) > 0 {
		sum := float64(0)
		for _, t := range validTemps {
			sum += t
		}
		return sum / float64(len(validTemps))
	}

	// Fallback: use any valid temperature in the 10-125 range
	for _, temp := range sensorsTemps {
		if temp.Temperature > 10 && temp.Temperature < 125 {
			validTemps = append(validTemps, temp.Temperature)
		}
	}
	if len(validTemps) > 0 {
		sum := float64(0)
		for _, t := range validTemps {
			sum += t
		}
		return sum / float64(len(validTemps))
	}

	return 0
}

func containsSensorKey(s string, substrings ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func getCurrentCPUFreq() float64 {
	return 0
}

// cpuUsageFromProvider uses gopsutil's cpu.Percent on macOS.
// The custom tick-ratio calculation is unreliable on Apple Silicon because
// host_processor_info may not account for parked efficiency cores correctly,
// inflating the busy/total ratio. gopsutil's Percent(0) uses its own internal
// delta cache and is tested on macOS.
// For mocked CPUInfoProvider (not *DefaultCPUInfoProvider), fall back to per-core calculation.
func cpuUsageFromProvider(cpuProvider CPUInfoProvider, cursorTotal, currentTotal []float64, timeDiff float64, numCPUs int) (float64, []float64) {
	if _, isDefault := cpuProvider.(*DefaultCPUInfoProvider); !isDefault {
		return calculateCPUPercentage(cursorTotal, currentTotal), nil
	}

	totalPercent := 0.0
	var corePercents []float64

	// Percent(0) returns delta from last Percent() call.
	// First call after the initial Percent(100ms) seed will have a good baseline.
	total, err := cpu.Percent(0, false)
	if err == nil && len(total) > 0 {
		totalPercent = total[0]
	}

	perCore, err := cpu.Percent(0, true)
	if err == nil {
		corePercents = perCore
	}

	return totalPercent, corePercents
}

// Prime gopsutil's internal CPU cache so subsequent Percent(0) calls work.
func primeCPUPercent() {
	_, _ = cpu.Percent(200*time.Millisecond, false)
	_, _ = cpu.Percent(200*time.Millisecond, true)
}
