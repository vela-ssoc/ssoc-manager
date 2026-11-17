package einotool

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/vela-ssoc/ssoc-common-mb/psutil"
	"github.com/vela-ssoc/ssoc-manager/agentic"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
)

type SystemProcess struct {
	svc *service.SystemProcess
	ftb agentic.BaseToolFactory
}

func NewSystemProcess(svc *service.SystemProcess, ftb agentic.BaseToolFactory) *SystemProcess {
	return &SystemProcess{svc: svc, ftb: ftb}
}

func (sp SystemProcess) NewEinoTool(cs agentic.ChatSession) ([]tool.BaseTool, error) {
	return sp.ftb.Builder().
		Add("system_process_list", "获取程序自身（管理端节点）所在机器的进程信息", sp.list).
		Build()
}

func (sp SystemProcess) list(ctx context.Context, _ request.Empty) (response.Records[*psutil.ProcessInfo], error) {
	return sp.svc.List(ctx)
}

func (sp SystemProcess) info(ctx context.Context, in request.SystemPID) (*psutil.ProcessInfo, error) {
	return sp.svc.Info(ctx, in.PID)
}
