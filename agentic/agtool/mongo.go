package agtool

import (
	"log/slog"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/agentic"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoOptions struct {
	Models  []repository.CollectionInfer
	Timeout time.Duration
}

type Mongo struct {
	pdb   polydb.Sourcer
	ftb   agentic.BaseToolFactory
	log   *slog.Logger
	opts  MongoOptions
	colls map[string]CollectionInfo
}

func NewMongo(pdb polydb.Sourcer, ftb agentic.BaseToolFactory, log *slog.Logger, opts MongoOptions) *Mongo {
	colls := make(map[string]CollectionInfo, len(opts.Models))
	for _, mod := range opts.Models {
		name, desc := mod.CollectionInfo()
		colls[name] = CollectionInfo{Collection: name, Description: desc}
	}

	return &Mongo{pdb: pdb, ftb: ftb, log: log, opts: opts, colls: colls}
}

func (mgo *Mongo) NewEinoTool(cs agentic.ChatSession) ([]tool.BaseTool, error) {
	inst := mgo.newInstance(cs)
	tb := mgo.ftb.Builder()
	{
		const name = "mongo_list_collections"
		const desc = "列出当前可访问的 MongoDB 集合及说明，只读。\n\n" +
			"不确定集合名时先调用本工具，再用其他 mongo_* 工具查询。"
		tb = tb.Add(name, desc, inst.collections)
	}

	{
		const name = "mongo_find"
		const desc = "按条件查询多个 MongoDB 文档，只读。\n\n" +
			"适用：明细列表、批量查看。\n" +
			"规则：collection 必须在允许列表中；filter 使用标准 MongoDB 查询语法；limit 必填，范围 1~200。"
		tb = tb.Add(name, desc, inst.find)
	}

	{
		const name = "mongo_find_one"
		const desc = "按条件查询单个 MongoDB 文档，只读。\n\n" +
			"适用：按 _id、hostname、slug、machine_id 等唯一键查详情；想取最新一条时配合 sort。\n" +
			"不适用：需要多条用 mongo_find；统计数量用 mongo_count_documents。\n\n" +
			"规则：\n" +
			"1. collection 必须在允许访问的集合列表中。\n" +
			"2. filter 使用标准 MongoDB 查询语法。\n" +
			"3. 没有匹配文档时 found=false。"
		tb = tb.Add(name, desc, inst.findOne)
	}

	{
		const name = "mongo_distinct"
		const desc = "查询某个字段的去重值，只读。\n\n" +
			"适用：状态、版本、机房、中继节点等枚举值。\n" +
			"field 支持点号路径，例如 broker.slug。"
		tb = tb.Add(name, desc, inst.distinct)
	}

	{
		const name = "mongo_count_documents"
		const desc = "统计匹配的 MongoDB 文档数量，只读，不返回文档内容。\n\n" +
			"适用：确认数据量、统计在线/离线数量。"
		tb = tb.Add(name, desc, inst.countDocuments)
	}

	{
		const name = "mongo_aggregate"
		const desc = "执行只读 MongoDB 聚合管道。\n\n" +
			"适用：分组统计、趋势、TopN、分布分析。\n" +
			"pipeline 使用标准 MongoDB 聚合语法；数据库账号必须是只读账号。"
		tb = tb.Add(name, desc, inst.aggregate)
	}

	return tb.Build()
}

func (mgo *Mongo) newInstance(cs agentic.ChatSession) *mongoInstance {
	return &mongoInstance{
		pdb:     mgo.pdb,
		log:     mgo.log,
		colls:   mgo.colls,
		sess:    cs,
		timeout: 10 * time.Second,
	}
}

type mongoInstance struct {
	pdb     polydb.Sourcer
	log     *slog.Logger
	opts    MongoOptions
	colls   map[string]CollectionInfo
	sess    agentic.ChatSession
	timeout time.Duration
}

func (mgo *mongoInstance) newCollection(collName string) *mongo.Collection {
	mdb := mgo.pdb.MDB()
	db := mdb.Database()

	return db.Collection(collName)
}

func (mgo *mongoInstance) isAllowedCollection(collName string) error {
	if _, exists := mgo.colls[collName]; exists {
		return nil
	}

	return errcode.FmtDisallowedCollection.Fmt(collName)
}
