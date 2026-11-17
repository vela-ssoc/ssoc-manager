package agentic

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tmc/a2ui"
	"github.com/tmc/a2ui/a2uistate"
)

var (
	// ErrA2UISurfaceNotFound 表示当前线程中不存在目标 surface。
	ErrA2UISurfaceNotFound = errors.New("a2ui: surface not found")
	// ErrA2UINilEmitter 表示没有可用的 AG-UI 出口。
	ErrA2UINilEmitter = errors.New("a2ui: nil emitter")
)

// A2UISurfaceStore 持久化线程下的 surface 操作序列。
//
// Load 返回 thread 下所有 surface 的完整操作序列；
// Save 保存单个 surface 的完整操作序列；
// Delete 删除单个 surface。
type A2UISurfaceStore interface {
	Load(ctx context.Context, threadID string) (map[string][]a2ui.AgentMessage, error)
	Save(ctx context.Context, threadID, surfaceID string, ops []a2ui.AgentMessage) error
	Delete(ctx context.Context, threadID, surfaceID string) error
}

// A2UIStates 管理全局的 ThreadState。
//
// 一个进程/服务通常只需要一个 A2UIStates 实例：
//
//	states := agentic.NewA2UIStates(store)
//
// ThreadState 只保存 surface 状态，不持有 emitter。
// 每次 run 通过 Run 获取一个 RunState，由 RunState 绑定自己的 AGUIEmit。
type A2UIStates struct {
	mu      sync.Mutex
	store   A2UISurfaceStore
	threads map[string]*A2UIThreadState
}

// NewA2UIStates 创建全局 ThreadState 管理器。
func NewA2UIStates(store A2UISurfaceStore) *A2UIStates {
	return &A2UIStates{
		store:   store,
		threads: make(map[string]*A2UIThreadState),
	}
}

// Thread 获取或创建线程状态。
//
// ThreadState 不绑定 emitter，因此可以被多个 run 共享。
func (ss *A2UIStates) Thread(ctx context.Context, threadID string) (*A2UIThreadState, error) {
	if threadID == "" {
		return nil, errors.New("a2ui: empty thread id")
	}

	ss.mu.Lock()
	ts, exists := ss.threads[threadID]
	ss.mu.Unlock()
	if exists {
		return ts, nil
	}

	ts = &A2UIThreadState{
		threadID: threadID,
		store:    ss.store,
		ops:      make(map[string][]a2ui.AgentMessage),
	}
	if ss.store != nil {
		stored, err := ss.store.Load(ctx, threadID)
		if err != nil {
			return nil, fmt.Errorf("a2ui: load thread %s: %w", threadID, err)
		}
		for surfaceID, ops := range stored {
			if err = validateOpsSurface(surfaceID, ops); err != nil {
				return nil, fmt.Errorf("a2ui: restore surface %s: %w", surfaceID, err)
			}
			if err = ts.surfaces.Apply(ops...); err != nil {
				return nil, fmt.Errorf("a2ui: restore surface %s: %w", surfaceID, err)
			}
			ts.ops[surfaceID] = cloneA2UIOps(ops)
		}
	}

	ss.mu.Lock()
	if existing, ok := ss.threads[threadID]; ok {
		ss.mu.Unlock()
		return existing, nil
	}
	ss.threads[threadID] = ts
	ss.mu.Unlock()

	return ts, nil
}

// Run 为一次 run 创建独立的 RunState。
//
// 每个 run 必须使用自己的 RunState，保证 surface 事件只发到自己的 AGUIEmit。
func (ss *A2UIStates) Run(
	ctx context.Context,
	threadID string,
	emit *AGUIEmit,
) (*A2UIRunState, error) {
	if emit == nil {
		return nil, ErrA2UINilEmitter
	}

	ts, err := ss.Thread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	return &A2UIRunState{
		thread: ts,
		emit:   emit,
		seeded: make(map[string]bool),
	}, nil
}

// Get 返回已经加载的线程状态。
func (ss *A2UIStates) Get(threadID string) (*A2UIThreadState, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ts, exists := ss.threads[threadID]
	return ts, exists
}

// Remove 从全局管理器中移除线程状态。
//
// 只移除内存缓存，不删除 Store 中的 surface；
// 已经创建出来的 RunState 仍然持有原 ThreadState 指针。
func (ss *A2UIStates) Remove(threadID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.threads, threadID)
}

// ThreadIDs 返回当前已加载的线程 ID。
func (ss *A2UIStates) ThreadIDs() []string {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	out := make([]string, 0, len(ss.threads))
	for threadID := range ss.threads {
		out = append(out, threadID)
	}
	return out
}

// A2UIThreadState 是单个聊天线程的 surface 状态。
//
// 它只保存 surfaces/ops，不持有 emitter，也不发送事件。
type A2UIThreadState struct {
	mu       sync.Mutex
	threadID string
	store    A2UISurfaceStore
	surfaces a2uistate.Surfaces
	ops      map[string][]a2ui.AgentMessage
}

