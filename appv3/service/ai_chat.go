package service

import (
	"context"
	"log/slog"
	"slices"

	"github.com/cloudwego/eino/schema"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AIChat struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewAIChat(pdb polydb.Sourcer, log *slog.Logger) *AIChat {
	return &AIChat{
		pdb: pdb,
		log: log,
	}
}

func (ac *AIChat) Chat(parent context.Context, sse *eventsource.Writer, input request.AIChatRunAgentInput) error {
	return nil
}

func (ac *AIChat) Threads(ctx context.Context, userID bson.ObjectID) ([]*model.AIChatThread, error) {
	mdb := ac.pdb.MDB()
	filter := bson.D{{Key: "operator.id", Value: userID}}
	coll := mdb.AIChatThread()
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})

	return coll.Find(ctx, filter, opts)
}

func (ac *AIChat) SchemaMessages(ctx context.Context, userID bson.ObjectID, sessID string) ([]*schema.Message, error) {
	coll := ac.pdb.MDB().AIChatMessage()
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(100)
	dats, err := coll.Messages(ctx, userID, sessID, opts)
	if err != nil {
		return nil, err
	}

	msgs := make([]*schema.Message, 0, len(dats))
	for _, v := range slices.Backward(dats) {
		switch v.Role {
		case model.AIRoleUser:
			msgs = append(msgs, schema.UserMessage(v.Content))
		case model.AIRoleAssistant:
			msgs = append(msgs, schema.AssistantMessage(v.Content, nil))
		}
	}

	return msgs, nil
}

func (ac *AIChat) LLM(ctx context.Context, modelName string) (*model.AIChatModel, error) {
	coll := ac.pdb.MDB().AIChatModel()
	mod, err := coll.Get(ctx)
	if err != nil {
		return nil, errcode.ErrLLModelNotConfigured
	}
	if !mod.Enabled {
		return nil, errcode.ErrLLModelDisabled
	}
	if !slices.Contains(mod.Models, modelName) {
		return nil, errcode.FmtForbiddenModel.Fmt(modelName)
	}

	return mod, nil
}
