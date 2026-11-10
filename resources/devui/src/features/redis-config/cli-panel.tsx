import { useEffect, useRef } from "react"
import { Terminal, type ITheme } from "@xterm/xterm"
import { FitAddon } from "@xterm/addon-fit"
import { PictureInPicture2 } from "lucide-react"

import "@xterm/xterm/css/xterm.css"

import { TooltipHint } from "@/components/tooltip-hint"
import { ApiError } from "@/lib/http"
import { usePictureInPicture } from "@/lib/use-picture-in-picture"
import { useTheme } from "@/providers/theme-provider"
import { redisConfigApi } from "./api"

const MAX_ARGS = 20

function promptOf(endpoint: string | undefined, db: number, cluster: boolean): string {
  const host = (endpoint ?? "").trim() || "redis"
  const suffix = `${db > 0 ? `[${db}]` : ""}${cluster ? "[cluster]" : ""}`
  return `\x1b[32m${host}${suffix}>\x1b[0m `
}

const CONFIRM_PROMPT =
  "\x1b[33m确认执行？\x1b[0m \x1b[90m(\x1b[0m\x1b[1;32my\x1b[0m\x1b[90m/N)\x1b[0m "

const DARK_THEME: ITheme = {
  background: "#000000",
  foreground: "#e0e0e0",
  cursor: "#e0e0e0",
  selectionBackground: "#3a3a3a",
  black: "#1f1f1f",
  red: "#e06c75",
  green: "#98c379",
  yellow: "#e5c07b",
  blue: "#61afef",
  magenta: "#c678dd",
  cyan: "#56b6c2",
  white: "#dcdfe4",
  brightBlack: "#8b93a1",
  brightRed: "#ff7b86",
  brightGreen: "#b3e09a",
  brightYellow: "#f5d38f",
  brightBlue: "#7cc4ff",
  brightMagenta: "#d8a0ee",
  brightCyan: "#7fd6e2",
  brightWhite: "#ffffff",
}

const LIGHT_THEME: ITheme = {
  background: "#ffffff",
  foreground: "#1f1f1f",
  cursor: "#1f1f1f",
  selectionBackground: "#c8d6e5",
  black: "#1f1f1f",
  red: "#b91c1c",
  green: "#15803d",
  yellow: "#8a6d00",
  blue: "#1d4ed8",
  magenta: "#7e22ce",
  cyan: "#0e7490",
  white: "#4b5563",
  brightBlack: "#6b7280",
  brightRed: "#dc2626",
  brightGreen: "#16a34a",
  brightYellow: "#a16207",
  brightBlue: "#2563eb",
  brightMagenta: "#9333ea",
  brightCyan: "#0891b2",
  brightWhite: "#111827",
}

const RULE =
  "\x1b[90m────────────────────────────────────────────────────────────\x1b[0m"

function bannerOf(target: string): string {
  return [
    RULE,
    "\x1b[1;36mSSOC Redis CLI\x1b[0m  \x1b[90m·  交互式命令终端\x1b[0m",
    `\x1b[90m目标实例\x1b[0m \x1b[36m${target}\x1b[0m`,
    "\x1b[90m输入命令回车执行 · ↑↓ 翻阅历史 · Ctrl+C 复制选区 / 中断 · Ctrl+L 清屏\x1b[0m",
    "",
    "\x1b[1;33m⚠ 命令行模式下可以执行所有的 Redis 命令\x1b[0m\x1b[33m，包括 FLUSHALL、FLUSHDB、DEL、",
    "  CONFIG SET 等高风险命令。执行后会直接影响线上数据与程序运行，请谨慎使用。\x1b[0m",
    RULE,
    "",
  ].join("\r\n")
}

