package gops

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInferVendor(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		line     string
		expected string
	}{
		{"apple gpu", "", "Apple M4", "Apple"},
		{"amd in line", "", "AMD Radeon RX 5700", "AMD"},
		{"ati in line", "", "ATI Radeon HD 5870", "AMD"},
		{"intel in line", "", "Intel UHD Graphics 630", "Intel"},
		{"unknown", "", "Some Random GPU", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferVendor(tt.driver, tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveVendorPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"AMD full", "Advanced Micro Devices, Inc. Radeon RX 5700", "Radeon RX 5700"},
		{"AMD/ATI", "AMD/ATI Radeon HD 7970", "Radeon HD 7970"},
		{"AMD only", "AMD Radeon RX 6800", "Radeon RX 6800"},
		{"ATI only", "ATI Radeon HD 5870", "Radeon HD 5870"},
		{"Intel Corporation", "Intel Corporation UHD Graphics 630", "UHD Graphics 630"},
		{"Intel only", "Intel Iris Xe Graphics", "Iris Xe Graphics"},
		{"Apple only", "Apple M4 Max", "M4 Max"},
		{"no prefix", "Radeon Pro Vega", "Radeon Pro Vega"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeVendorPrefixes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildFullName(t *testing.T) {
	tests := []struct {
		name        string
		vendor      string
		displayName string
		expected    string
	}{
		{"Apple Silicon GPU", "Apple", "Apple M4", "Apple M4"},
		{"AMD GPU", "AMD", "Radeon RX 5700", "AMD Radeon RX 5700"},
		{"Intel GPU", "Intel", "UHD Graphics 630", "Intel UHD Graphics 630"},
		{"Unknown vendor", "SomeVendor", "Some GPU", "Some GPU"},
		{"Unknown display", "Apple", "Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFullName(tt.vendor, tt.displayName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkRemoveVendorPrefixes(b *testing.B) {
	testString := "Intel Corporation UHD Graphics 630"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		removeVendorPrefixes(testString)
	}
}

func BenchmarkInferVendor(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inferVendor("", "Apple M4 Pro")
	}
}

func TestGetGPUInfoWithStats_SamplesAllGPUs(t *testing.T) {
	var sampled []string
	g := &GopsUtil{
		detectGPUEntries: func() ([]gpuEntry, error) {
			return []gpuEntry{
				{Driver: "apple_agx", Vendor: "Apple", Model: "Apple M4"},
			}, nil
		},
		sampleGPUs: func(entries []gpuEntry, _ gpuCursor, now time.Time) ([]gpuSample, gpuCursor) {
			samples := make([]gpuSample, len(entries))
			for i, e := range entries {
				sampled = append(sampled, e.Model)
				samples[i] = gpuSample{
					Temperature: 48.5,
					Usage:       ptr(12.0),
				}
			}
			cursor := newGPUCursor(now)
			cursor.Counters["apple_agx"] = map[string]uint64{"busy": 1}
			return samples, cursor
		},
	}

	info, err := g.GetGPUInfoWithStats("")
	require.NoError(t, err)
	assert.Equal(t, []string{"Apple M4"}, sampled)
	require.Len(t, info.GPUs, 1)
	assert.Equal(t, "Apple M4", info.GPUs[0].FullName)
	assert.Equal(t, 48.5, info.GPUs[0].Temperature)
	require.NotNil(t, info.GPUs[0].Usage)
	assert.Equal(t, 12.0, *info.GPUs[0].Usage)
	assert.NotEmpty(t, info.Cursor)
}
