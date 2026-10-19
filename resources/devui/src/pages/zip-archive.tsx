import { useCallback, useEffect, useRef, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "@/lib/toast"
import {
  Archive,
  Download,
  Edit3,
  Plus,
  Search,
  Trash2,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DataPagination } from "@/components/data-pagination"
import { Input } from "@/components/ui/input"
import { TooltipHint } from "@/components/tooltip-hint"
import { zipArchiveApi } from "@/features/zip-archive/api"
import { zipFileApi } from "@/features/zip-file/api"
import type { ZipArchive, ZipArchiveForm } from "@/features/zip-archive/types"
import { methodLabel } from "@/features/zip-archive/types"
import {
  ArchiveFormDialog,
  CreateArchiveDialog,
} from "@/features/zip-archive/dialogs"
import { formatDateTime } from "@/lib/format"

export function ZipArchivePage() {
  const navigate = useNavigate()
  const [size, setSize] = useState(12)
  const [records, setRecords] = useState<ZipArchive[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState("")
  const [search, setSearch] = useState("")

  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<ZipArchive | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const requestId = useRef(0)

  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const ret = await zipArchiveApi.page({ page, size, q: search })
      if (id !== requestId.current) return
      setRecords(ret.records ?? [])
      setTotal(ret.total ?? 0)
      setPage(ret.page ?? page)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [page, size, search])

  useEffect(() => {
    void load()
  }, [load])

  const handleSubmit = async (form: ZipArchiveForm, id?: string) => {
    setSubmitting(true)
    try {
      if (id) {
        await zipArchiveApi.update(id, form)
        toast.success("更新成功")
      } else {
        await zipArchiveApi.create(form)
        toast.success("创建成功")
      }
      setCreateOpen(false)
      setEditing(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存失败")
    } finally {
      setSubmitting(false)
    }
  }

  const handleCreate = async (form: ZipArchiveForm, file?: File) => {
    setSubmitting(true)
    try {
      if (file) {
        await zipFileApi.unzip(form, file)
        toast.success("上传成功")
      } else {
        await zipArchiveApi.create(form)
        toast.success("创建成功")
      }
      setCreateOpen(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "创建失败")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={keyword}
            placeholder="搜索"
            className="h-7 pl-9"
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                setPage(1)
                setSearch(keyword.trim())
              }
            }}
          />
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus /> 新建
        </Button>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            加载中…
          </div>
        ) : records.length === 0 ? (
          <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-3 text-sm">
            <Archive className="size-10 opacity-40" />
            <span>暂无数据</span>
          </div>
        ) : (
          <div className="grid min-h-0 flex-1 auto-rows-min grid-cols-1 gap-2 overflow-auto [scrollbar-gutter:stable] sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {records.map((item) => (
              <div
                key={item.id}
                role="button"
                tabIndex={0}
                className="group flex flex-col gap-3 rounded-md border bg-card p-4 transition-all hover:border-border hover:bg-accent/40"
                onClick={() => navigate(`/zip-archives/${item.name}`)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") navigate(`/zip-archives/${item.name}`)
                }}
              >
                <div className="flex items-start justify-between gap-2">
                  <h3 className="truncate text-sm font-semibold">{item.name}</h3>
                  <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <TooltipHint content="下载">
                      <a
                        href={zipArchiveApi.downloadUrl(item.name)}
                        aria-label="下载"
                        className="flex size-7 items-center justify-center rounded-md hover:bg-accent hover:text-accent-foreground"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <Download className="size-3.5" />
                      </a>
                    </TooltipHint>
                    <TooltipHint content="编辑">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7"
                        aria-label="编辑"
                        onClick={(e) => {
                          e.stopPropagation()
                          setEditing(item)
                        }}
                      >
                        <Edit3 />
                      </Button>
                    </TooltipHint>
                    <TooltipHint content="删除">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7 text-destructive"
                        aria-label="删除"
                        onClick={(e) => {
                          e.stopPropagation()
                          toast.error("删除归档功能还在开发中...")
                        }}
                      >
                        <Trash2 />
                      </Button>
                    </TooltipHint>
                  </div>
                </div>

                {item.description ? (
                  <p className="text-muted-foreground line-clamp-2 text-xs leading-relaxed">
                    {item.description}
                  </p>
                ) : (
                  <p className="text-muted-foreground/60 text-xs">暂无描述</p>
                )}

                <div className="mt-auto flex items-center gap-2">
                  <Badge variant="secondary">
                    {methodLabel(item.method)}
                  </Badge>
                  <span className="text-muted-foreground text-xs">
                    修改 {item.version} 次
                  </span>
                  <span className="text-muted-foreground/60 ml-auto text-xs">
                    {formatDateTime(item.updated_at)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <DataPagination
        current={page}
        size={size}
        total={total}
        loading={loading}
        onPageChange={setPage}
        onPageSizeChange={(nextSize) => {
          setSize(nextSize)
          setPage(1)
        }}
      />

      <CreateArchiveDialog
        open={createOpen}
        onOpenChange={(open) => {
          if (!open) setCreateOpen(false)
        }}
        submitting={submitting}
        onSubmit={handleCreate}
      />

      <ArchiveFormDialog
        open={editing !== null}
        onOpenChange={(open) => {
          if (!open) {
            setEditing(null)
          }
        }}
        initial={editing ?? undefined}
        submitting={submitting}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
