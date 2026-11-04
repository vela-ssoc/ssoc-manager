package response

import (
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
	"github.com/vela-ssoc/ssoc-common-mb/pubsub"
)

type sseEvent struct {
	sse *eventsource.Writer
}

func NewSSEEvent(sse *eventsource.Writer) pubsub.Subscriber {
	return &sseEvent{sse: sse}
}

func (s *sseEvent) OnEvent(event string, data any) error {
	return s.sse.JSON(event, data)
}
