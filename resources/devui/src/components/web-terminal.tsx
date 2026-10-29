import { useCallback, useEffect, useRef } from "react"
import { Terminal, type ITheme } from "@xterm/xterm"
import { FitAddon } from "@xterm/addon-fit"
import { UnicodeGraphemesAddon } from "@xterm/addon-unicode-graphemes"
import { PictureInPicture2 } from "lucide-react"
import { toast } from "@/lib/toast"

import "@xterm/xterm/css/xterm.css"

import { TooltipHint } from "@/components/tooltip-hint"
import { getToken } from "@/lib/http"
import { usePictureInPicture } from "@/lib/use-picture-in-picture"
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

export function WebTerminal({ path, title }: { path: string; title?: string }) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const dark = useIsDark()

  const sendResize = useCallback(() => {
    const term = termRef.current
    const ws = wsRef.current
    if (!term || !ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type: "resize", data: { cols: term.cols, rows: term.rows } }))
  }, [])

  const pip = usePictureInPicture({
    containerRef,
    onResize: () => {
      fitRef.current?.fit()
      sendResize()
    },
  })

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
    fitRef.current = fit
    const graphemes = new UnicodeGraphemesAddon()
    term.loadAddon(fit)
    term.loadAddon(graphemes)
    term.open(container)
    fit.fit()

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
    const token = getToken() ?? ""
    const url = `${protocol}//${window.location.host}${path}?authorization=${encodeURIComponent(token)}`
    let ws: WebSocket | null = null
    let disposed = false

    try {
      ws = new WebSocket(url)
      wsRef.current = ws
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
        const data = typeof msg.data === "string" ? msg.data : ""
        if (msg.type === "stdout") {
          term.write(data)
        } else if (msg.type === "error" && data) {
          term.write(`\r\n\x1b[31m[错误] ${data}\x1b[0m\r\n`)
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

    const activeDocument = () => containerRef.current?.ownerDocument ?? document

    const notify = (message: string) => {
      if (activeDocument() !== document) {
        term.write(`\r\n\x1b[31m[${message}]\x1b[0m\r\n`)
      } else {
        toast.error(message)
      }
    }

    const fallbackCopy = (text: string, doc: Document) => {
      const area = doc.createElement("textarea")
      area.value = text
      area.style.position = "fixed"
      area.style.left = "-9999px"
      doc.body.appendChild(area)
      area.select()
      const copied = doc.execCommand("copy")
      area.remove()
      term.focus()
      if (!copied) throw new Error("copy failed")
    }

    const copySelection = async () => {
      const text = term.getSelection()
      if (!text) return
      const doc = activeDocument()
      const win = doc.defaultView ?? window
      const clipboard = win.navigator.clipboard ?? navigator.clipboard
      try {
        if (clipboard?.writeText) {
          await clipboard.writeText(text)
        } else {
          fallbackCopy(text, doc)
        }
        term.clearSelection()
        return
      } catch {
        try {
          fallbackCopy(text, doc)
          term.clearSelection()
        } catch {
          notify("复制失败，请手动复制")
        }
      }
    }

    const pasteFromClipboard = async () => {
      const doc = activeDocument()
      const win = doc.defaultView ?? window
      const clipboard = win.navigator.clipboard ?? navigator.clipboard
      if (!clipboard?.readText) {
        notify("当前浏览器不支持读取剪贴板")
        return
      }
      try {
        const text = await clipboard.readText()
        if (text) term.paste(text)
      } catch {
        notify("无法读取剪贴板")
      }
    }

    term.attachCustomKeyEventHandler((event) => {
      if (
        event.type !== "keydown" ||
        event.altKey ||
        !(event.ctrlKey || event.metaKey)
      ) {
        return true
      }

      const key = event.key.toLowerCase()
      if (key === "c") {
        if (!term.hasSelection()) return true
        event.preventDefault()
        void copySelection()
        return false
      }
      if (key === "v") {
        event.preventDefault()
        if (!event.repeat) void pasteFromClipboard()
        return false
      }

      return true
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
        wsRef.current = null
      }
      term.dispose()
      termRef.current = null
      fitRef.current = null
    }
  }, [path, sendResize])

  return (
    <div className="relative h-full w-full overflow-hidden rounded-md border bg-background p-2">
      {pip.supported && !pip.active && (
        <TooltipHint content="画中画打开">
          <button
            type="button"
            aria-label="画中画打开"
            className="text-muted-foreground hover:bg-accent hover:text-accent-foreground absolute top-3 right-3 z-10 flex size-6 items-center justify-center rounded transition-colors"
            onClick={() => void pip.open(title)}
          >
            <PictureInPicture2 className="size-3.5" />
          </button>
        </TooltipHint>
      )}
      <div ref={containerRef} className="h-full w-full" />
    </div>
  )
}