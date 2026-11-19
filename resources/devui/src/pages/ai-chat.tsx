import { useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import type { ActionPayload } from "@a2ui/web_core"
import {
  ArrowUp,
  Loader2,
  Plus,
  Settings2,
  Sparkles,
  Square,
  TriangleAlert,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { toast } from "@/lib/toast"
import { streamAgent } from "@/features/ai-chat/api"
import { A2uiMessages } from "@/features/ai-chat/a2ui-renderer"
import { ToolCallView } from "@/features/ai-chat/tool-call"
import { aiChatModelApi } from "@/features/ai-chat-model/api"
import type { AgUiMessage } from "@/features/ai-chat/api"
import type {
  ChatToolBlock,
  ChatToolCall,
  ChatTurn,
  ChatUIBlock,
} from "@/features/ai-chat/types"

const MODEL_STORAGE_KEY = "devui.ai-chat.model"

const SUGGESTIONS = [
  "统计当前探针节点的在线 / 离线分布，用饼图展示",
  "按中继节点统计在线探针数量，用柱状图展示",
  "统计探针版本分布情况",
  "最近 7 天探针上线趋势",
]

function newID() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
}

function readStoredModel() {
  try {
    return localStorage.getItem(MODEL_STORAGE_KEY) ?? ""
  } catch {
    return ""
  }
}

function historyMessages(turns: ChatTurn[]): AgUiMessage[] {
  return turns
    .filter((turn) => turn.role === "user" || turn.content.trim() !== "")
    .map((turn) => ({
      id: turn.id,
      role: turn.role,
      content:
        turn.role === "user" ? (turn.wire ?? turn.content) : turn.content,
    }))
}

function parseToolError(content: string) {
  try {
    const value = JSON.parse(content) as { error?: unknown }
    return typeof value.error === "string" && value.error ? value.error : null
  } catch {
    return null
  }
}

function failRunningTools(turn: ChatTurn, message: string): ChatTurn {
  return {
    ...turn,
    blocks: turn.blocks.map((block) => {
      if (block.kind === "tool" && block.toolCall.status === "running") {
        return {
          ...block,
          toolCall: { ...block.toolCall, status: "error", error: message },
        }
      }
      return block
    }),
  }
}

function missingModelHint() {
  return "未配置可用模型，请先在「配置管理 → 聊天模型」中配置并启用"
}

function actionLabel(action: ActionPayload) {
  const confirmed = action.context?.confirmed
  if (typeof confirmed === "boolean") {
    return confirmed ? "已确认操作" : "已取消操作"
  }
  return "已提交表单"
}

export function AIChatPage() {
  const [turns, setTurns] = useState<ChatTurn[]>([])
  const [input, setInput] = useState("")
  const [sid, setSid] = useState(newID)
  const [streaming, setStreaming] = useState(false)
  const [models, setModels] = useState<string[]>([])
  const [modelsLoaded, setModelsLoaded] = useState(false)
  const [model, setModel] = useState("")
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const stickToBottom = useRef(true)

  const scrollToBottom = () => {
    const el = scrollRef.current
    if (el) el.scrollTop = el.scrollHeight
  }

  useEffect(() => {
    if (stickToBottom.current) scrollToBottom()
  }, [turns])

  useEffect(() => {
    const content = contentRef.current
    if (!content || typeof ResizeObserver === "undefined") return
    const observer = new ResizeObserver(() => {
      if (stickToBottom.current) scrollToBottom()
    })
    observer.observe(content)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    let cancelled = false
    aiChatModelApi
      .models()
      .then((data) => {
        if (cancelled) return
        const list = Array.from(
          new Set((data.records ?? []).filter(Boolean))
        )
        setModels(list)
        setModel((prev) => {
          if (prev && list.includes(prev)) return prev
          const stored = readStoredModel()
          if (stored && list.includes(stored)) return stored
          return list[0] ?? ""
        })
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelled) setModelsLoaded(true)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const selectModel = (value: string) => {
    setModel(value)
    try {
      localStorage.setItem(MODEL_STORAGE_KEY, value)
    } catch {
      return
    }
  }

  const patch = (id: string, next: (turn: ChatTurn) => ChatTurn) =>
    setTurns((prev) =>
      prev.map((turn) => (turn.id === id ? next(turn) : turn))
    )

  const run = async (messages: AgUiMessage[], userTurn: ChatTurn | null) => {
    const assistantID = newID()
    stickToBottom.current = true
    setStreaming(true)
    setTurns((prev) => [
      ...prev,
      ...(userTurn ? [userTurn] : []),
      {
        id: assistantID,
        role: "assistant",
        content: "",
        blocks: [],
        status: "streaming",
      },
    ])

    const controller = new AbortController()
    abortRef.current = controller

    const appendText = (delta: string) => {
      patch(assistantID, (turn) => {
        const blocks = [...turn.blocks]
        const last = blocks[blocks.length - 1]
        if (last && last.kind === "text") {
          blocks[blocks.length - 1] = { ...last, text: last.text + delta }
        } else {
          blocks.push({ id: newID(), kind: "text", text: delta })
        }
        return { ...turn, content: turn.content + delta, blocks }
      })
    }

    const upsertUI = (
      id: string,
      operations: unknown[],
      replace: boolean
    ) => {
      const parsed = operations.filter(
        (item): item is ChatUIBlock["messages"][number] =>
          Boolean(item) && typeof item === "object"
      )
      if (!parsed.length) return
      const blockID = id || `activity-${newID()}`

      patch(assistantID, (turn) => {
        const index = turn.blocks.findIndex(
          (block) => block.kind === "ui" && block.id === blockID
        )
        if (index < 0) {
          return {
            ...turn,
            blocks: [
              ...turn.blocks,
              { id: blockID, kind: "ui", revision: 0, messages: parsed },
            ],
          }
        }

        const block = turn.blocks[index] as ChatUIBlock
        const blocks = [...turn.blocks]
        blocks[index] = replace
          ? { ...block, revision: block.revision + 1, messages: parsed }
          : { ...block, messages: [...block.messages, ...parsed] }
        return { ...turn, blocks }
      })
    }

    const patchTool = (
      toolCallID: string,
      next: (call: ChatToolCall) => ChatToolCall,
      name = ""
    ) => {
      patch(assistantID, (turn) => {
        const index = turn.blocks.findIndex(
          (block) => block.kind === "tool" && block.id === toolCallID
        )
        if (index < 0) {
          const toolCall = next({
            id: toolCallID,
            name,
            arguments: "",
            status: "running",
          })
          return {
            ...turn,
            blocks: [
              ...turn.blocks,
              { id: toolCallID, kind: "tool", toolCall },
            ],
          }
        }

        const block = turn.blocks[index] as ChatToolBlock
        const blocks = [...turn.blocks]
        blocks[index] = { ...block, toolCall: next(block.toolCall) }
        return { ...turn, blocks }
      })
    }

    const applyDelta = (event: Record<string, unknown>) => {
      const activityID =
        typeof event.messageId === "string" ? event.messageId : ""
      const patchList = Array.isArray(event.patch) ? event.patch : []
      const added: unknown[] = []
      let replaced: unknown[] | null = null

      for (const raw of patchList) {
        if (!raw || typeof raw !== "object") continue
        const op = raw as { op?: string; path?: string; value?: unknown }
        if (op.op === "add" && op.path === "/a2ui_operations/-") {
          added.push(op.value)
        } else if (
          op.op === "replace" &&
          op.path === "/a2ui_operations" &&
          Array.isArray(op.value)
        ) {
          replaced = op.value
        }
      }

      if (replaced) {
        upsertUI(activityID, replaced, true)
      } else if (added.length) {
        upsertUI(activityID, added, false)
      }
    }

    try {
      await streamAgent(
        {
          threadId: sid,
          runId: newID(),
          model,
          messages,
          tools: [],
          context: [],
          state: {},
          forwardedProps: {},
        },
        (event) => {
          switch (event.type) {
            case "TEXT_MESSAGE_CONTENT": {
              const delta =
                typeof event.delta === "string" ? event.delta : ""
              if (!delta) return
              appendText(delta)
              return
            }
            case "TOOL_CALL_START": {
              const toolCallID =
                typeof event.toolCallId === "string" ? event.toolCallId : ""
              const name =
                typeof event.toolCallName === "string"
                  ? event.toolCallName
                  : ""
              if (!toolCallID) return
              patchTool(
                toolCallID,
                (call) => ({
                  ...call,
                  name: name || call.name,
                  status: "running",
                }),
                name
              )
              return
            }
            case "TOOL_CALL_ARGS": {
              const toolCallID =
                typeof event.toolCallId === "string" ? event.toolCallId : ""
              const delta =
                typeof event.delta === "string" ? event.delta : ""
              if (!toolCallID || !delta) return
              patchTool(toolCallID, (call) => ({
                ...call,
                arguments: call.arguments + delta,
              }))
              return
            }
            case "TOOL_CALL_END":
              return
            case "TOOL_CALL_RESULT": {
              const toolCallID =
                typeof event.toolCallId === "string" ? event.toolCallId : ""
              if (!toolCallID) return
              const raw = event.content
              const content =
                typeof raw === "string"
                  ? raw
                  : raw == null
                    ? ""
                    : JSON.stringify(raw)
              const toolError = parseToolError(content)
              patchTool(toolCallID, (call) => ({
                ...call,
                result: content,
                status: toolError ? "error" : "done",
                error: toolError ?? undefined,
              }))
              return
            }
            case "ACTIVITY_SNAPSHOT": {
              if (event.activityType !== "a2ui-surface") return
              const activityID =
                typeof event.messageId === "string" ? event.messageId : ""
              const content = event.content as
                | { a2ui_operations?: unknown }
                | undefined
              if (Array.isArray(content?.a2ui_operations)) {
                upsertUI(
                  activityID,
                  content.a2ui_operations,
                  event.replace !== false
                )
              }
              return
            }
            case "ACTIVITY_DELTA": {
              if (event.activityType !== "a2ui-surface") return
              applyDelta(event)
              return
            }
            case "CUSTOM": {
              if (event.name === "a2ui" && event.value) {
                upsertUI(`custom-${newID()}`, [event.value], true)
              }
              return
            }
            case "TEXT_MESSAGE_END":
              patch(assistantID, (turn) => ({ ...turn, status: "done" }))
              return
            case "RUN_FINISHED":
              patch(assistantID, (turn) =>
                failRunningTools(
                  { ...turn, status: "done" },
                  "工具未返回结果"
                )
              )
              return
            case "RUN_ERROR": {
              const message =
                typeof event.message === "string"
                  ? event.message
                  : "AI 服务返回错误"
              patch(assistantID, (turn) =>
                failRunningTools(
                  { ...turn, status: "error", error: message },
                  message
                )
              )
              return
            }
          }
        },
        controller.signal
      )
      patch(assistantID, (turn) =>
        turn.status === "streaming" ? { ...turn, status: "done" } : turn
      )
    } catch (err) {
      const aborted =
        controller.signal.aborted ||
        (err instanceof DOMException && err.name === "AbortError")
      if (aborted) {
        patch(assistantID, (turn) =>
          failRunningTools({ ...turn, status: "done" }, "已停止生成")
        )
      } else {
        const message =
          err instanceof Error ? err.message : "AI 对话请求失败"
        patch(assistantID, (turn) =>
          failRunningTools({ ...turn, status: "error", error: message }, message)
        )
      }
    } finally {
      abortRef.current = null
      setStreaming(false)
    }
  }

  const send = async (text?: string) => {
    const query = (text ?? input).trim()
    if (!query || streaming) return
    if (!model) {
      toast.error(missingModelHint())
      return
    }

    const userID = newID()
    const userTurn: ChatTurn = {
      id: userID,
      role: "user",
      content: query,
      blocks: [],
      status: "done",
    }
    const messages: AgUiMessage[] = [
      ...historyMessages(turns),
      { id: userID, role: "user", content: query },
    ]

    setInput("")
    await run(messages, userTurn)
  }

  const performAction = (action: ActionPayload) => {
    if (streaming) {
      toast.error("请等待当前回复结束")
      return
    }
    if (!model) {
      toast.error(missingModelHint())
      return
    }

    const userID = newID()
    const wire = `A2UI user action:\n${JSON.stringify(action)}`
    const userTurn: ChatTurn = {
      id: userID,
      role: "user",
      content: actionLabel(action),
      wire,
      blocks: [],
      status: "done",
    }
    const messages: AgUiMessage[] = [
      ...historyMessages(turns),
      {
        id: userID,
        role: "user",
        name: "a2ui_action_context",
        content: wire,
      },
    ]
    void run(messages, userTurn)
  }

  const stop = () => {
    abortRef.current?.abort()
  }

  const reset = () => {
    if (streaming) stop()
    setTurns([])
    setInput("")
    setSid(newID())
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 px-4 py-2">
        <div className="mx-auto flex w-full max-w-3xl items-center gap-2">
          <Sparkles className="text-muted-foreground size-4 shrink-0" />
          <h2 className="text-sm font-semibold">AI 助手</h2>
          <Badge variant="outline">实验性</Badge>
          <span className="min-w-0 flex-1" />
          <Button variant="outline" size="sm" onClick={reset}>
            <Plus /> 新会话
          </Button>
        </div>
      </div>

      <div
        ref={scrollRef}
        className="min-h-0 flex-1 [overflow-anchor:none] overflow-y-auto overscroll-contain"
        onScroll={(event) => {
          const el = event.currentTarget
          stickToBottom.current =
            el.scrollHeight - el.scrollTop - el.clientHeight < 80
        }}
      >
        <div
          ref={contentRef}
          className={
            turns.length === 0
              ? "mx-auto flex min-h-full w-full max-w-3xl items-center justify-center px-4 py-6"
              : "mx-auto w-full max-w-3xl px-4 py-6"
          }
        >
          {turns.length === 0 ? (
            <EmptyState onPick={(text) => void send(text)} />
          ) : (
            <div className="flex flex-col gap-6">
              {turns.map((turn) => (
                <TurnView
                  key={turn.id}
                  turn={turn}
                  onAction={performAction}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="shrink-0 border-t">
        <div className="mx-auto w-full max-w-3xl px-4 py-3">
          <div className="bg-card focus-within:ring-ring/40 rounded-md border shadow-sm transition-shadow focus-within:ring-2">
            <Textarea
              value={input}
              aria-label="输入问题"
              placeholder={
                modelsLoaded && !model
                  ? "请先在「配置管理 → 聊天模型」中配置并启用模型"
                  : "询问探针、中继节点、资产或告警数据…"
              }
              className="max-h-40 min-h-[64px] resize-none border-0 bg-transparent px-3 pt-2.5 pb-1 shadow-none focus-visible:ring-0"
              onChange={(event) => setInput(event.target.value)}
              onKeyDown={(event) => {
                if (
                  event.key === "Enter" &&
                  !event.shiftKey &&
                  !event.nativeEvent.isComposing
                ) {
                  event.preventDefault()
                  void send()
                }
              }}
            />
            <div className="flex items-center gap-1.5 px-1.5 pb-1.5">
              {models.length > 0 && (
                <Select
                  value={model}
                  onValueChange={selectModel}
                  disabled={streaming}
                >
                  <SelectTrigger
                    size="sm"
                    className="text-muted-foreground hover:bg-accent/60 data-[size=sm]:h-6 h-6 max-w-[12rem] gap-1 border-0 bg-transparent px-1.5 font-mono text-[11px] shadow-none focus-visible:ring-1 [&_svg]:size-3"
                    aria-label="选择聊天模型"
                  >
                    <SelectValue placeholder="选择模型" />
                  </SelectTrigger>
                  <SelectContent side="top" align="start" className="max-h-72">
                    {models.map((item) => (
                      <SelectItem
                        key={item}
                        value={item}
                        className="font-mono text-xs"
                      >
                        {item}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              {modelsLoaded && models.length === 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  asChild
                  className="text-muted-foreground h-6 px-1.5 text-[11px]"
                >
                  <Link to="/configs/ai-chat-model">
                    <Settings2 /> 配置模型
                  </Link>
                </Button>
              )}
              <span className="min-w-0 flex-1" />
              {streaming ? (
                <Button
                  size="icon"
                  variant="outline"
                  className="size-7"
                  aria-label="停止生成"
                  onClick={stop}
                >
                  <Square />
                </Button>
              ) : (
                <Button
                  size="icon"
                  className="size-7"
                  aria-label="发送"
                  disabled={!input.trim() || !model}
                  onClick={() => void send()}
                >
                  <ArrowUp />
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function EmptyState({ onPick }: { onPick: (text: string) => void }) {
  return (
    <div className="flex flex-col items-center gap-5 py-10 text-center">
      <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-lg">
        <Sparkles className="size-6" />
      </div>
      <div className="grid gap-1">
        <h3 className="text-base font-semibold">SOC 智能助手</h3>
        <p className="text-muted-foreground text-xs">
          用自然语言查询探针与中继节点数据，Agent 会自动生成图表
        </p>
      </div>
      <div className="grid w-full max-w-lg gap-2 sm:grid-cols-2">
        {SUGGESTIONS.map((text) => (
          <button
            key={text}
            type="button"
            className="bg-card hover:bg-accent/60 rounded-md border px-3 py-2 text-left text-xs leading-relaxed transition-colors"
            onClick={() => onPick(text)}
          >
            {text}
          </button>
        ))}
      </div>
    </div>
  )
}

function TurnView({
  turn,
  onAction,
}: {
  turn: ChatTurn
  onAction: (action: ActionPayload) => void
}) {
  if (turn.role === "user") {
    return (
      <div className="flex justify-end">
        <div className="bg-muted max-w-[85%] rounded-2xl rounded-br-md px-4 py-2.5 text-sm leading-relaxed break-words whitespace-pre-wrap">
          {turn.content}
        </div>
      </div>
    )
  }

  const pending = turn.status === "streaming" && turn.blocks.length === 0

  return (
    <div className="space-y-2.5">
      <div className="text-muted-foreground flex items-center gap-1.5 text-xs font-medium">
        <span className="bg-primary/10 text-primary flex size-5 shrink-0 items-center justify-center rounded-md">
          <Sparkles className="size-3" />
        </span>
        AI 助手
        {turn.status === "streaming" && (
          <Loader2 className="size-3 animate-spin" />
        )}
      </div>

      {turn.blocks.map((block) => {
        if (block.kind === "text") {
          return (
            <div key={block.id} className="markdown-body text-sm">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {block.text}
              </ReactMarkdown>
            </div>
          )
        }
        if (block.kind === "ui") {
          return (
            <A2uiMessages
              key={`${block.id}-${block.revision}`}
              messages={block.messages}
              onAction={onAction}
            />
          )
        }
        return <ToolCallView key={block.id} call={block.toolCall} />
      })}
      {pending && (
        <div className="text-muted-foreground flex items-center gap-2 text-xs">
          <Loader2 className="size-3.5 animate-spin" />
          正在思考…
        </div>
      )}
      {turn.status === "done" && turn.blocks.length === 0 && (
        <div className="text-muted-foreground text-xs">（无回复）</div>
      )}
      {turn.status === "error" && (
        <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-start gap-2 rounded-md border p-2.5 text-xs">
          <TriangleAlert className="mt-px size-4 shrink-0" />
          <span className="break-words">{turn.error || "AI 服务返回错误"}</span>
        </div>
      )}
    </div>
  )
}