package apiclient

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/httpclient"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
)

type ClientAgentTask struct {
	hub muxproto.Huber
	cli httpclient.Client
	log *slog.Logger
}

func (ct ClientAgentTask) Resync(ctx context.Context, minionID int64) error {
	host := muxproto.Int64ToHost(minionID)
	rawURL := muxproto.BuildAgentURL(host, "/api/v3/agent-task/resync")
	strURL := rawURL.String()
	body := &IDParams{ID: []int64{minionID}}

	return ct.cli.PostJSON(ctx, strURL, body, new(empty))
}

func (ct ClientAgentTask) ResyncTags(ctx context.Context, tags []string) error {
	return nil
}

func (ct ClientAgentTask) Reload(ctx context.Context, minionID, substanceID int64) error {
	host := muxproto.Int64ToHost(minionID)
	rawURL := muxproto.BuildAgentURL(host, "/api/v3/agent-task/reload")
	strURL := rawURL.String()
	body := &AgentTaskReloadParams{ID: minionID, SubstanceID: substanceID}

	return ct.cli.PostJSON(ctx, strURL, body, new(empty))
}

type AgentTaskReloadParams struct {
	ID          int64 `json:"id"`
	SubstanceID int64 `json:"substance_id"`
}
