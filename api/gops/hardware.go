package gops_handler

import (
	"context"

	"github.com/JonasNowak/dms-gop/models"
	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-go/log"
)

type SystemHardwareResponse struct {
	Body *models.SystemHardware
}

type GPUResponse struct {
	Body *models.GPUInfo
}

// GET /hardware
func (self *HandlerGroup) SystemHardware(ctx context.Context, input *struct{}) (*SystemHardwareResponse, error) {
	systemInfo, err := self.srv.Gops.GetSystemHardware()
	if err != nil {
		log.Error("Error getting system hardware info")
		return nil, huma.Error500InternalServerError("Unable to retrieve system hardware info")
	}

	return &SystemHardwareResponse{Body: systemInfo}, nil
}

// GET /gpu
func (self *HandlerGroup) GPU(ctx context.Context, input *struct{}) (*GPUResponse, error) {
	gpuInfo, err := self.srv.Gops.GetGPUInfo()
	if err != nil {
		log.Error("Error getting GPU info")
		return nil, huma.Error500InternalServerError("Unable to retrieve GPU info")
	}

	return &GPUResponse{Body: gpuInfo}, nil
}
