package restapi

import (
	"net/http"

	psprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type SystemProcess struct {
	svc *service.SystemProcess
}

func NewSystemProcess(svc *service.SystemProcess) *SystemProcess {
	return &SystemProcess{svc: svc}
}

func (sp *SystemProcess) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/system/processes").GET(sp.list)                  // 进程列表
	g.ExposeV3.Group.Route("/system/process").GET(sp.info)                    // 进程信息
	g.ExposeV3.Group.Route("/system/process/connections").GET(sp.connections) // 进程打开的连接数
	g.ExposeV3.Group.Route("/system/process/openfiles").GET(sp.openfiles)     // 进程关联的文件描述符

	return nil
}

func (sp *SystemProcess) list(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := sp.svc.List(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (sp *SystemProcess) info(c *ship.Context) error {
	var param request.SystemPID
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := sp.svc.Info(ctx, param.PID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (sp *SystemProcess) connections(c *ship.Context) error {
	var param request.SystemPID
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	proc, err := psprocess.NewProcessWithContext(ctx, param.PID)
	if err != nil {
		return err
	}
	stats, err := proc.ConnectionsWithContext(ctx)
	if err != nil {
		return err
	}
	ret := response.NewRecords(stats)

	return c.JSON(http.StatusOK, ret)
}

func (sp *SystemProcess) openfiles(c *ship.Context) error {
	var param request.SystemPID
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	proc, err := psprocess.NewProcessWithContext(ctx, param.PID)
	if err != nil {
		return err
	}
	stats, err := proc.OpenFilesWithContext(ctx)
	if err != nil {
		return err
	}
	ret := response.NewRecords(stats)

	return c.JSON(http.StatusOK, ret)
}
