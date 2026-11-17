import { getToken } from "@/lib/http"

const CHAT_PATH = "/api/v3/sse/ai-chat/run"

export interface AgUiMessage {
  id: string
  role: "user" | "assistant"
  name?: string
  content: string
}

export interface RunAgentInput {
  threadId: string
  runId: string
  model?: string
  messages: AgUiMessage[]
  tools: unknown[]
  context: unknown[]
  state: Record<string, unknown>
  forwardedProps: Record<string, unknown>
}

export interface AgUiEvent {
  type: string
  [key: string]: unknown
}

export type AgUiHandler = (event: AgUiEvent) => void

export async function streamAgent(
  input: RunAgentInput,
  onEvent: AgUiHandler,
  signal?: AbortSignal
) {
  const token = getToken()
  const response = await fetch(CHAT_PATH, {
    method: "POST",
    headers: {
      Accept: "text/event-stream",
      "Cache-Control": "no-cache",
      "Content-Type": "application/json",
      ...(token ? { Authorization: token } : {}),
    },
    body: JSON.stringify({
      protocolVersion: "1.0",
      ...input,
    }),
    signal,
  })

  if (!response.ok || !response.body) {
    throw new Error(await errorMessage(response))
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""

  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let index = buffer.indexOf("\n\n")
    while (index >= 0) {
      dispatch(buffer.slice(0, index), onEvent)
      buffer = buffer.slice(index + 2)
      index = buffer.indexOf("\n\n")
    }
  }
  if (buffer.trim()) dispatch(buffer, onEvent)
}

function dispatch(chunk: string, onEvent: AgUiHandler) {
  const data: string[] = []
  for (const raw of chunk.split("\n")) {
    const line = raw.replace(/\r$/, "")
    if (!line.startsWith("data:")) continue
    data.push(line.slice(5).replace(/^ /, ""))
  }
  if (!data.length) return

  const payload = data.join("\n")
  if (!payload) return
  try {
    const event = JSON.parse(payload) as AgUiEvent
    if (event && typeof event.type === "string") onEvent(event)
  } catch {
    throw new Error("AI 返回了无法解析的事件数据")
  }
}

async function errorMessage(response: Response) {
  try {
    const payload = (await response.json()) as {
      detail?: string
      title?: string
    }
    const text = payload?.detail || payload?.title
    if (text) return text
  } catch {
    return `AI 对话请求失败（${response.status}）`
  }
  return `AI 对话请求失败（${response.status}）`
}