// ThreadID 返回线程 ID。
func (ts *A2UIThreadState) ThreadID() string {
	return ts.threadID
}

// HasSurface 报告当前线程是否存在该 surface。
func (ts *A2UIThreadState) HasSurface(surfaceID string) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	_, exists := ts.surfaces.Surface(surfaceID)
	return exists
}

// Surface 返回内部 surface，只读使用。
//
// 不要在锁外调用返回值的 Apply；修改状态请使用 RunState。
func (ts *A2UIThreadState) Surface(surfaceID string) (*a2uistate.Surface, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.surfaces.Surface(surfaceID)
}

// Ops 返回某个 surface 的完整操作序列副本。
func (ts *A2UIThreadState) Ops(surfaceID string) ([]a2ui.AgentMessage, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ops, exists := ts.ops[surfaceID]
	return cloneA2UIOps(ops), exists
}

// SurfaceIDs 返回当前线程所有未删除的 surface ID。
func (ts *A2UIThreadState) SurfaceIDs() []string {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	out := make([]string, 0, len(ts.ops))
	for surfaceID := range ts.ops {
		out = append(out, surfaceID)
	}
	return out
}

// A2UISurfaceInfo 是给模型/审计使用的 surface 摘要。
type A2UISurfaceInfo struct {
	SurfaceID string `json:"surface_id"`
	CatalogID string `json:"catalog_id,omitempty"`
	Deleted   bool   `json:"deleted,omitempty"`
}

// Infos 返回当前线程的 surface 摘要列表。
func (ts *A2UIThreadState) Infos() []A2UISurfaceInfo {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	out := make([]A2UISurfaceInfo, 0, len(ts.ops))
	for surfaceID := range ts.ops {
		surf, exists := ts.surfaces.Surface(surfaceID)
		if !exists {
			continue
		}
		out = append(out, A2UISurfaceInfo{
			SurfaceID: surfaceID,
			CatalogID: surf.CatalogID(),
			Deleted:   surf.Deleted(),
		})
	}
	return out
}

// A2UIRunState 是一次 run 的 surface 视图。
//
// 它持有自己的 AGUIEmit 和 seeded 标记，保证：
//   - A 的 surface 只推送到 A 的 SSE 通道；
//   - B 的 baseline 不会影响 A 的 delta 判断。
type A2UIRunState struct {
	mu     sync.Mutex
	thread *A2UIThreadState
	emit   *AGUIEmit
	seeded map[string]bool
}

// Thread 返回底层 ThreadState。
func (rs *A2UIRunState) Thread() *A2UIThreadState {
	return rs.thread
}

