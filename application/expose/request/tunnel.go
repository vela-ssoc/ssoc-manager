package request

type TunnelOpen struct {
	Protocol string `json:"protocol" query:"protocol"`
}

func (t *TunnelOpen) ProtocolType() string {
	if t.Protocol == "yamux" {
		return t.Protocol
	}

	return "smux"
}
