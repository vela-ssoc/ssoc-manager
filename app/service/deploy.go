package service

import "context"

type DeployService interface {
	LAN(ctx context.Context) string
}

func Deploy(store StoreService) DeployService {
	return &deployService{
		store: store,
	}
}

type deployService struct {
	store StoreService
}

func (biz *deployService) LAN(ctx context.Context) string {
	const key = "global.local.addr"
	if st, _ := biz.store.FindID(ctx, key); st != nil {
		return string(st.Value)
	}

	return ""
}
