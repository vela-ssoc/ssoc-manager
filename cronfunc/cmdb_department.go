package cronfunc

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (cj *CMDB) departmentAll(ctx context.Context, now time.Time) error {
	seq := cj.cli.DepartmentAll(ctx, 800)
	errcnt := new(errorContinuous)
	for resp, err := range seq {
		if err != nil {
			return err
		}

		err = cj.departmentUpsert(ctx, resp.Result, now)
		if errcnt.threshold(err, 10) {
			return err
		}
	}

	return nil
}

func (cj *CMDB) departmentUpsert(ctx context.Context, depts ywcmdb.Departments, now time.Time) error {
	mods := make([]mongo.WriteModel, 0, len(depts))
	for _, dept := range depts {
		data := cj.departmentConvert(dept)
		mod := cj.newCMDBBulkWriteModel(data, now)
		mods = append(mods, mod)
	}

	return cj.bulkWriteCMDB(ctx, mods)
}

func (cj *CMDB) departmentConvert(dept ywcmdb.Department) *model.CMDB {
	return &model.CMDB{
		CIID:                dept.ID,
		CIType:              dept.CIType,
		DepPath:             dept.DepPath,
		Department:          dept.DepPath,
		Name:                dept.Name,
		Owner:               cj.dutiesConvert(dept.Owner),
		Status:              dept.Status,
		BusinessType:        dept.BusinessType,
		ManagementType:      dept.ManagementType,
		ProjectID:           dept.ProjectID,
		UserDepartment:      dept.UserDepartment,
		User:                cj.dutiesConvert(dept.User),
		WorkflowApprover:    cj.dutiesConvert(dept.WorkflowApprover),
		TemporaryDepartment: dept.TemporaryDepartment,
	}
}
