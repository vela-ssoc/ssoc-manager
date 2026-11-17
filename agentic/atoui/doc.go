// Package atoui 把 A2UI 的声明式 UI 包装成 eino 可识别的工具。
//
// 用法：
//
//	ui := atoui.NewUI(atoui.UIOptions{})
//	tools, err := ui.NewEinoTool(sess)
//
// 每个工具负责把输入映射成 A2UI surface，并通过 ChatSession.Emit 推送。
// 新增一种 UI：定义 Input/Output、Build 函数，并在 NewEinoTool 中注册。
package atoui
