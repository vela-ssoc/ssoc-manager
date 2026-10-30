package ywcmdb

type PrivateDNSRecord struct {
	ID          int64  `json:"_id,string"`
	CIType      string `json:"ci_type"`
	CreatedTime string `json:"created_time,omitzero"`
	Department  string `json:"department"`
	DomainName  string `json:"domain_name"`
	Duty        Duties `json:"duty"`
	RecordType  string `json:"record_type"`
	RecordValue string `json:"record_value"`
	TTL         int    `json:"ttl"`
	UpdateTime  string `json:"update_time"`
}

type PrivateDNSRecords []PrivateDNSRecord

type PublicDNSRecord struct {
	ID               int64    `json:"_id,string"`
	AccountID        string   `json:"account_id"`
	CIType           string   `json:"ci_type"`
	CreatedTime      string   `json:"created_time"`
	Department       string   `json:"department"`
	Duty             Duties   `json:"duty,omitzero"`
	ISP              string   `json:"isp"`
	PublicDomainName string   `json:"public_domain_name"`
	RecordType       string   `json:"record_type"`
	RecordValue      string   `json:"record_value"`
	Status           string   `json:"status"`
	UpdateTime       string   `json:"update_time"`
	DeviceName       string   `json:"device_name,omitzero"`
	IDC              string   `json:"idc,omitzero"`
	PrivateIP        []string `json:"private_ip,omitzero"`
	ServerRoom       string   `json:"server_room,omitzero"`
}

type PublicDNSRecords []PublicDNSRecord
