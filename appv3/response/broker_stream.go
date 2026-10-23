package response

import "time"

type BrokerSubstream struct {
	ID            uint64    `json:"id,string"`
	RX            uint64    `json:"rx"`
	TX            uint64    `json:"tx"`
	EstablishedAt time.Time `json:"established_at"`
}
