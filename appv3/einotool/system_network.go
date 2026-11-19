package einotool

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/vela-ssoc/ssoc-manager/agentic"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
)

type SystemNetwork struct {
	ftb agentic.BaseToolFactory
}

func NewSystemNetwork(ftb agentic.BaseToolFactory) *SystemNetwork {
	return &SystemNetwork{ftb: ftb}
}

func (sp SystemNetwork) NewEinoTool(cs agentic.ChatSession) ([]tool.BaseTool, error) {
	return sp.ftb.Builder().
		Add("system_network_interfaces", "获取管理端节点机器的网卡简要信息", sp.list).
		Add("system_network_connections", "获取管理端节点机器的网络连接列表信息", sp.connections).
		Build()
}

func (sp SystemNetwork) list(ctx context.Context, _ request.Empty) (*response.Records[psnet.InterfaceStat], error) {
	stats, err := psnet.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	ret := &response.Records[psnet.InterfaceStat]{Records: stats}

	return ret, nil
}

func (sp SystemNetwork) connections(ctx context.Context, in request.SystemNetworkConnection) (*response.Records[psnet.ConnectionStat], error) {
	stats, err := psnet.Connections(in.Kind)
	if err != nil {
		return nil, err
	}
	ret := &response.Records[psnet.ConnectionStat]{Records: stats}

	return ret, nil
}
