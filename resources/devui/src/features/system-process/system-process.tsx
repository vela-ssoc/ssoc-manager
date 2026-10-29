import { useCallback, useEffect, useMemo, useState } from "react"
import {
  FileText,
  ListTree,
  Loader2,
  Network,
  RefreshCw,
  Search,
} from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { formatBytes } from "@/features/broker/types"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"

import { systemProcessApi } from "./api"
import type {
  ProcessConnection,
  ProcessConnectionAddress,
  ProcessInfo,
  ProcessOpenFile,
} from "./types"

const FAMILY_LABELS: Record<number, string> = {
  2: "IPv4",
  10: "IPv6",
  23: "IPv6",
  30: "IPv6",
}

const TYPE_LABELS: Record<number, string> = {
  1: "TCP",
  2: "UDP",
}

export function SystemProcessSection({ slug }: { slug?: string }) {
  const [records, setRecords] = useState<ProcessInfo[]>([])
  const [search, setSearch] = useState("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [refreshToken, setRefreshToken] = useState(0)
  const [selected, setSelected] = useState<ProcessInfo | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result = await systemProcessApi.list(slug)
      setRecords(result.records ?? [])
      setError(false)
    } catch (err) {
      setError(true)
      toast.error(err instanceof Error ? err.message : "加载系统进程失败")
    } finally {
      setLoading(false)
    }
  }, [slug])

  useEffect(() => {
    void load()
  }, [load, refreshToken])

  const filtered = useMemo(() => {
    const keyword = search.trim().toLowerCase()
    if (!keyword) return [...records].sort((a, b) => a.pid - b.pid)
    const list = records.filter((process) =>
      [
        String(process.pid),
        process.name,
        process.username,
        process.exe,
        process.cwd,
        process.cmdline,
      ].some((value) => value?.toLowerCase().includes(keyword))
    )
    return [...list].sort((a, b) => a.pid - b.pid)
  }, [records, search])

  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={search}
            placeholder="搜索 PID、名称、用户、命令行"
            className="h-7 pl-9"
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <span className="text-muted-foreground text-xs">
          共 {records.length} 个进程
          {search.trim() ? `，匹配 ${filtered.length} 个` : ""}
        </span>
        <TooltipHint content="刷新">
          <Button
            variant="ghost"
            size="icon"
            className="ml-auto size-7"
            aria-label="刷新"
            disabled={loading}
            onClick={() => setRefreshToken((token) => token + 1)}
          >
            <RefreshCw className={cn("size-3.5", loading && "animate-spin")} />
          </Button>
        </TooltipHint>
      </div>

      <div className="min-h-0 flex-1 overflow-auto rounded-md border [scrollbar-gutter:stable]">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center gap-2 text-sm">
            <Loader2 className="size-4 animate-spin" />
            正在加载系统进程…
          </div>
        ) : error && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-3 text-sm">
            <ListTree className="size-10 opacity-40" />
            <span>加载系统进程失败</span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setRefreshToken((token) => token + 1)}
            >
              重试
            </Button>
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2 text-sm">
            <ListTree className="size-10 opacity-40" />
            <span>{records.length === 0 ? "暂无系统进程" : "没有匹配项"}</span>
          </div>
        ) : (
          <table className="w-full min-w-[860px] table-fixed text-xs">
            <thead className="sticky top-0 z-10">
              <tr className="border-b">
                <Th className="w-20 text-right">PID</Th>
                <Th className="w-48">名称</Th>
                <Th className="w-32">用户</Th>
                <Th className="w-20 text-right">CPU</Th>
                <Th className="w-16 text-right">线程</Th>
                <Th className="w-16 text-right">FD</Th>
                <Th className="w-24 text-right">内存</Th>
                <Th className="w-40">启动时间</Th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((process) => (
                <tr
                  key={process.pid}
                  tabIndex={0}
                  className="hover:bg-accent/40 focus-visible:bg-accent/40 cursor-pointer border-b outline-none last:border-b-0"
                  onClick={() => setSelected(process)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault()
                      setSelected(process)
                    }
                  }}
                >
                  <td className="px-3 py-1.5 text-right font-mono">{process.pid}</td>
                  <td className="px-3 py-1.5">
                    <TooltipHint content={process.exe || process.name || "—"}>
                      <div className="truncate font-medium">{process.name || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-3 py-1.5">
                    <TooltipHint content={process.username || "—"}>
                      <div className="truncate">{process.username || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-3 py-1.5 text-right font-mono">
                    {cpuText(process.cpu_percent)}
                  </td>
                  <td className="px-3 py-1.5 text-right font-mono">
                    {countText(process.num_threads)}
                  </td>
                  <td className="px-3 py-1.5 text-right font-mono">
                    {countText(process.num_fds)}
                  </td>
                  <td className="px-3 py-1.5 text-right font-mono">
                    {memoryText(process)}
                  </td>
                  <td className="text-muted-foreground px-3 py-1.5">
                    {processTime(process.created_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <ProcessDetailDialog
        slug={slug}
        process={selected}
        onClose={() => setSelected(null)}
      />
    </div>
  )
}

function ProcessDetailDialog({
  slug,
  process,
  onClose,
}: {
  slug?: string
  process: ProcessInfo | null
  onClose: () => void
}) {
  const [connections, setConnections] = useState<ProcessConnection[]>([])
  const [openFiles, setOpenFiles] = useState<ProcessOpenFile[]>([])
  const [connectionsLoading, setConnectionsLoading] = useState(false)
  const [openFilesLoading, setOpenFilesLoading] = useState(false)
  const [connectionsError, setConnectionsError] = useState(false)
  const [openFilesError, setOpenFilesError] = useState(false)
  const [detailToken, setDetailToken] = useState(0)

  useEffect(() => {
    if (!process) {
      setConnections([])
      setOpenFiles([])
      setConnectionsLoading(false)
      setOpenFilesLoading(false)
      setConnectionsError(false)
      setOpenFilesError(false)
      return
    }

    let cancelled = false
    setConnectionsLoading(true)
    setOpenFilesLoading(true)
    setConnectionsError(false)
    setOpenFilesError(false)

    systemProcessApi
      .connections(slug, process.pid)
      .then((result) => {
        if (!cancelled) setConnections(result.records ?? [])
      })
      .catch(() => {
        if (!cancelled) setConnectionsError(true)
      })
      .finally(() => {
        if (!cancelled) setConnectionsLoading(false)
      })

    systemProcessApi
      .openFiles(slug, process.pid)
      .then((result) => {
        if (!cancelled) setOpenFiles(result.records ?? [])
      })
      .catch(() => {
        if (!cancelled) setOpenFilesError(true)
      })
      .finally(() => {
        if (!cancelled) setOpenFilesLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [detailToken, process, slug])

  const sortedConnections = useMemo(
    () => [...connections].sort((a, b) => a.fd - b.fd),
    [connections]
  )
  const sortedOpenFiles = useMemo(
    () => [...openFiles].sort((a, b) => a.fd - b.fd),
    [openFiles]
  )

  return (
    <Dialog
      open={process !== null}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogContent className="sm:max-w-5xl" aria-describedby={undefined}>
        {process && (
          <>
            <DialogHeader>
              <div className="flex items-start justify-between gap-2 pr-8">
                <div className="min-w-0">
                  <DialogTitle className="flex min-w-0 items-baseline gap-2">
                    <span className="truncate">{process.name || "未知进程"}</span>
                    <span className="text-muted-foreground shrink-0 font-mono text-xs font-normal">
                      PID {process.pid}
                    </span>
                  </DialogTitle>
                  <DialogDescription className="truncate">
                    {process.cmdline || process.exe || "—"}
                  </DialogDescription>
                </div>
                <TooltipHint content="刷新详情">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7 shrink-0"
                    aria-label="刷新详情"
                    disabled={connectionsLoading || openFilesLoading}
                    onClick={() => setDetailToken((token) => token + 1)}
                  >
                    <RefreshCw
                      className={cn(
                        "size-3.5",
                        (connectionsLoading || openFilesLoading) && "animate-spin"
                      )}
                    />
                  </Button>
                </TooltipHint>
              </div>
            </DialogHeader>

            <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-1">
              <section className="grid gap-3">
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  <InfoItem label="用户" value={process.username} />
                  <InfoItem label="CPU" value={cpuText(process.cpu_percent)} />
                  <InfoItem label="线程数" value={countText(process.num_threads)} />
                  <InfoItem label="文件描述符" value={countText(process.num_fds)} />
                  <InfoItem label="内存 RSS" value={memoryValue(process, "rss")} />
                  <InfoItem label="内存 VMS" value={memoryValue(process, "vms")} />
                  <InfoItem label="启动时间" value={processTime(process.created_at)} />
                </div>
                <div className="grid gap-3 sm:grid-cols-2">
                  <InfoItem label="可执行文件" value={process.exe} mono />
                  <InfoItem label="工作目录" value={process.cwd} mono />
                </div>
                <div className="grid gap-1">
                  <div className="text-muted-foreground text-[11px]">命令行</div>
                  <div className="bg-muted/30 rounded-md border px-2.5 py-1.5 font-mono text-xs break-all">
                    {process.cmdline || "—"}
                  </div>
                </div>
              </section>

              <section className="grid gap-2">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5">
                    <Network className="text-muted-foreground size-3.5" />
                    <h4 className="text-sm font-semibold">打开的连接</h4>
                    <span className="text-muted-foreground text-xs">
                      ({sortedConnections.length})
                    </span>
                  </div>
                  {connectionsLoading && (
                    <Loader2 className="text-muted-foreground size-3.5 animate-spin" />
                  )}
                </div>

                {connectionsError ? (
                  <div className="text-muted-foreground rounded-md border border-dashed px-3 py-4 text-center text-xs">
                    加载连接失败，请点击右上角刷新重试
                  </div>
                ) : sortedConnections.length === 0 ? (
                  <div className="text-muted-foreground rounded-md border border-dashed px-3 py-4 text-center text-xs">
                    {connectionsLoading ? "正在加载连接…" : "暂无连接"}
                  </div>
                ) : (
                  <div className="overflow-x-auto rounded-md border">
                    <table className="w-full min-w-[640px] table-fixed text-xs">
                      <thead>
                        <tr className="bg-muted/40 text-muted-foreground border-b">
                          <th className="w-14 px-2 py-1.5 text-right font-medium">FD</th>
                          <th className="w-28 px-2 py-1.5 text-left font-medium">协议</th>
                          <th className="px-2 py-1.5 text-left font-medium">本地地址</th>
                          <th className="px-2 py-1.5 text-left font-medium">远端地址</th>
                          <th className="w-28 px-2 py-1.5 text-left font-medium">状态</th>
                        </tr>
                      </thead>
                      <tbody>
                        {sortedConnections.map((connection, index) => {
                          const local = formatAddress(connection.localaddr)
                          const remote = formatAddress(connection.remoteaddr)
                          return (
                            <tr
                              key={`${connection.fd}-${index}`}
                              className="border-b last:border-b-0"
                            >
                              <td className="px-2 py-1.5 text-right font-mono">
                                {connection.fd}
                              </td>
                              <td className="px-2 py-1.5">
                                {protocolText(connection)}
                              </td>
                              <td className="px-2 py-1.5">
                                <TooltipHint content={local}>
                                  <div className="truncate font-mono">{local}</div>
                                </TooltipHint>
                              </td>
                              <td className="px-2 py-1.5">
                                <TooltipHint content={remote}>
                                  <div className="truncate font-mono">{remote}</div>
                                </TooltipHint>
                              </td>
                              <td className="px-2 py-1.5">
                                <TooltipHint content={connection.status || "—"}>
                                  <div className="truncate">
                                    {connection.status || "—"}
                                  </div>
                                </TooltipHint>
                              </td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>

              <section className="grid gap-2">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5">
                    <FileText className="text-muted-foreground size-3.5" />
                    <h4 className="text-sm font-semibold">打开的文件</h4>
                    <span className="text-muted-foreground text-xs">
                      ({sortedOpenFiles.length})
                    </span>
                  </div>
                  {openFilesLoading && (
                    <Loader2 className="text-muted-foreground size-3.5 animate-spin" />
                  )}
                </div>

                {openFilesError ? (
                  <div className="text-muted-foreground rounded-md border border-dashed px-3 py-4 text-center text-xs">
                    加载文件失败，请点击右上角刷新重试
                  </div>
                ) : sortedOpenFiles.length === 0 ? (
                  <div className="text-muted-foreground rounded-md border border-dashed px-3 py-4 text-center text-xs">
                    {openFilesLoading ? "正在加载文件…" : "暂无文件"}
                  </div>
                ) : (
                  <div className="overflow-x-auto rounded-md border">
                    <table className="w-full min-w-[480px] table-fixed text-xs">
                      <thead>
                        <tr className="bg-muted/40 text-muted-foreground border-b">
                          <th className="w-14 px-2 py-1.5 text-right font-medium">FD</th>
                          <th className="px-2 py-1.5 text-left font-medium">路径</th>
                        </tr>
                      </thead>
                      <tbody>
                        {sortedOpenFiles.map((file, index) => (
                          <tr
                            key={`${file.fd}-${index}`}
                            className="border-b last:border-b-0"
                          >
                            <td className="px-2 py-1.5 text-right font-mono">
                              {file.fd}
                            </td>
                            <td className="px-2 py-1.5">
                              <TooltipHint content={file.path || "—"}>
                                <div className="truncate font-mono">
                                  {file.path || "—"}
                                </div>
                              </TooltipHint>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={onClose}>
                关闭
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

function Th({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <th className={cn("bg-muted px-3 py-2 font-medium", className)}>{children}</th>
  )
}

function InfoItem({
  label,
  value,
  mono = false,
}: {
  label: string
  value?: string
  mono?: boolean
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="grid content-start gap-0.5">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={text}>
        <div className={cn("truncate text-xs", mono && "font-mono")}>{text}</div>
      </TooltipHint>
    </div>
  )
}

function cpuText(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "—"
  return `${value.toFixed(1)}%`
}

function countText(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "—"
  return String(value)
}

function memoryText(process: ProcessInfo): string {
  if (!process.memory) return "—"
  return formatBytes(process.memory.rss)
}

function memoryValue(
  process: ProcessInfo,
  key: "rss" | "vms" | "hwm"
): string {
  const value = process.memory?.[key]
  return value === undefined ? "—" : formatBytes(value)
}

function processTime(value?: string): string {
  if (!value) return "—"
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 1980) return "—"
  return formatDateTime(value)
}

function familyLabel(family: number): string {
  return FAMILY_LABELS[family] ?? String(family)
}

function typeLabel(type: number): string {
  return TYPE_LABELS[type] ?? String(type)
}

function protocolText(connection: ProcessConnection): string {
  return `${typeLabel(connection.type)} · ${familyLabel(connection.family)}`
}

function formatAddress(address: ProcessConnectionAddress): string {
  const ip = address.ip?.trim()
  if (!ip) return "—"
  const host = ip.includes(":") ? `[${ip}]` : ip
  return `${host}:${address.port}`
}