export function RedisCliPanel({
  endpoint,
  db = 0,
  cluster = false,
  master,
}: {
  endpoint?: string
  db?: number
  cluster?: boolean
  master?: string
}) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const safeRef = useRef<Set<string>>(new Set())
  const { isDark: dark } = useTheme()

  const pip = usePictureInPicture({
    containerRef,
    onResize: () => fitRef.current?.fit(),
  })

  useEffect(() => {
    const repaint = () => {
      fitRef.current?.fit()
      const term = termRef.current
      if (term) term.refresh(0, Math.max(0, term.rows - 1))
    }

    repaint()

    const windows: Window[] = [window]
    const hostWindow = containerRef.current?.ownerDocument.defaultView ?? null
    if (hostWindow && hostWindow !== window) windows.push(hostWindow)

    const onVisible = () => repaint()
    for (const win of windows) {
      win.addEventListener("focus", onVisible)
      win.addEventListener("pageshow", onVisible)
      win.document.addEventListener("visibilitychange", onVisible)
    }

    const timer = window.setTimeout(repaint, 120)

    return () => {
      window.clearTimeout(timer)
      for (const win of windows) {
        win.removeEventListener("focus", onVisible)
        win.removeEventListener("pageshow", onVisible)
        win.document.removeEventListener("visibilitychange", onVisible)
      }
    }
  }, [pip.active])

  useEffect(() => {
    const term = termRef.current
    if (!term) return
    term.options.theme = dark ? DARK_THEME : LIGHT_THEME
    term.refresh(0, Math.max(0, term.rows - 1))
  }, [dark])

  useEffect(() => {
    let cancelled = false
    redisConfigApi
      .safeCommands()
      .then((res) => {
        if (cancelled) return
        const list = res?.data ?? []
        safeRef.current = new Set(list.map((item) => item.trim().toUpperCase()))
      })
      .catch(() => {
        // 拉取失败时保持空名单：所有命令都需要二次确认
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const term = new Terminal({
      cursorBlink: true,
      convertEol: true,
      fontSize: 12,
      lineHeight: 1.35,
      scrollback: 2000,
      fontFamily:
        'SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
      theme: document.documentElement.classList.contains("dark")
        ? DARK_THEME
        : LIGHT_THEME,
    })
    termRef.current = term

    const fit = new FitAddon()
    fitRef.current = fit
    term.loadAddon(fit)
    term.open(container)

    let line = ""
    let pos = 0
    const basePrompt = promptOf(endpoint, db, cluster)
    let prompt = basePrompt
    const history: string[] = []
    let historyIndex = -1
    let busy = false
    let disposed = false
    let waiting: string[] | null = null

    const redraw = () => {
      term.write(`\r\x1b[2K${prompt}${line}`)
      const back = line.length - pos
      if (back > 0) term.write(`\x1b[${back}D`)
    }

    const writeOut = (text: string) => {
      term.write(`\r\x1b[2K${text.replace(/\n/g, "\r\n")}\r\n`)
      redraw()
    }

    const recall = (step: number) => {
      if (history.length === 0) return
      const next =
        historyIndex < 0 ? (step < 0 ? history.length - 1 : -1) : historyIndex + step
      if (next < 0) {
        historyIndex = -1
        line = ""
        pos = 0
        redraw()
        return
      }
      if (next >= history.length) return
      historyIndex = next
      line = history[next]
      pos = line.length
      redraw()
    }

    const run = async (args: string[], privilege: boolean) => {
      busy = true
      try {
        const res = await redisConfigApi.do(args, privilege)
        if (disposed) return
        writeOut(formatRedisResult(res?.data))
      } catch (err) {
        if (disposed) return
        writeOut(formatCliError(err))
      } finally {
        busy = false
        if (!disposed) redraw()
      }
    }

    const cancelWait = (message = "已取消执行，命令未发送到服务端") => {
      waiting = null
      prompt = basePrompt
      writeOut(`\x1b[90m${message}\x1b[0m`)
    }

    const submit = async () => {
      const raw = line.trim()
      term.write("\r\n")
      line = ""
      pos = 0

      if (waiting) {
        const args = waiting
        const answer = raw.toLowerCase()
        const confirmed = answer === "y" || answer === "yes"
        waiting = null
        prompt = basePrompt
        if (confirmed) await run(args, true)
        else cancelWait()
        return
      }

      if (!raw) {
        redraw()
        return
      }

      history.push(raw)
      historyIndex = -1

      const args = splitArgs(raw)
      if (args.length === 0) {
        redraw()
        return
      }
      if (args.length > MAX_ARGS) {
        writeOut(`\x1b[31m参数过多，最多 ${MAX_ARGS} 个\x1b[0m`)
        return
      }

      const command = args[0].trim().toUpperCase()
      if (safeRef.current.has(command)) {
        await run(args, false)
        return
      }

      waiting = args
      prompt = CONFIRM_PROMPT
      term.write(
        `\r\x1b[2K\x1b[33m⚠ ${command} 存在数据修改风险，执行后可能影响程序状态。\x1b[0m\r\n`
      )
      redraw()
    }

    const dataSub = term.onData((data) => {
      if (disposed || busy) return

      switch (data) {
        case "\r":
          void submit()
          return
        case "\u007f":
          if (pos > 0) {
            line = line.slice(0, pos - 1) + line.slice(pos)
            pos -= 1
            redraw()
          }
          return
        case "\u0003":
          term.write("^C\r\n")
          line = ""
          pos = 0
          if (waiting) {
            cancelWait()
            return
          }
          redraw()
          return
        case "\u000c":
          term.clear()
          redraw()
          return
        case "\u0001":
          pos = 0
          redraw()
          return
        case "\u0005":
          pos = line.length
          redraw()
          return
        case "\u0015":
          line = line.slice(pos)
          pos = 0
          redraw()
          return
        case "\u001b[A":
          recall(-1)
          return
        case "\u001b[B":
          recall(1)
          return
        case "\u001b[C":
          if (pos < line.length) {
            pos += 1
            redraw()
          }
          return
        case "\u001b[D":
          if (pos > 0) {
            pos -= 1
            redraw()
          }
          return
        case "\u001b[H":
          pos = 0
          redraw()
          return
        case "\u001b[F":
          pos = line.length
          redraw()
          return
        case "\u001b[3~":
          if (pos < line.length) {
            line = line.slice(0, pos) + line.slice(pos + 1)
            redraw()
          }
          return
        default:
          break
      }

      const text = data.replace(/[\r\n\t]/g, " ")
      if (!text) return
      line = line.slice(0, pos) + text + line.slice(pos)
      pos += text.length
      redraw()
    })

    const observer = new ResizeObserver(() => {
      if (container.clientWidth === 0 || container.clientHeight === 0) return
      fit.fit()
    })
    observer.observe(container)

    const applyTheme = () => {
      const isDark = document.documentElement.classList.contains("dark")
      term.options.theme = isDark ? DARK_THEME : LIGHT_THEME
      term.refresh(0, Math.max(0, term.rows - 1))
    }
    applyTheme()
    const themeObserver = new MutationObserver(applyTheme)
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    })

    const activeWindow = () =>
      containerRef.current?.ownerDocument.defaultView ?? window

    const insertText = (value: string) => {
      const text = value.replace(/[\r\n\t]/g, " ")
      if (!text) return
      line = line.slice(0, pos) + text + line.slice(pos)
      pos += text.length
      redraw()
    }

    term.attachCustomKeyEventHandler((event) => {
      if (event.type !== "keydown") return true
      if (event.altKey || !(event.ctrlKey || event.metaKey)) return true

      const key = event.key.toLowerCase()

      if (key === "c") {
        if (!term.hasSelection()) return true
        event.preventDefault()
        const text = term.getSelection()
        const clipboard = activeWindow().navigator.clipboard ?? navigator.clipboard
        if (clipboard?.writeText) {
          void clipboard.writeText(text).catch(() => undefined)
        }
        term.clearSelection()
        return false
      }

      if (key === "v") {
        const clipboard = activeWindow().navigator.clipboard ?? navigator.clipboard
        if (!clipboard?.readText) return true

        event.preventDefault()
        void clipboard
          .readText()
          .then((text) => {
            if (disposed) return
            insertText(text ?? "")
          })
          .catch(() => {
            if (disposed) return
            term.write(
              "\r\n\x1b[90m无法读取剪贴板，请授权后重试或使用 Ctrl+Shift+V\x1b[0m\r\n"
            )
            redraw()
          })
        return false
      }

      return true
    })

    const target = (endpoint ?? "").trim() || "redis"
    const meta = [target]
    if (master?.trim()) meta.push(`哨兵 ${master.trim()}`)
    if (db > 0) meta.push(`DB ${db}`)
    if (cluster) meta.push("集群模式")
    const banner = bannerOf(meta.join("  ·  "))

    requestAnimationFrame(() => {
      fit.fit()
      term.write(banner)
      redraw()
      term.focus()
    })

    return () => {
      disposed = true
      observer.disconnect()
      themeObserver.disconnect()
      dataSub.dispose()
      term.dispose()
      termRef.current = null
      fitRef.current = null
    }
  }, [endpoint, db, cluster, master])

  return (
    <div className="relative h-full px-3 py-2">
      {pip.supported && !pip.active && (
        <TooltipHint content="画中画打开">
          <button
            type="button"
            aria-label="画中画打开"
            className="text-muted-foreground hover:bg-accent hover:text-accent-foreground absolute top-2 right-3 z-10 flex size-6 items-center justify-center rounded transition-colors"
            onClick={() => void pip.open("Redis CLI")}
          >
            <PictureInPicture2 className="size-3.5" />
          </button>
        </TooltipHint>
      )}
      <div ref={containerRef} className="h-full w-full" />
    </div>
  )
}

