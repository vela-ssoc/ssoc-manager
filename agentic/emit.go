package agentic

import (
	"context"
	"sync"

	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/tmc/a2ui"
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
)

type AGUIEmit struct {
	sse      *eventsource.Writer
	mutex    sync.Mutex
	baseline map[string][]a2ui.AgentMessage
}

func NewAGUIEmit(sse *eventsource.Writer) *AGUIEmit {
	return &AGUIEmit{
		sse:      sse,
		baseline: make(map[string][]a2ui.AgentMessage, 4),
	}
}

func (ae *AGUIEmit) Emit(_ context.Context, evt events.Event) error {
	return ae.sse.JSON("", evt)
}

func (ae *AGUIEmit) Error(ctx context.Context, msgID string, err error) error {
	return ae.Failed(ctx, msgID, err.Error())
}

func (ae *AGUIEmit) Failed(ctx context.Context, msgID, failedMsg string) error {
	nerr := ae.Emit(ctx, events.NewTextMessageEndEvent(msgID))
	if nerr == nil {
		_ = ae.Emit(ctx, events.NewRunErrorEvent(failedMsg))
	}

	return nerr
}

// A2UISnapshot 创建或整体替换一个 surface。
func (ae *AGUIEmit) A2UISnapshot(ctx context.Context, surfaceID string, msgs []a2ui.AgentMessage) error {
	messageID := ae.surfaceMessageID(surfaceID)
	content := map[string]any{"a2ui_operations": msgs}
	evt := events.NewActivitySnapshotEvent(messageID, "a2ui-surface", content)

	ae.mutex.Lock()
	defer ae.mutex.Unlock()

	if err := ae.Emit(ctx, evt); err != nil {
		return err
	}
	ae.baseline[surfaceID] = msgs

	return nil
}

// A2UIDelta 在已有 surface 上追加 A2UI 操作。
//
// 没有基线时返回 ErrSurfaceNotFound，由调用方决定是否提示模型先创建 surface。
// 不在这里自动退化成 snapshot，因为 updateDataModel / updateComponents
// 单独发 snapshot 会得到一个不存在的 surface。
func (ae *AGUIEmit) A2UIDelta(ctx context.Context, surfaceID string, msgs []a2ui.AgentMessage) error {
	ae.mutex.Lock()
	defer ae.mutex.Unlock()

	base := ae.baseline[surfaceID]
	if len(base) == 0 {
		// TODO 返回错误
		// return fmt.Errorf("%w: %s", ErrSurfaceNotFound, surfaceID)
		return nil
	}

	patch := make([]events.JSONPatchOperation, 0, len(msgs))
	for _, msg := range msgs {
		op := events.JSONPatchOperation{
			Op:    "add",
			Path:  "/a2ui_operations/-",
			Value: msg,
		}
		patch = append(patch, op)
	}

	messageID := ae.surfaceMessageID(surfaceID)
	evt := events.NewActivityDeltaEvent(messageID, "a2ui-surface", patch)
	if err := ae.Emit(ctx, evt); err != nil {
		return err
	}
	ae.baseline[surfaceID] = append(base, msgs...)

	return nil
}

// A2UIDelete 移除一个 surface。
// 当前运行创建过该 surface 时发 delta，否则发一次只包含 deleteSurface 的 snapshot。
func (ae *AGUIEmit) A2UIDelete(ctx context.Context, surfaceID string) error {
	op := a2ui.AgentMessage{
		Version: a2ui.Version,
		DeleteSurface: &a2ui.DeleteSurface{
			SurfaceID: surfaceID,
		},
	}
	messageID := ae.surfaceMessageID(surfaceID)

	ae.mutex.Lock()
	defer ae.mutex.Unlock()

	var evt events.Event
	base := ae.baseline[surfaceID]
	if len(base) == 0 {
		content := map[string]any{"a2ui_operations": []a2ui.AgentMessage{op}}
		evt = events.NewActivitySnapshotEvent(messageID, "a2ui-surface", content)
	} else {
		patch := []events.JSONPatchOperation{{
			Op:    "add",
			Path:  "/a2ui_operations/-",
			Value: op,
		}}

		evt = events.NewActivityDeltaEvent(messageID, "a2ui-surface", patch)
	}
	if err := ae.Emit(ctx, evt); err != nil {
		return err
	}
	delete(ae.baseline, surfaceID)

	return nil
}

func (*AGUIEmit) surfaceMessageID(surfaceID string) string {
	return "a2ui-surface-" + surfaceID
}
