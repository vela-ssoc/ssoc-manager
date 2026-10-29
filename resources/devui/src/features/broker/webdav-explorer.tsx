import { useEffect, useMemo, useState } from "react"
import { createPortal } from "react-dom"
import { toast } from "@/lib/toast"
import {
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ChevronRight,
  Loader2,
  PictureInPicture2,
  RefreshCw,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { usePortalPictureInPicture } from "@/lib/use-picture-in-picture"
import { ContentPreview, type PreviewFile } from "@/features/zip-file/content-preview"
import { FileIcon } from "@/features/zip-file/file-icon"
import { formatFileSize } from "@/features/zip-file/types"

import {
  davDownloadUrl,
  davList,
  splitPath,
  type DavEndpoint,
  type DavEntry,
} from "./webdav"

export function WebDavExplorer({
  clientUrl,
  apiPath,
  title,
}: {
  clientUrl: string
  apiPath: string
  title?: string
}) {
  const endpoint: DavEndpoint = { clientUrl, apiPath }
  const [path, setPath] = useState("")
  const [backStack, setBackStack] = useState<string[]>([])
  const [forwardStack, setForwardStack] = useState<string[]>([])
  const [entries, setEntries] = useState<DavEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [preview, setPreview] = useState<{ stat: DavEntry; path: string } | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  const pip = usePortalPictureInPicture()

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

  if (pip.active && pip.container) {
    return createPortal(content(), pip.container)
  }

  return content()

  function content() {
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
          {pip.supported && !pip.active && (
            <TooltipHint content="画中画打开">
              <Button
                variant="ghost"
                size="icon"
                className="size-7"
                aria-label="画中画打开"
                onClick={() => void pip.open(title)}
              >
                <PictureInPicture2 className="size-3.5" />
              </Button>
            </TooltipHint>
          )}

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

        <div className="flex min-h-0 flex-1 overflow-hidden rounded-md border bg-card">
          <div className="flex min-w-0 flex-[3] flex-col">
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
                <div className="grid grid-cols-[repeat(auto-fill,minmax(6.5rem,1fr))] content-start gap-2 p-2">
                  {sorted.map((stat) => (
                    <FileTile
                      key={stat.path}
                      stat={stat}
                      selected={preview?.path === stat.path}
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
}

function FileTile({
  stat,
  selected,
  onClick,
}: {
  stat: DavEntry
  selected: boolean
  onClick: () => void
}) {
  const sizeText = formatFileSize(stat.size)
  const timeText = stat.lastmod ? ` · ${formatDateTime(stat.lastmod)}` : ""
  const detail = stat.isDir ? stat.name : `${stat.name}\n${sizeText}${timeText}`

  return (
    <TooltipHint content={detail}>
      <div
        role="button"
        tabIndex={0}
        className={cn(
          "flex min-h-24 min-w-0 cursor-pointer flex-col items-center gap-1 rounded-md border border-transparent px-1.5 py-2 text-center transition-colors hover:bg-accent/40",
          selected && "border-primary/30 bg-accent/60"
        )}
        onClick={onClick}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault()
            onClick()
          }
        }}
      >
        <span className="flex h-12 shrink-0 items-center justify-center">
          <FileIcon name={stat.name} isDir={stat.isDir} className="size-11" />
        </span>
        <span className="line-clamp-2 w-full break-all text-xs leading-tight">
          {stat.name}
        </span>
      </div>
    </TooltipHint>
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
  const file: PreviewFile = {
    id: path,
    file_name: stat.name,
    full_name: path,
    file_size: stat.size,
    mime_type: "",
    is_directory: stat.isDir,
  }

  return (
    <ContentPreview
      key={path}
      file={file}
      source={{
        load: async () => {
          const response = await fetch(davDownloadUrl(endpoint, path))
          if (!response.ok) {
            throw new Error(`请求失败（${response.status}）`)
          }
          return response.blob()
        },
        downloadUrl: davDownloadUrl(endpoint, path),
      }}
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