function formatCliError(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.kind === "network") {
      return `\x1b[31m(error) 网络不可用：${err.message}\x1b[0m`
    }
    if (err.kind === "aborted") {
      return "\x1b[31m(error) 请求已取消\x1b[0m"
    }

    const problem = err.detail
    const detail = problem?.detail?.trim()
    const title = problem?.title?.trim()
    const text = detail || title || err.message
    const status = problem?.status ?? err.status
    const suffix = status ? ` \x1b[90m[HTTP ${status}]\x1b[0m` : ""

    return `\x1b[31m(error) ${text}\x1b[0m${suffix}`
  }

  if (err instanceof Error) {
    return `\x1b[31m(error) ${err.message}\x1b[0m`
  }

  return "\x1b[31m(error) 执行失败\x1b[0m"
}

function splitArgs(input: string): string[] {
  const out: string[] = []
  let current = ""
  let quote: '"' | "'" | null = null
  let started = false

  for (let i = 0; i < input.length; i++) {
    const ch = input[i]
    if (quote) {
      if (ch === "\\" && i + 1 < input.length) {
        current += input[++i]
        continue
      }
      if (ch === quote) {
        quote = null
        continue
      }
      current += ch
      continue
    }
    if (ch === '"' || ch === "'") {
      quote = ch
      started = true
      continue
    }
    if (ch === "\\" && i + 1 < input.length) {
      current += input[++i]
      started = true
      continue
    }
    if (/\s/.test(ch)) {
      if (started) {
        out.push(current)
        current = ""
        started = false
      }
      continue
    }
    current += ch
    started = true
  }
  if (started) out.push(current)

  return out
}

function formatRedisResult(value: unknown, depth = 0): string {
  if (value === null || value === undefined) return "(nil)"
  if (typeof value === "number") return `(integer) ${value}`
  if (typeof value === "boolean") return value ? "(integer) 1" : "(integer) 0"
  if (typeof value === "string") return quote(value)

  if (Array.isArray(value)) {
    if (value.length === 0) return "(empty array)"
    const pad = "   ".repeat(depth)
    return value
      .map((item, index) => `${pad}${index + 1}) ${formatRedisResult(item, depth + 1)}`)
      .join("\n")
  }

  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function quote(value: string): string {
  if (value === "") return '""'
  const escaped = value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')
  return `"${escaped}"`
}