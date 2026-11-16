package gops

import "github.com/JonasNowak/dms-gop/models"

type cpuStatic struct {
	minMHz, maxMHz float64
	physicalCores  int
	baseMHz        float64
	cache          *models.CPUCache
}

type cpuDynamic struct {
	coreMHz []float64
}
