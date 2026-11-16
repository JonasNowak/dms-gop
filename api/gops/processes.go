package gops_handler

import (
	"context"

	"github.com/JonasNowak/dms-gop/gops"
	"github.com/JonasNowak/dms-gop/models"
	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-go/log"
)

type ProcessInput struct {
	SortBy         gops.ProcSortBy `query:"sort_by" required:"true" default:"cpu"`
	Limit          int             `query:"limit"`
	DisableProcCPU bool            `query:"disable_proc_cpu" default:"false"`
	Cursor         string          `query:"cursor" required:"false"`
	MergeChildren  bool            `query:"merge_children" default:"true"`
	IO             bool            `query:"io" default:"false"`
	Net            bool            `query:"net" default:"false"`
	Members        bool            `query:"members" default:"false" doc:"With merge_children, list each merged row's processes in children"`
}

type ProcessResponse struct {
	Body struct {
		Data   []*models.ProcessInfo `json:"data"`
		Cursor string                `json:"cursor,omitempty"`
	}
}

// GET /processes
func (self *HandlerGroup) Processes(ctx context.Context, input *ProcessInput) (*ProcessResponse, error) {
	enableCPU := !input.DisableProcCPU

	result, err := self.srv.Gops.GetProcessesWithOptions(gops.ProcessOptions{
		SortBy: input.SortBy, Limit: input.Limit, EnableCPU: enableCPU, Cursor: input.Cursor,
		MergeChildren: input.MergeChildren, IO: input.IO, Net: input.Net, Members: input.Members,
	})
	if err != nil {
		log.Error("Error getting process info")
		return nil, huma.Error500InternalServerError("Unable to retrieve process info")
	}

	resp := &ProcessResponse{}
	resp.Body.Data = result.Processes
	resp.Body.Cursor = result.Cursor
	return resp, nil
}
