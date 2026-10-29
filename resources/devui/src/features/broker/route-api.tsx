import { useEffect, useMemo, useRef, useState } from "react"
import { createPortal } from "react-dom"
import Editor from "@monaco-editor/react"
import { Braces, Loader2, Play, Trash2 } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"
import { rawRequest, type RawResponse } from "@/lib/http"
import { useIsDark } from "@/features/zip-file/file-icon"
import { getMonacoFontFamily } from "@/lib/fonts"
import { useTheme } from "@/providers/theme-provider"
import { brokerApi } from "@/features/broker/api"
import type { BrokerRouteInfo } from "@/features/broker/types"

type RouteKind = "managers" | "agents" | "exposes"

export function RouteApiSection({
  slug,
  online,
  kind,
}: {
  slug: string
  online: boolean
  kind: RouteKind
}) {
  const [routes, setRoutes] = useState<BrokerRouteInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!online) {
      setRoutes([])
      setError(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(false)
    brokerApi
      .routes(slug, kind)
      .then((data) => {
        if (!cancelled) setRoutes(data.records ?? [])
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [slug, online, kind])

  if (!online) return null

  if (loading && routes.length === 0) {
    return (
      <div className="text-muted-foreground flex items-center gap-2 text-xs">
        <Loader2 className="size-3.5 animate-spin" />
        正在加载接口列表…
      </div>
    )
  }

  if (error && routes.length === 0) {
    return (
      <div className="text-muted-foreground text-xs">
        无法获取接口列表，请确认节点在线且连接正常
      </div>
    )
  }

  if (kind === "agents" || kind === "exposes") {
    return <AgentRoutesTable routes={routes} />
  }

  return <RoutesExplorer slug={slug} routes={routes} />
}

function AgentRoutesTable({ routes }: { routes: BrokerRouteInfo[] }) {
  if (routes.length === 0) {
    return <div className="text-muted-foreground text-xs">暂无探针路由</div>
  }

  return (
    <div className="overflow-hidden rounded-md border">
      <table className="w-full table-fixed text-xs">
        <thead>
          <tr className="bg-muted/40 text-muted-foreground border-b">
            <th className="w-28 px-3 py-1.5 text-left font-medium">方法</th>
            <th className="px-3 py-1.5 text-left font-medium">路径</th>
          </tr>
        </thead>
        <tbody>
          {routes.map((route, index) => (
            <tr key={index} className="border-b last:border-b-0">
              <td className="px-3 py-1.5">
                <span className={cn("font-mono text-xs font-semibold", methodClass(route.method))}>
                  {route.method}
                </span>
              </td>
              <td className="px-3 py-1.5">
                <TooltipHint content={route.path}>
                  <div className="truncate font-mono">{route.path}</div>
                </TooltipHint>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function fillPath(template: string, values: Record<string, string>): string {
  const segments = template.split("/")
  const filled = segments.map((segment) => {
    if (segment.startsWith(":")) {
      return values[segment.slice(1)] ?? segment
    }
    if (segment.startsWith("*")) {
      return values[segment.slice(1)] ?? ""
    }

    return segment
  })

  return filled.join("/")
}

type RequestPane = "path" | "query" | "headers" | "body"

const METHOD_ORDER = ["GET", "POST", "PUT", "PATCH", "DELETE"]

const HTTP_HEADER_NAMES = [
  "Accept",
  "Accept-Charset",
  "Accept-Encoding",
  "Accept-Language",
  "Accept-Ranges",
  "Access-Control-Request-Headers",
  "Access-Control-Request-Method",
  "Authorization",
  "Cache-Control",
  "Connection",
  "Content-Encoding",
  "Content-Length",
  "Content-Type",
  "Cookie",
  "Date",
  "DNT",
  "Expect",
  "Forwarded",
  "From",
  "Host",
  "If-Match",
  "If-Modified-Since",
  "If-None-Match",
  "If-Range",
  "If-Unmodified-Since",
  "Keep-Alive",
  "Max-Forwards",
  "Origin",
  "Pragma",
  "Prefer",
  "Proxy-Authorization",
  "Range",
  "Referer",
  "Referrer-Policy",
  "Sec-CH-UA",
  "Sec-CH-UA-Mobile",
  "Sec-CH-UA-Platform",
  "Sec-Fetch-Dest",
  "Sec-Fetch-Mode",
  "Sec-Fetch-Site",
  "Sec-Fetch-User",
  "TE",
  "Trailer",
  "Transfer-Encoding",
  "Upgrade",
  "Upgrade-Insecure-Requests",
  "User-Agent",
  "Via",
  "Warning",
  "X-Forwarded-For",
  "X-Forwarded-Host",
  "X-Forwarded-Proto",
  "X-Requested-With",
]

function RoutesExplorer({
  slug,
  routes,
}: {
  slug: string
  routes: BrokerRouteInfo[]
}) {
  const methods = useMemo(() => {
    const set = new Set<string>()
    routes.forEach((route) => set.add(route.method))
    const rest = Array.from(set).filter((m) => !METHOD_ORDER.includes(m))
    return [...METHOD_ORDER.filter((m) => set.has(m)), ...rest]
  }, [routes])

  const [method, setMethod] = useState<string>(methods[0] ?? "")
  const [pathIndex, setPathIndex] = useState(0)

  const methodRoutes = useMemo(
    () => routes.filter((route) => route.method === method),
    [routes, method]
  )

  const selected = methodRoutes[pathIndex]

  const [params, setParams] = useState<Record<string, string>>({})
  const [query, setQuery] = useState<KeyValueEntry[]>([{ key: "", value: "" }])
  const [headers, setHeaders] = useState<KeyValueEntry[]>([{ key: "", value: "" }])
  const [body, setBody] = useState("")
  const [bodyMode, setBodyMode] = useState<"json" | "text">("json")
  const [tab, setTab] = useState<RequestPane>("path")
  const [sending, setSending] = useState(false)
  const [result, setResult] = useState<RawResponse | null>(null)

  const pathParams = selected ? pathParamNames(selected.path) : []
  const hasPathParams = pathParams.length > 0
  const hasBody = ["POST", "PUT", "PATCH", "DELETE"].includes(selected?.method ?? "")

  const panes: RequestPane[] = useMemo(() => {
    const list: RequestPane[] = []
    if (hasPathParams) list.push("path")
    list.push("query")
    list.push("headers")
    if (hasBody) list.push("body")
    return list
  }, [hasPathParams, hasBody])

  const activeTab: RequestPane = panes.includes(tab) ? tab : panes[0]

  const changeMethod = (value: string) => {
    setMethod(value)
    setPathIndex(0)
    setBodyMode("json")
    setTab("path")
    resetRequest(setParams, setQuery, setHeaders, setBody, setResult)
  }

  const changePath = (index: number) => {
    setPathIndex(index)
    setBodyMode("json")
    setTab("path")
    resetRequest(setParams, setQuery, setHeaders, setBody, setResult)
  }

  const send = async () => {
    if (!selected) return
    const url = buildExplorerUrl(slug, selected.path, params, query)

    setSending(true)
    setResult(null)
    try {
      const init: { method: string; headers?: Record<string, string>; body?: string } = {
        method: selected.method,
      }
      const finalHeaders: Record<string, string> = {}
      headers.forEach(({ key, value }) => {
        if (key.trim() && value.trim()) finalHeaders[key.trim()] = value
      })
      if (Object.keys(finalHeaders).length > 0) init.headers = finalHeaders
      if (hasBody && body.trim()) init.body = body

      const res = await rawRequest(url, init)
      setResult(res)
    } catch (err) {
      setResult({
        status: 0,
        body: err instanceof Error ? err.message : "请求失败",
        contentType: "text/plain",
        headers: {},
      })
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-stretch gap-0 overflow-hidden rounded-md border">
        <div className="flex w-32 shrink-0 items-center border-r">
          <Select value={method} onValueChange={changeMethod}>
            <SelectTrigger className="h-full w-full rounded-none border-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {methods.map((m) => (
                <SelectItem key={m} value={m}>
                  <span className={cn("font-mono font-semibold", methodClass(m))}>{m}</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex min-w-0 flex-1 items-center border-r">
          <Select value={String(pathIndex)} onValueChange={(value) => changePath(Number(value))}>
            <SelectTrigger className="h-full w-full rounded-none border-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {methodRoutes.map((route, index) => (
                <SelectItem key={index} value={String(index)}>
                  <span className="font-mono">{route.path}</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button className="shrink-0 rounded-none" disabled={sending} onClick={() => void send()}>
          <Play /> {sending ? "发送中…" : "发送"}
        </Button>
      </div>

      <div className="mt-2 flex items-center gap-1">
        {hasPathParams && (
          <RequestTab active={activeTab === "path"} onClick={() => setTab("path")}>
            路径参数
          </RequestTab>
        )}
        <RequestTab active={activeTab === "query"} onClick={() => setTab("query")}>
          查询参数
        </RequestTab>
        <RequestTab active={activeTab === "headers"} onClick={() => setTab("headers")}>
          Header
        </RequestTab>
        {hasBody && (
          <RequestTab active={activeTab === "body"} onClick={() => setTab("body")}>
            请求报文
          </RequestTab>
        )}
      </div>

      <div className="mt-2 h-64 shrink-0 overflow-y-auto">
        {activeTab === "path" && (
          <KeyValueTable
            rows={buildPathParamRows(selected.path, params, setParams)}
          />
        )}

        {activeTab === "query" && (
          <KeyValueTable
            rows={buildQueryRows(query, setQuery)}
          />
        )}

        {activeTab === "headers" && (
          <KeyValueTable
            rows={buildHeaderRows(headers, setHeaders)}
          />
        )}

        {activeTab === "body" && (
          <div className="grid gap-2">
            <div className="flex items-center gap-1">
              <span className="text-muted-foreground text-xs font-medium">语法</span>
              <RequestBodyModeButton
                active={bodyMode === "json"}
                onClick={() => setBodyMode("json")}
              >
                JSON
              </RequestBodyModeButton>
              <RequestBodyModeButton
                active={bodyMode === "text"}
                onClick={() => setBodyMode("text")}
              >
                文本
              </RequestBodyModeButton>
            </div>
            <RequestBodyEditor
              mode={bodyMode}
              value={body}
              onChange={setBody}
            />
          </div>
        )}
      </div>

      <div className="mt-2 min-h-64 flex-1 border-t pt-2">
        <ResponseSection result={result} />
      </div>
    </div>
  )
}

function ResponseSection({ result }: { result: RawResponse | null }) {
  const [formatted, setFormatted] = useState(false)
  const headers = Object.entries(result?.headers ?? {})

  if (!result) {
    return (
      <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
        暂无响应，点击发送执行请求
      </div>
    )
  }

  const ct = result.contentType
  let language = "plaintext"
  if (ct.includes("json")) language = "json"
  else if (ct.includes("html")) language = "html"
  else if (ct.includes("xml")) language = "xml"
  else if (ct.includes("javascript")) language = "javascript"
  else if (ct.includes("css")) language = "css"

  const isJson = ct.includes("json")
  const displayBody = isJson && formatted ? prettyJson(result.body) : result.body

  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      <div className="flex shrink-0 items-center gap-2">
        <span className="text-muted-foreground text-xs font-medium">响应码</span>
        <span className={cn("font-mono text-xs font-semibold", statusClass(result.status))}>
          {result.status === 0 ? "发送失败" : result.status}
        </span>
        <span className="text-muted-foreground text-xs">{result.contentType}</span>
      </div>

      <div className="shrink-0 overflow-hidden rounded-md border p-3">
        <div className="text-muted-foreground text-xs font-medium">响应头</div>
        {headers.length === 0 ? (
          <div className="text-muted-foreground mt-1 text-xs">无</div>
        ) : (
          <div className="mt-1 grid max-h-40 gap-y-1 overflow-auto">
            {headers.map(([key, value]) => (
              <div key={key} className="grid grid-cols-[180px_1fr] gap-2 text-xs">
                <span className="text-muted-foreground truncate font-mono">{key}</span>
                <span className="min-w-0 truncate font-mono">{value}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="flex min-h-0 flex-1 flex-col gap-2 overflow-hidden rounded-md border p-3">
        <div className="flex shrink-0 items-center gap-1.5">
          <div className="text-muted-foreground text-xs font-medium">响应报文</div>
          {isJson && (
            <TooltipHint content={formatted ? "还原原始报文" : "格式化 JSON"}>
              <button
                type="button"
                className="text-muted-foreground hover:text-foreground inline-flex size-5 items-center justify-center rounded transition-colors"
                onClick={() => setFormatted((value) => !value)}
              >
                <Braces className="size-3.5" />
              </button>
            </TooltipHint>
          )}
        </div>
        <div className="min-h-0 flex-1 overflow-hidden">
          {result.imageUrl ? (
            <img
              src={result.imageUrl}
              alt="响应图片预览"
              className="max-h-80 w-fit max-w-full rounded-md border"
            />
          ) : (
            <ResponseEditor language={language} value={displayBody} />
          )}
        </div>
      </div>
    </div>
  )
}

function prettyJson(body: string): string {
  try {
    return JSON.stringify(JSON.parse(body), null, 2)
  } catch {
    return body
  }
}

function ResponseEditor({ language, value }: { language: string; value: string }) {
  const dark = useIsDark()
  const { font } = useTheme()

  return (
    <Editor
      path={`response.${language === "plaintext" ? "txt" : language}`}
      language={language}
      value={value}
      theme={dark ? "near-black" : "light"}
      beforeMount={(monaco) => {
        monaco.editor.defineTheme("near-black", {
          base: "vs-dark",
          inherit: true,
          rules: [],
          colors: {
            "editor.background": "#000000",
          },
        })
      }}
      options={{
        readOnly: true,
        minimap: { enabled: false },
        fontSize: 13,
        fontFamily: getMonacoFontFamily(font),
        scrollBeyondLastLine: false,
        automaticLayout: true,
        wordWrap: "on",
        domReadOnly: true,
      }}
      height="100%"
    />
  )
}

function resetRequest(
  setParams: (rows: Record<string, string>) => void,
  setQuery: (rows: KeyValueEntry[]) => void,
  setHeaders: (rows: KeyValueEntry[]) => void,
  setBody: (value: string) => void,
  setResult: (value: RawResponse | null) => void
) {
  setParams({})
  setQuery([{ key: "", value: "" }])
  setHeaders([{ key: "", value: "" }])
  setBody("")
  setResult(null)
}

function statusClass(status: number): string {
  if (status >= 200 && status < 300) return "text-green-600 dark:text-green-400"
  if (status === 0) return "text-destructive"

  return "text-orange-500"
}

function buildExplorerUrl(
  slug: string,
  template: string,
  params: Record<string, string>,
  query: KeyValueEntry[]
): string {
  const path = fillPath(template, params)
  const qs = new URLSearchParams()
  query.forEach(({ key, value }) => {
    if (key && value) qs.set(key, value)
  })
  const queryText = qs.toString()

  return `/br/${slug}${path}` + (queryText ? `?${queryText}` : "")
}

function pathParamNames(template: string): string[] {
  return template
    .split("/")
    .filter((segment) => segment.startsWith(":") || segment.startsWith("*"))
    .map((segment) => segment.slice(1))
}

function buildPathParamRows(
  template: string,
  params: Record<string, string>,
  setParams: (rows: Record<string, string>) => void
): KeyValueRowProps[] {
  return template
    .split("/")
    .filter((segment) => segment.startsWith(":") || segment.startsWith("*"))
    .map((segment) => {
      const name = segment.slice(1)
      const wildcard = segment.startsWith("*")
      return {
        key: name,
        value: params[name] ?? "",
        keyReadonly: true,
        keyHint: wildcard ? "(通配)" : "",
        valuePlaceholder: wildcard ? "匹配后续任意路径段" : "单个路径参数",
        onChangeValue: (value) => setParams({ ...params, [name]: value }),
      }
    })
}

function buildQueryRows(
  query: KeyValueEntry[],
  setQuery: (rows: KeyValueEntry[]) => void
): KeyValueRowProps[] {
  return query.map(({ key, value }, index) => ({
    key,
    value,
    onChangeKey: (nextKey) => {
      setQuery(upsertEntry(query, index, { key: nextKey }))
    },
    onChangeValue: (nextValue) => {
      setQuery(upsertEntry(query, index, { value: nextValue }))
    },
    onRemove: () => setQuery(removeEntry(query, index)),
  }))
}

function buildHeaderRows(
  headers: KeyValueEntry[],
  setHeaders: (rows: KeyValueEntry[]) => void
): KeyValueRowProps[] {
  return headers.map(({ key, value }, index) => ({
    key,
    value,
    keySuggestions: HTTP_HEADER_NAMES,
    onChangeKey: (nextKey) => {
      setHeaders(upsertEntry(headers, index, { key: nextKey }))
    },
    onChangeValue: (nextValue) => {
      setHeaders(upsertEntry(headers, index, { value: nextValue }))
    },
    onRemove: () => setHeaders(removeEntry(headers, index)),
  }))
}

interface KeyValueEntry {
  key: string
  value: string
}

function upsertEntry(
  entries: KeyValueEntry[],
  index: number,
  patch: Partial<KeyValueEntry>
): KeyValueEntry[] {
  const next = entries.map((entry, i) => (i === index ? { ...entry, ...patch } : entry))
  const current = next[index]
  if (current && (current.key.trim() || current.value.trim())) {
    const last = next[next.length - 1]
    if (last.key.trim() || last.value.trim()) next.push({ key: "", value: "" })
  }
  return next
}

function removeEntry(entries: KeyValueEntry[], index: number): KeyValueEntry[] {
  const next = entries.filter((_, i) => i !== index)
  if (next.length === 0) next.push({ key: "", value: "" })
  else {
    const last = next[next.length - 1]
    if (last.key.trim() || last.value.trim()) next.push({ key: "", value: "" })
  }
  return next
}

interface KeyValueRowProps {
  key: string
  value: string
  keyReadonly?: boolean
  keyHint?: string
  keySuggestions?: string[]
  valuePlaceholder?: string
  onChangeKey?: (value: string) => void
  onChangeValue?: (value: string) => void
  onRemove?: () => void
}

function KeyValueTable({
  rows,
}: {
  rows: KeyValueRowProps[]
}) {
  return (
    <div className="overflow-hidden rounded-md border">
      <table className="w-full table-fixed">
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b last:border-b-0">
              <td className="w-[30%] border-r p-0">
                <KeyInput
                  value={row.key}
                  readOnly={row.keyReadonly}
                  suggestions={row.keySuggestions}
                  placeholder={row.keyReadonly ? "参数名" : "键"}
                  onChange={(value) => row.onChangeKey?.(value)}
                />
              </td>
              <td className="p-0">
                <Input
                  value={row.value}
                  placeholder={row.valuePlaceholder ?? "值"}
                  className="h-8 rounded-none border-0 focus-visible:ring-0"
                  onChange={(event) => row.onChangeValue?.(event.target.value)}
                />
              </td>
              <td className="w-10 border-l p-0 text-center">
                {row.onRemove ? (
                  <button
                    type="button"
                    className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive mx-auto my-1 flex size-6 items-center justify-center rounded transition-colors"
                    aria-label="移除"
                    onClick={row.onRemove}
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                ) : null}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {rows.length === 0 && (
        <div className="text-muted-foreground px-2 py-2 text-xs">无</div>
      )}
    </div>
  )
}

function KeyInput({
  value,
  readOnly,
  suggestions,
  placeholder,
  onChange,
}: {
  value: string
  readOnly?: boolean
  suggestions?: string[]
  placeholder?: string
  onChange: (value: string) => void
}) {
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(0)
  const [pos, setPos] = useState({ top: 0, left: 0, width: 0 })
  const wrapRef = useRef<HTMLDivElement | null>(null)

  const filtered = useMemo(() => {
    if (!suggestions?.length) return []
    const query = value.trim().toLowerCase()
    if (!query) return suggestions
    return suggestions.filter((name) => name.toLowerCase().includes(query))
  }, [suggestions, value])

  useEffect(() => {
    const handler = (event: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handler)
    return () => document.removeEventListener("mousedown", handler)
  }, [])

  const updatePos = () => {
    const rect = wrapRef.current?.getBoundingClientRect()
    if (rect) setPos({ top: rect.bottom + 4, left: rect.left, width: rect.width })
  }

  const commit = (name: string) => {
    onChange(name)
    setOpen(false)
  }

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (!open) {
      if (suggestions?.length) setOpen(true)
      return
    }
    if (event.key === "ArrowDown") {
      event.preventDefault()
      setActiveIndex((index) => Math.min(index + 1, filtered.length - 1))
    } else if (event.key === "ArrowUp") {
      event.preventDefault()
      setActiveIndex((index) => Math.max(index - 1, 0))
    } else if (event.key === "Enter") {
      event.preventDefault()
      if (filtered[activeIndex]) commit(filtered[activeIndex])
    } else if (event.key === "Escape") {
      setOpen(false)
    }
  }

  return (
    <div ref={wrapRef} className="relative">
      <Input
        value={value}
        readOnly={readOnly}
        placeholder={placeholder}
        className="h-8 rounded-none border-0 focus-visible:ring-0"
        onFocus={() => {
          updatePos()
          setOpen(true)
        }}
        onKeyDown={handleKeyDown}
        onChange={(event) => {
          onChange(event.target.value)
          setActiveIndex(0)
          setOpen(true)
        }}
      />
      {open &&
        filtered.length > 0 &&
        createPortal(
          <div
            className="bg-popover text-popover-foreground fixed z-50 max-h-48 overflow-y-auto rounded-md border p-1 shadow-md"
            style={{ top: pos.top, left: pos.left, width: pos.width }}
          >
            {filtered.map((name, index) => (
              <button
                key={name}
                type="button"
                className={cn(
                  "text-accent-foreground block w-full rounded-sm px-2 py-1 text-left text-sm transition-colors",
                  index === activeIndex ? "bg-accent" : "hover:bg-accent/60"
                )}
                onMouseDown={(event) => {
                  event.preventDefault()
                  commit(name)
                }}
                onMouseEnter={() => setActiveIndex(index)}
              >
                {name}
              </button>
            ))}
          </div>,
          document.body
        )}
    </div>
  )
}

function methodClass(method: string): string {
  if (method === "GET") return "text-green-600 dark:text-green-400"
  if (method === "POST") return "text-blue-600 dark:text-blue-400"
  if (method === "PUT" || method === "PATCH") return "text-orange-500"
  if (method === "DELETE") return "text-destructive"

  return "text-muted-foreground"
}

function RequestBodyModeButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      className={cn(
        "rounded border px-2 py-0.5 text-xs transition-colors",
        active
          ? "border-primary bg-primary/10 text-foreground font-medium"
          : "text-muted-foreground hover:bg-accent/50"
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}

function RequestBodyEditor({
  mode,
  value,
  onChange,
}: {
  mode: "json" | "text"
  value: string
  onChange: (value: string) => void
}) {
  const dark = useIsDark()
  const { font } = useTheme()

  return (
    <Editor
      path={`request-body.${mode === "json" ? "json" : "txt"}`}
      language={mode === "json" ? "json" : "plaintext"}
      value={value}
      theme={dark ? "near-black" : "light"}
      beforeMount={(monaco) => {
        monaco.editor.defineTheme("near-black", {
          base: "vs-dark",
          inherit: true,
          rules: [],
          colors: {
            "editor.background": "#000000",
          },
        })
      }}
      options={{
        minimap: { enabled: false },
        fontSize: 13,
        fontFamily: getMonacoFontFamily(font),
        scrollBeyondLastLine: false,
        automaticLayout: true,
        wordWrap: "on",
        tabSize: 2,
      }}
      height="200px"
      onChange={(next) => onChange(next ?? "")}
    />
  )
}

function RequestTab({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      className={cn(
        "rounded-md border px-2.5 py-1 text-xs transition-colors",
        active
          ? "border-transparent bg-primary/10 text-foreground font-medium"
          : "text-muted-foreground border-transparent hover:bg-accent/50 hover:text-foreground"
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}