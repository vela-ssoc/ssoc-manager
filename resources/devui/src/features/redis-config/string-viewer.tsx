import { useMemo, useState } from "react"
import Editor from "@monaco-editor/react"
import { decode } from "@msgpack/msgpack"

import { useIsDark } from "@/features/zip-file/file-icon"
import { useTheme } from "@/providers/theme-provider"
import { getMonacoFontFamily } from "@/lib/fonts"
import { cn } from "@/lib/utils"

type Format = "auto" | "json" | "markdown" | "hex" | "binary" | "msgpack"

const FORMATS: Array<{ value: Format; label: string }> = [
  { value: "auto", label: "自动" },
  { value: "json", label: "JSON" },
  { value: "markdown", label: "Markdown" },
  { value: "msgpack", label: "Msgpack" },
  { value: "hex", label: "Hex" },
  { value: "binary", label: "Binary" },
]

const LANGUAGE: Record<Format, string> = {
  auto: "plaintext",
  json: "json",
  markdown: "markdown",
  hex: "plaintext",
  binary: "plaintext",
  msgpack: "json",
}

interface Render {
  text: string
  language: string
  error?: string
  hint?: string
}

export function RedisStringViewer({ text }: { text: string }) {
  const [format, setFormat] = useState<Format>("auto")
  const dark = useIsDark()
  const { font } = useTheme()

  const bytes = useMemo(() => new TextEncoder().encode(text), [text])
  const render = useMemo(() => renderValue(text, bytes, format), [text, bytes, format])

  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b px-3 py-1.5">
        <div className="bg-muted flex items-center gap-0.5 rounded-md p-0.5">
          {FORMATS.map((item) => (
            <button
              key={item.value}
              type="button"
              aria-pressed={format === item.value}
              className={cn(
                "rounded px-1.5 py-0.5 text-[11px] font-medium transition-colors",
                format === item.value
                  ? "bg-card text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
              onClick={() => setFormat(item.value)}
            >
              {item.label}
            </button>
          ))}
        </div>
        <span className="text-muted-foreground ml-auto text-[11px] tabular-nums">
          {text.length === 0 ? "空字符串" : `${text.length} 字符 · ${bytes.length} 字节`}
        </span>
      </div>

      {render.error && (
        <div className="border-b bg-amber-500/10 px-3 py-1.5 text-[11px] text-amber-600 dark:text-amber-400">
          {render.error}
        </div>
      )}

      {render.hint && (
        <div className="bg-muted/40 border-b px-3 py-1.5 text-[11px] text-muted-foreground">
          {render.hint}
        </div>
      )}

      <div className="h-[360px]">
        <Editor
          height="100%"
          language={render.language}
          value={render.text}
          theme={dark ? "vs-dark" : "light"}
          beforeMount={(monaco) => {
            monaco.editor.defineTheme("near-black", {
              base: "vs-dark",
              inherit: true,
              rules: [],
              colors: { "editor.background": "#000000" },
            })
          }}
          onMount={(_, monaco) => {
            monaco.editor.setTheme(dark ? "near-black" : "light")
          }}
          options={{
            minimap: { enabled: false },
            fontSize: 12,
            fontFamily: getMonacoFontFamily(font),
            scrollBeyondLastLine: false,
            automaticLayout: true,
            wordWrap: "on",
            readOnly: true,
            renderLineHighlight: "none",
            contextmenu: false,
            scrollbar: { verticalScrollbarSize: 8, horizontalScrollbarSize: 8 },
          }}
        />
      </div>
    </div>
  )
}

function renderValue(text: string, bytes: Uint8Array, format: Format): Render {
  if (format === "auto") {
    const sniffed = sniffSerialized(bytes)
    if (sniffed) {
      return {
        text: hexDump(bytes),
        language: LANGUAGE.hex,
        hint: sniffed,
      }
    }
  }

  const target = format === "auto" ? detectFormat(text) : format

  switch (target) {
    case "json":
      return renderJson(text)
    case "markdown":
      return { text, language: LANGUAGE.markdown }
    case "hex":
      return { text: hexDump(bytes), language: LANGUAGE.hex }
    case "binary":
      return { text: binaryDump(bytes), language: LANGUAGE.binary }
    case "msgpack":
      return renderMsgpack(bytes)
    default:
      return { text, language: LANGUAGE.auto }
  }
}

function sniffSerialized(bytes: Uint8Array): string | undefined {
  if (
    bytes.length >= 4 &&
    bytes[0] === 0xac &&
    bytes[1] === 0xed &&
    bytes[2] === 0x00 &&
    bytes[3] === 0x05
  ) {
    return "检测到 Java 序列化数据（JDK ObjectOutputStream），暂不支持解析，已展示为 Hex 视图"
  }
  if (bytes.length >= 2 && bytes[0] === 0x80 && bytes[1] >= 0x02 && bytes[1] <= 0x05) {
    return "检测到 Python pickle 数据，暂不支持解析，已展示为 Hex 视图"
  }
  if (bytes.length >= 2 && bytes[0] === 0x80 && bytes[1] === 0x01) {
    return "检测到 Python pickle 数据，暂不支持解析，已展示为 Hex 视图"
  }
  return undefined
}

function detectFormat(text: string): Format {
  const head = text.trimStart()
  if (head.startsWith("{") || head.startsWith("[")) {
    try {
      JSON.parse(head)
      return "json"
    } catch {
      // 继续尝试其它格式
    }
  }
  if (/^#{1,6}\s/m.test(head) || head.includes("```") || /^[-*]\s/m.test(head)) {
    return "markdown"
  }
  return "auto"
}

function renderJson(text: string): Render {
  try {
    const parsed = JSON.parse(text)
    return { text: JSON.stringify(parsed, null, 2), language: LANGUAGE.json }
  } catch (err) {
    return {
      text,
      language: LANGUAGE.json,
      error: `无法按 JSON 解析：${err instanceof Error ? err.message : "格式错误"}`,
    }
  }
}

function renderMsgpack(bytes: Uint8Array): Render {
  try {
    const parsed = decode(bytes)
    return {
      text: JSON.stringify(parsed, replacer, 2),
      language: LANGUAGE.json,
    }
  } catch {
    return {
      text: hexDump(bytes),
      language: LANGUAGE.hex,
      error: "无法按 Msgpack 解析，已回退为 Hex 视图",
    }
  }
}

function replacer(_key: string, value: unknown): unknown {
  if (typeof value === "bigint") return value.toString()
  if (value instanceof Map) return Object.fromEntries(value)
  if (value instanceof Set) return [...value]
  if (value instanceof Uint8Array) return Array.from(value)
  return value
}

function hexDump(bytes: Uint8Array): string {
  if (bytes.length === 0) return ""
  const lines: string[] = []
  for (let offset = 0; offset < bytes.length; offset += 16) {
    const slice = bytes.subarray(offset, offset + 16)
    const hex = Array.from(slice)
      .map((byte) => byte.toString(16).padStart(2, "0"))
      .join(" ")
      .padEnd(47, " ")
    const ascii = Array.from(slice)
      .map((byte) => (byte >= 0x20 && byte < 0x7f ? String.fromCharCode(byte) : "."))
      .join("")
    lines.push(`${offset.toString(16).padStart(8, "0")}  ${hex}  |${ascii}|`)
  }
  return lines.join("\n")
}

function binaryDump(bytes: Uint8Array): string {
  if (bytes.length === 0) return ""
  const lines: string[] = []
  for (let offset = 0; offset < bytes.length; offset += 8) {
    const slice = bytes.subarray(offset, offset + 8)
    const bits = Array.from(slice)
      .map((byte) => byte.toString(2).padStart(8, "0"))
      .join(" ")
    lines.push(`${offset.toString(16).padStart(8, "0")}  ${bits}`)
  }
  return lines.join("\n")
}