package serverd

import "github.com/vela-ssoc/ssoc-common/linkhub"

type BrokerNotifier interface {
	BrokerConnected(peer linkhub.Peer)
	BrokerDisconnected(brokerID int64)
}

type brokerNotifier struct{}

func (brokerNotifier) BrokerConnected(linkhub.Peer) {}

func (brokerNotifier) BrokerDisconnected(int64) {}
