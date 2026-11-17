// Package agtool 提供一组只读的 MongoDB eino 工具。
//
// 用法：
//
//	mongo := agtool.NewMongo(pdb, []repository.CollectionInfer{
//	    model.Agent{},
//	    model.Broker{},
//	}, log)
//	tools, err := mongo.NewEinoTool(sess)
//
// 集合白名单由 mods 决定；不在列表中的集合会被拒绝。
// MongoDB 账号必须是只读账号。
package agtool
