package sheet

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"gorm.io/gorm"
)

func MinionCSV(ctx context.Context, limit int, bom bool) CSVReader {
	return &minionCSVReader{
		ctx:   ctx,
		limit: limit,
		bom:   bom,
	}
}

type minionCSVReader struct {
	ctx     context.Context
	current int
	limit   int
	bom     bool
}

func (r *minionCSVReader) UTF8BOM() bool {
	return r.bom
}

func (r *minionCSVReader) Filename() string {
	at := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("minion-%s.csv", at)
}

func (r *minionCSVReader) Header() []string {
	return []string{
		"ID", "IPv4", "操作系统", "系统架构", "版本", "状态", "代理节点",
		"IDC", "部门", "业务类型", "备注", "运维负责人", "可登录帐号",
	}
}

func (r *minionCSVReader) Next() ([][]string, error) {
	offset := r.current * r.limit
	r.current++

	tbl := query.Minion
	mons, _, err := tbl.WithContext(r.ctx).FindByPage(offset, r.limit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, io.EOF
		}
		return nil, err
	}
	if len(mons) == 0 {
		return nil, io.EOF
	}

	records := make([][]string, 0, len(mons))
	for _, mon := range mons {
		id := strconv.FormatInt(mon.ID, 10)
		status := mon.Status.String()
		record := []string{
			id, mon.Inet, mon.Goos, mon.Arch, mon.Edition, status, mon.BrokerName,
			mon.IDC, mon.IBu, mon.Category, mon.Comment, mon.OpDuty, mon.Identity,
		}
		records = append(records, record)
	}

	return records, nil
}
