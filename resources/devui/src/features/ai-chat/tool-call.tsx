import { ChevronDown, Loader2, TriangleAlert, Wrench } from "lucide-react"

import { cn } from "@/lib/utils"
import type { ChatToolCall } from "@/features/ai-chat/types"

const MAX_PREVIEW = 20000

function shortText(value: string, max = 80) {
  return value.length > max ? `${value.slice(0, max)}…` : value
}

function parsedResult(call: ChatToolCall): Record<string, unknown> | null {
  if (!call.result) return null
  try {
    const value = JSON.parse(call.result) as unknown
    if (value && typeof value === "object" && !Array.isArray(value)) {
      return value as Record<string, unknown>
    }
  } catch {
    return null
  }
  return null
}

function summary(call: ChatToolCall) {
  if (call.status === "running") return "运行中"
  if (call.status === "error") return "失败"

  const value = parsedResult(call)
  if (value) {
    if (typeof value.error === "string" && value.error) return "失败"
    const parts: string[] = []
    if (typeof value.collection === "string" && value.collection) {
      parts.push(value.collection)
    }
    if (typeof value.count === "number") parts.push(`${value.count} 条`)
    if (value.truncated === true) parts.push("已截断")
    if (parts.length) return parts.join(" · ")
    if (typeof value.message === "string" && value.message) {
      return shortText(value.message)
    }
    if (typeof value.surface_id === "string") return "已渲染画布"
  }

  return "完成"
}

function prettyJSON(text: string) {
  let output: string
  try {
    output = JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    output = text
  }
  return output.length > MAX_PREVIEW
    ? `${output.slice(0, MAX_PREVIEW)}\n…`
    : output
}

function Section({ title, text }: { title: string; text: string }) {
  return (
    <div>
      <div className="text-muted-foreground mb-1 text-[11px]">{title}</div>
      <pre className="bg-muted max-h-80 overflow-auto rounded p-2 text-[11px] leading-relaxed break-words whitespace-pre-wrap">
        {prettyJSON(text)}
      </pre>
    </div>
  )
}

export function ToolCallView({ call }: { call: ChatToolCall }) {
  const running = call.status === "running"
  const failed = call.status === "error"

  return (
    <details className="bg-card group rounded-md border">
      <summary className="flex cursor-pointer list-none items-center gap-2 px-3 py-2 text-xs [&::-webkit-details-marker]:hidden">
        <Wrench className="text-muted-foreground size-3.5 shrink-0" />
        <span className="font-mono">{call.name || "tool"}</span>
        {running && <Loader2 className="size-3 animate-spin" />}
        {failed && (
          <TriangleAlert className="size-3.5 shrink-0 text-red-600 dark:text-red-400" />
        )}
        <span
          className={cn(
            "ml-auto min-w-0 truncate",
            running && "text-muted-foreground",
            !running && !failed && "text-green-600 dark:text-green-400",
            failed && "text-red-600 dark:text-red-400"
          )}
        >
          {summary(call)}
        </span>
        <ChevronDown className="text-muted-foreground size-3.5 shrink-0 transition-transform group-open:rotate-180" />
      </summary>

      <div className="space-y-2 border-t px-3 py-2">
        {call.arguments ? <Section title="参数" text={call.arguments} /> : null}
        {call.result ? <Section title="结果" text={call.result} /> : null}
        {failed && call.error ? (
          <div className="text-xs break-words text-red-600 dark:text-red-400">
            {call.error}
          </div>
        ) : null}
        {!call.arguments && !call.result && !(failed && call.error) ? (
          <div className="text-muted-foreground text-xs">
            {running ? "等待工具返回…" : "工具未返回内容"}
          </div>
        ) : null}
      </div>
    </details>
  )
}