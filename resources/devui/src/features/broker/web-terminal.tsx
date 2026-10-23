import { useEffect, useRef } from "react"
import { Terminal, type ITheme } from "@xterm/xterm"
import { FitAddon } from "@xterm/addon-fit"
import { UnicodeGraphemesAddon } from "@xterm/addon-unicode-graphemes"
import { toast } from "@/lib/toast"

import "@xterm/xterm/css/xterm.css"

import { getToken } from "@/lib/http"
import { useIsDark } from "@/features/zip-file/file-icon"

const DARK_THEME: ITheme = {
  background: "#000000",
  foreground: "#e0e0e0",
  cursor: "#e0e0e0",
  selectionBackground: "#3a3a3a",
}

const LIGHT_THEME: ITheme = {
  background: "#f9f7f5",
  foreground: "#1f1f1f",
  cursor: "#1f1f1f",
  selectionBackground: "#c8d6e5",
}

export function WebTerminal({ slug }: { slug: string }) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const termRef = useRef<Terminal | null>(null)
  const dark = useIsDark()

  useEffect(() => {
    if (termRef.current) {
      termRef.current.options.theme = dark ? DARK_THEME : LIGHT_THEME
    }
  }, [dark])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const term = new Terminal({
      cursorBlink: true,
      allowProposedApi: true,
      fontFamily:
        'SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
      fontSize: 13,
      theme: document.documentElement.classList.contains("dark")
        ? DARK_THEME
        : LIGHT_THEME,
    })
    termRef.current = term
    const fit = new FitAddon()
    const graphemes = new UnicodeGraphemesAddon()
    term.loadAddon(fit)
    term.loadAddon(graphemes)
    term.open(container)
    fit.fit()

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
    const token = getToken() ?? ""
    const url = `${protocol}//${window.location.host}/api/v3/br/${slug}/api/v3/system/pty?authorization=${encodeURIComponent(token)}`
    let ws: WebSocket | null = null
    let disposed = false

    try {
      ws = new WebSocket(url)
    } catch {
      if (!disposed) toast.error("终端连接失败")
      term.dispose()
      return
    }

    let resized = false
    let heartbeat: ReturnType<typeof setInterval> | null = null
    const stopHeartbeat = () => {
      if (heartbeat !== null) {
        clearInterval(heartbeat)
        heartbeat = null
      }
    }
    const sendResize = () => {
      if (!ws || ws.readyState !== WebSocket.OPEN) return
      const { cols, rows } = term
      ws.send(JSON.stringify({ type: "resize", data: { cols, rows } }))
    }

    ws.onopen = () => {
      if (disposed) return
      sendResize()
      resized = true
      stopHeartbeat()
      heartbeat = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: "ping", data: "" }))
        }
      }, 10_000)
    }

    ws.onmessage = (event: MessageEvent) => {
      if (disposed) return
      try {
        const msg = JSON.parse(event.data as string) as { type?: string; data?: string }
        if (msg.type === "stdout") {
          term.write(typeof msg.data === "string" ? msg.data : "")
        }
      } catch {
        // ignore non-JSON frames
      }
    }

    ws.onerror = () => {
      if (disposed) return
      toast.error("终端连接异常")
    }

    ws.onclose = () => {
      stopHeartbeat()
      if (disposed) return
      term.write("\r\n\x1b[90m[连接已断开]\x1b[0m\r\n")
    }

    const dataSub = term.onData((data) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "stdin", data }))
      }
    })

    const resizeSub = term.onResize(() => {
      if (resized) sendResize()
    })

    const handleWindowResize = () => {
      fit.fit()
      if (resized) sendResize()
    }
    window.addEventListener("resize", handleWindowResize)

    return () => {
      disposed = true
      stopHeartbeat()
      window.removeEventListener("resize", handleWindowResize)
      dataSub.dispose()
      resizeSub.dispose()
      if (ws) {
        ws.onopen = null
        ws.onmessage = null
        ws.onerror = null
        ws.onclose = null
        ws.close()
      }
      term.dispose()
      termRef.current = null
    }
  }, [slug])

  return (
    <div className="h-full w-full overflow-hidden rounded-md border bg-background p-2">
      <div ref={containerRef} className="h-full w-full" />
    </div>
  )
}