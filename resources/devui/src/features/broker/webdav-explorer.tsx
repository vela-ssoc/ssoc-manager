import { useEffect, useMemo, useState } from "react"
import Editor from "@monaco-editor/react"
import { toast } from "@/lib/toast"
import {
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ChevronRight,
  Download,
  FileText,
  FileWarning,
  Folder,
  Loader2,
  RefreshCw,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime } from "@/lib/format"
import { getMonacoFontFamily } from "@/lib/fonts"
import { cn } from "@/lib/utils"
import { useIsDark } from "@/features/zip-file/file-icon"
import { useTheme } from "@/providers/theme-provider"
import {
  fileExtension,
  formatFileSize,
  isImageFile,
  isTextFile,
} from "@/features/zip-file/types"

import {
  davDownloadUrl,
  davList,
  splitPath,
  type DavEndpoint,
  type DavEntry,
} from "./webdav"

const MAX_PREVIEW_SIZE = 1024 * 1024

const LANGUAGE_BY_EXT: Record<string, string> = {
  js: "javascript", mjs: "javascript", cjs: "javascript", jsx: "javascript",
  ts: "typescript", tsx: "typescript",
  json: "json", jsonc: "json", json5: "json",
  html: "html", htm: "html", vue: "html", svelte: "html",
  css: "css", scss: "scss", less: "less",
  md: "markdown", markdown: "markdown",
  yaml: "yaml", yml: "yaml",
  xml: "xml", svg: "xml",
  sh: "shell", bash: "shell", zsh: "shell",
  py: "python", rb: "ruby", php: "php",
  sql: "sql", java: "java", c: "c", h: "c", cpp: "cpp", hpp: "cpp",
  cs: "csharp", rs: "rust", go: "go", kt: "kotlin", kts: "kotlin",
  swift: "swift", scala: "scala", lua: "lua", r: "r", dart: "dart",
  toml: "ini", ini: "ini", conf: "ini", properties: "ini",
  proto: "proto", graphql: "graphql", gql: "graphql",
}

