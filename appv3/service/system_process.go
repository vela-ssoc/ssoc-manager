package service

import (
	"context"

	psprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/vela-ssoc/ssoc-common-mb/psutil"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
)

type SystemProcess struct {
}

func NewSystemProcess() *SystemProcess {
	return &SystemProcess{}
}

func (sp *SystemProcess) List(ctx context.Context) (response.Records[*psutil.ProcessInfo], error) {
	infos, err := psutil.ProcessesWithContext(ctx)
	if err != nil {
		return response.Records[*psutil.ProcessInfo]{}, err
	}
	ret := response.NewRecords(infos)

	return ret, nil
}

func (sp *SystemProcess) Info(ctx context.Context, pid int32) (*psutil.ProcessInfo, error) {
	proc, err := psprocess.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}

	return psutil.ProcessInfoLookup(ctx, proc)
}