// Snapshot 创建或整体替换一个 surface，只发送到当前 run 的 emitter。
func (rs *A2UIRunState) Snapshot(ctx context.Context, surfaceID string, ops ...a2ui.AgentMessage) error {
	if surfaceID == "" {
		return errors.New("a2ui: empty surface id")
	}
	if len(ops) == 0 {
		return errors.New("a2ui: empty operations")
	}
	if err := validateOpsSurface(surfaceID, ops); err != nil {
		return err
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.thread.mu.Lock()
	defer rs.thread.mu.Unlock()

	if _, exists := rs.thread.surfaces.Surface(surfaceID); exists {
		_ = rs.thread.surfaces.Apply(a2uiDeleteMessage(surfaceID))
	}
	if err := rs.thread.surfaces.Apply(ops...); err != nil {
		return err
	}

	cloned := cloneA2UIOps(ops)
	rs.thread.ops[surfaceID] = cloned
	rs.seeded[surfaceID] = true

	if rs.thread.store != nil {
		if err := rs.thread.store.Save(ctx, rs.thread.threadID, surfaceID, cloned); err != nil {
			return fmt.Errorf("a2ui: save surface %s: %w", surfaceID, err)
		}
	}

	return rs.emit.A2UISnapshot(ctx, surfaceID, cloned)
}

// Delta 在已有 surface 上追加操作，只发送到当前 run 的 emitter。
//
// 当前 run 第一次更新历史 surface 时，会先向当前 emitter 发送完整 snapshot，
// 建立当前 run 的 baseline，再发送 delta。
func (rs *A2UIRunState) Delta(ctx context.Context, surfaceID string, ops ...a2ui.AgentMessage) error {
	if surfaceID == "" {
		return errors.New("a2ui: empty surface id")
	}
	if len(ops) == 0 {
		return errors.New("a2ui: empty operations")
	}
	if err := validateOpsSurface(surfaceID, ops); err != nil {
		return err
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.thread.mu.Lock()
	defer rs.thread.mu.Unlock()

	if _, exists := rs.thread.surfaces.Surface(surfaceID); !exists {
		return fmt.Errorf("%w: %s", ErrA2UISurfaceNotFound, surfaceID)
	}

	base := rs.thread.ops[surfaceID]
	if len(base) == 0 {
		return fmt.Errorf("%w: %s", ErrA2UISurfaceNotFound, surfaceID)
	}

	if !rs.seeded[surfaceID] {
		if err := rs.emit.A2UISnapshot(ctx, surfaceID, base); err != nil {
			return err
		}
		rs.seeded[surfaceID] = true
	}

	if err := rs.thread.surfaces.Apply(ops...); err != nil {
		return err
	}

	rs.thread.ops[surfaceID] = append(base, cloneA2UIOps(ops)...)
	if rs.thread.store != nil {
		if err := rs.thread.store.Save(ctx, rs.thread.threadID, surfaceID, rs.thread.ops[surfaceID]); err != nil {
			return fmt.Errorf("a2ui: save surface %s: %w", surfaceID, err)
		}
	}

	return rs.emit.A2UIDelta(ctx, surfaceID, ops)
}

// Delete 删除一个 surface，只发送到当前 run 的 emitter。
func (rs *A2UIRunState) Delete(ctx context.Context, surfaceID string) error {
	if surfaceID == "" {
		return errors.New("a2ui: empty surface id")
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.thread.mu.Lock()
	defer rs.thread.mu.Unlock()

	if _, exists := rs.thread.surfaces.Surface(surfaceID); !exists {
		return fmt.Errorf("%w: %s", ErrA2UISurfaceNotFound, surfaceID)
	}

	if err := rs.thread.surfaces.Apply(a2uiDeleteMessage(surfaceID)); err != nil {
		return err
	}

	delete(rs.thread.ops, surfaceID)
	delete(rs.seeded, surfaceID)

	if rs.thread.store != nil {
		if err := rs.thread.store.Delete(ctx, rs.thread.threadID, surfaceID); err != nil {
			return fmt.Errorf("a2ui: delete surface %s: %w", surfaceID, err)
		}
	}

	return rs.emit.A2UIDelete(ctx, surfaceID)
}

// MemoryA2UISurfaceStore 是 A2UISurfaceStore 的内存实现，适合开发和测试。
type MemoryA2UISurfaceStore struct {
	mu   sync.Mutex
	data map[string]map[string][]a2ui.AgentMessage
}

func NewMemoryA2UISurfaceStore() *MemoryA2UISurfaceStore {
	return &MemoryA2UISurfaceStore{
		data: make(map[string]map[string][]a2ui.AgentMessage),
	}
}

func (ms *MemoryA2UISurfaceStore) Load(_ context.Context, threadID string) (map[string][]a2ui.AgentMessage, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	surfaces := ms.data[threadID]
	out := make(map[string][]a2ui.AgentMessage, len(surfaces))
	for surfaceID, ops := range surfaces {
		out[surfaceID] = cloneA2UIOps(ops)
	}
	return out, nil
}

func (ms *MemoryA2UISurfaceStore) Save(_ context.Context, threadID, surfaceID string, ops []a2ui.AgentMessage) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.data == nil {
		ms.data = make(map[string]map[string][]a2ui.AgentMessage)
	}
	surfaces := ms.data[threadID]
	if surfaces == nil {
		surfaces = make(map[string][]a2ui.AgentMessage)
		ms.data[threadID] = surfaces
	}
	surfaces[surfaceID] = cloneA2UIOps(ops)
	return nil
}

func (ms *MemoryA2UISurfaceStore) Delete(_ context.Context, threadID, surfaceID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	surfaces := ms.data[threadID]
	delete(surfaces, surfaceID)
	if len(surfaces) == 0 {
		delete(ms.data, threadID)
	}
	return nil
}

func a2uiDeleteMessage(surfaceID string) a2ui.AgentMessage {
	return a2ui.AgentMessage{
		Version: a2ui.Version,
		DeleteSurface: &a2ui.DeleteSurface{
			SurfaceID: surfaceID,
		},
	}
}

func validateOpsSurface(surfaceID string, ops []a2ui.AgentMessage) error {
	for i, op := range ops {
		id := a2uiSurfaceID(op)
		if id != "" && id != surfaceID {
			return fmt.Errorf("operation %d targets surface %q, want %q", i, id, surfaceID)
		}
	}
	return nil
}

func a2uiSurfaceID(op a2ui.AgentMessage) string {
	switch {
	case op.CreateSurface != nil:
		return op.CreateSurface.SurfaceID
	case op.UpdateComponents != nil:
		return op.UpdateComponents.SurfaceID
	case op.UpdateDataModel != nil:
		return op.UpdateDataModel.SurfaceID
	case op.DeleteSurface != nil:
		return op.DeleteSurface.SurfaceID
	default:
		return ""
	}
}

func cloneA2UIOps(ops []a2ui.AgentMessage) []a2ui.AgentMessage {
	if len(ops) == 0 {
		return nil
	}
	return append([]a2ui.AgentMessage(nil), ops...)
}
