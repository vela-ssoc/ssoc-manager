package ywcmdb

type Department struct {
	ID                  int64    `json:"_id,string"`
	CIType              string   `json:"ci_type"`
	DepPath             string   `json:"dep_path"`
	Name                string   `json:"name"`
	Owner               Duties   `json:"owner,omitzero"`
	Status              string   `json:"status,omitzero"`
	BusinessType        string   `json:"business_type,omitzero"`
	ManagementType      string   `json:"management_type,omitzero"`
	ProjectID           string   `json:"project_id,omitzero"`
	UserDepartment      []string `json:"user_department,omitzero"`
	User                Duties   `json:"user,omitzero"`
	WorkflowApprover    Duties   `json:"workflow_approver,omitzero"`
	TemporaryDepartment string   `json:"temporary_department,omitzero"`
}

type Departments []Department
