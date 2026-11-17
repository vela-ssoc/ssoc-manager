import type { AgentToRendererMessage } from "@a2ui/web_core/v1_0"

export type ChatRole = "user" | "assistant"

export type ChatStatus = "streaming" | "done" | "error"

export type ChatToolStatus = "running" | "done" | "error"

export interface ChatTextBlock {
  id: string
  kind: "text"
  text: string
}

export interface ChatUIBlock {
  id: string
  kind: "ui"
  revision: number
  messages: AgentToRendererMessage[]
}

export interface ChatToolCall {
  id: string
  name: string
  arguments: string
  result?: string
  status: ChatToolStatus
  error?: string
}

export interface ChatToolBlock {
  id: string
  kind: "tool"
  toolCall: ChatToolCall
}

export type ChatBlock = ChatTextBlock | ChatUIBlock | ChatToolBlock

export interface ChatTurn {
  id: string
  role: ChatRole
  content: string
  wire?: string
  blocks: ChatBlock[]
  status: ChatStatus
  error?: string
}