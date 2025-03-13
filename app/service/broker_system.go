package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
)

func NewBrokerSystem(hub linkhub.Huber) *BrokerSystem {
	return &BrokerSystem{
		hub: hub,
	}
}

type BrokerSystem struct {
	hub linkhub.Huber
}

// Exit 让 broker 退出执行。
// 重启 broker 时可以用到，一般 broker 退出后，systemctl 会拉起 broker 程序，
// 从而实现重启的效果。
func (bc *BrokerSystem) Exit(ctx context.Context, id int64) error {
	_ = bc.hub.Oneway(ctx, id, "/api/v1/system/exit", nil)
	return nil
}

// Update 通知 broker 更新
func (bc *BrokerSystem) Update(ctx context.Context, id int64, semver model.Semver) error {
	req := struct {
		Semver model.Semver `json:"semver"`
	}{
		Semver: semver,
	}
	_ = bc.hub.Oneway(ctx, id, "/api/v1/system/update", req)
	return nil
}