export function WebDavExplorer({
  clientUrl,
  apiPath,
}: {
  clientUrl: string
  apiPath: string
}) {
  const endpoint: DavEndpoint = { clientUrl, apiPath }
  const [path, setPath] = useState("")
  const [backStack, setBackStack] = useState<string[]>([])
  const [forwardStack, setForwardStack] = useState<string[]>([])
  const [entries, setEntries] = useState<DavEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [preview, setPreview] = useState<{ stat: DavEntry; path: string } | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    davList(endpoint, path)
      .then((list) => {
        if (!cancelled) setEntries(list)
      })
      .catch((err) => {
        if (!cancelled) toast.error(err instanceof Error ? err.message : "加载目录失败")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clientUrl, apiPath, path, refreshKey])

  const navigate = (dir: string) => {
    if (dir === path) return
    setBackStack((stack) => [...stack, path])
    setForwardStack([])
    setPreview(null)
    setPath(dir)
  }

  const goBack = () => {
    if (backStack.length === 0) return
    const target = backStack[backStack.length - 1]
    setBackStack((stack) => stack.slice(0, -1))
    setForwardStack((stack) => [path, ...stack])
    setPreview(null)
    setPath(target)
  }

  const goForward = () => {
    if (forwardStack.length === 0) return
    const target = forwardStack[0]
    setForwardStack((stack) => stack.slice(1))
    setBackStack((stack) => [...stack, path])
    setPreview(null)
    setPath(target)
  }

  const goUp = () => {
    const parent = parentPath(path)
    if (parent === path) return
    navigate(parent)
  }

  const crumbs = useMemo(() => splitPath(path), [path])

  const sorted = useMemo(() => sortedEntries(entries), [entries])

  const openItem = (stat: DavEntry) => {
    if (stat.isDir) {
      navigate(stat.path)
      return
    }
    setPreview({ stat, path: stat.path })
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
      <div className="flex items-center gap-1">
        <TooltipHint content="后退">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="后退"
            disabled={backStack.length === 0}
            onClick={goBack}
          >
            <ArrowLeft className="size-3.5" />
          </Button>
        </TooltipHint>
        <TooltipHint content="前进">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="前进"
            disabled={forwardStack.length === 0}
            onClick={goForward}
          >
            <ArrowRight className="size-3.5" />
          </Button>
        </TooltipHint>
        <TooltipHint content="上级目录">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="上级目录"
            disabled={!path}
            onClick={goUp}
          >
            <ArrowUp className="size-3.5" />
          </Button>
        </TooltipHint>
        <TooltipHint content="刷新">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="刷新"
            disabled={loading}
            onClick={() => setRefreshKey((key) => key + 1)}
          >
            <RefreshCw className={cn("size-3.5", loading && "animate-spin")} />
          </Button>
        </TooltipHint>

        <nav className="text-muted-foreground ml-1 flex min-w-0 items-center gap-1 text-xs">
          <button
            type="button"
            className="hover:text-foreground shrink-0 rounded px-1 py-0.5"
            onClick={() => navigate("")}
          >
            根目录
          </button>
          {crumbs.map((crumb) => (
            <span key={crumb.path} className="flex min-w-0 items-center gap-1">
              <ChevronRight className="size-3 shrink-0" />
              <button
                type="button"
                className="hover:text-foreground truncate rounded px-1 py-0.5"
                onClick={() => navigate(crumb.path)}
              >
                {crumb.name}
              </button>
            </span>
          ))}
        </nav>
      </div>

      <div className="flex min-h-0 flex-1 overflow-hidden rounded-md border">
        <div className="flex min-w-0 flex-[3] flex-col">
          <ListHeader />
          <div className="min-h-0 flex-1 overflow-auto">
            {loading && sorted.length === 0 ? (
              <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
                <Loader2 className="size-4 animate-spin" />
              </div>
            ) : sorted.length === 0 ? (
              <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
                空目录
              </div>
            ) : (
              <div className="grid auto-rows-min">
                {sorted.map((stat) => (
                  <ListRow
                    key={stat.path}
                    stat={stat}
                    endpoint={endpoint}
                    onClick={() => openItem(stat)}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="min-w-0 flex-[2] border-l">
          <PreviewPane endpoint={endpoint} preview={preview} />
        </div>
      </div>
    </div>
  )
}

function ListHeader() {
  return (
    <div className="bg-muted/40 text-muted-foreground flex items-center gap-2 border-b px-3 py-1.5 text-xs font-medium">
      <span className="min-w-0 flex-1">名称</span>
      <span className="w-20 shrink-0 text-right">大小</span>
      <span className="w-36 shrink-0 text-right">修改时间</span>
    </div>
  )
}

function ListRow({
  stat,
  endpoint,
  onClick,
}: {
  stat: DavEntry
  endpoint: DavEndpoint
  onClick: () => void
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      className="hover:bg-accent/40 group flex cursor-pointer items-center gap-2 border-b px-3 py-1.5 text-xs last:border-b-0"
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter") onClick()
      }}
    >
      {stat.isDir ? (
        <Folder className="text-primary size-4 shrink-0" />
      ) : (
        <FileText className="text-muted-foreground size-4 shrink-0" />
      )}
      <span className="min-w-0 flex-1 truncate">{stat.name}</span>
      <span className="text-muted-foreground w-20 shrink-0 text-right font-mono">
        {stat.isDir ? "—" : formatFileSize(stat.size)}
      </span>
      <span className="text-muted-foreground w-36 shrink-0 text-right">
        {stat.lastmod ? formatDateTime(stat.lastmod) : "—"}
      </span>
      <span className="flex w-6 shrink-0 justify-end">
        {!stat.isDir && (
          <TooltipHint content="下载">
            <a
              href={davDownloadUrl(endpoint, stat.path)}
              download={stat.name}
              aria-label="下载"
              className="hover:bg-accent hover:text-accent-foreground flex size-6 items-center justify-center rounded opacity-0 transition-opacity group-hover:opacity-100"
              onClick={(e) => e.stopPropagation()}
            >
              <Download className="size-3.5" />
            </a>
          </TooltipHint>
        )}
      </span>
    </div>
  )
}

function PreviewPane({
  endpoint,
  preview,
}: {
  endpoint: DavEndpoint
  preview: { stat: DavEntry; path: string } | null
}) {
  if (!preview) {
    return (
      <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
        选择文件预览
      </div>
    )
  }

  const { stat, path } = preview
  const name = stat.name
  const image = isImageFile(name)
  const text = isTextFile(name)

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between gap-2 border-b px-2 py-1">
        <span className="text-muted-foreground min-w-0 truncate text-xs">
          {name} · {formatFileSize(stat.size)}
        </span>
        <a
          href={davDownloadUrl(endpoint, path)}
          download={name}
          className="hover:bg-accent hover:text-accent-foreground flex size-6 items-center justify-center rounded"
          aria-label="下载"
        >
          <Download className="size-3.5" />
        </a>
      </div>

      <div className="min-h-0 flex-1 overflow-hidden">
        {stat.size > MAX_PREVIEW_SIZE
          ? (
            <Placeholder
              icon={FileWarning}
              title="文件过大，暂不支持在线预览"
              detail={`${formatFileSize(stat.size)}（上限 1 MB）`}
            />
          )
          : image
            ? (
              <div className="flex h-full items-center justify-center overflow-auto p-4">
                <img
                  src={davDownloadUrl(endpoint, path)}
                  alt={name}
                  className="max-h-full max-w-full object-contain"
                />
              </div>
            )
            : text
              ? <TextEditor endpoint={endpoint} path={path} name={name} />
              : (
                <Placeholder
                  icon={FileWarning}
                  title="该类型文件不支持预览"
                  detail={formatFileSize(stat.size)}
                />
              )}
      </div>
    </div>
  )
}

function Placeholder({
  icon: Icon,
  title,
  detail,
}: {
  icon: typeof FileWarning
  title: string
  detail?: string
}) {
  return (
    <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2 px-4 text-center">
      <Icon className="size-8 opacity-40" />
      <span className="text-sm">{title}</span>
      {detail && <span className="text-xs">{detail}</span>}
    </div>
  )
}

function TextEditor({
  endpoint,
  path,
  name,
}: {
  endpoint: DavEndpoint
  path: string
  name: string
}) {
  const dark = useIsDark()
  const { font } = useTheme()
  const [content, setContent] = useState("")
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetch(davDownloadUrl(endpoint, path))
      .then((res) => {
        if (!res.ok) throw new Error(`请求失败（${res.status}）`)
        return res.text()
      })
      .then((text) => {
        if (!cancelled) setContent(text)
      })
      .catch((err) => {
        if (!cancelled) toast.error(err instanceof Error ? err.message : "加载文件失败")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [endpoint.clientUrl, endpoint.apiPath, path])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="text-muted-foreground size-4 animate-spin" />
      </div>
    )
  }

  const ext = fileExtension(name)
  const language = LANGUAGE_BY_EXT[ext] ?? "plaintext"

  return (
    <Editor
      path={name}
      language={language}
      value={content}
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

function parentPath(path: string): string {
  if (!path) return ""
  const idx = path.lastIndexOf("/")
  return idx <= 0 ? "" : path.slice(0, idx)
}

function sortedEntries(list: DavEntry[]): DavEntry[] {
  return [...list].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    return a.name.localeCompare(b.name)
  })
}