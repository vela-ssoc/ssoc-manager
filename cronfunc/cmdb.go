package cronfunc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
)

func NewCMDB(pdb polydb.Sourcer, cli ywcmdb.Client, log *slog.Logger) *CMDB {
	return &CMDB{
		pdb: pdb,
		cli: cli,
		log: log,
	}
}

type CMDB struct {
	pdb polydb.Sourcer
	cli ywcmdb.Client
	log *slog.Logger
}

func (cj *CMDB) Options() crontab.JobOptions {
	return crontab.JobOptions{
		ID:   "3b8aaa9703c646394c8d8409bdd5fb7c",
		Name: "同步 CMDB 资产",
		Expr: "TZ=Asia/Shanghai 51 21 * * *",
		//Expr:    "TZ=Asia/Shanghai 0 1 * * *",
		Timeout: 2 * time.Hour,
	}
}

func (cj *CMDB) Execute(ctx context.Context) error {
	now := time.Now()
	var errs []error
	if err := cj.departmentAll(ctx, now); err != nil {
		errs = append(errs, err)
		cj.log.ErrorContext(ctx, "处理 CMDB 部门发生错误", "err", err)
	} else {
		cj.log.InfoContext(ctx, "CMDB 部门处理完毕")
	}

	if err := cj.serverAll(ctx, now); err != nil {
		errs = append(errs, err)
		cj.log.ErrorContext(ctx, "处理 CMDB 服务器发生错误", "err", err)
	} else {
		cj.log.InfoContext(ctx, "CMDB 服务器处理完毕")
	}

	return errors.Join(errs...)
}
