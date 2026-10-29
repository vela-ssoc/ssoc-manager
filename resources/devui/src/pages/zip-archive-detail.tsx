import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  Download,
  Edit3,
  FilePlus2,
  FolderPlus,
  Info,
  Loader2,
  Trash2,
  Upload,
} from "lucide-react"
import { toast } from "@/lib/toast"

import { cn } from "@/lib/utils"

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
import { Label } from "@/components/ui/label"
import { TooltipHint } from "@/components/tooltip-hint"
import {
  ContextMenu,
  ContextMenuItem,
  ContextMenuSeparator,
  type ContextMenuState,
} from "@/components/ui/context-menu"
import { zipArchiveApi } from "@/features/zip-archive/api"
import { ArchiveFormDialog } from "@/features/zip-archive/dialogs"
import type { ZipArchive, ZipArchiveForm } from "@/features/zip-archive/types"
import { formatDateTime } from "@/lib/format"
import { zipFileApi } from "@/features/zip-file/api"
import { ContentPreview } from "@/features/zip-file/content-preview"
import { FileTree } from "@/features/zip-file/file-tree"
import { buildFileTree } from "@/features/zip-file/tree"
import { formatFileSize } from "@/features/zip-file/types"
import type { TreeNode, ZipFile } from "@/features/zip-file/types"

export function ZipArchiveDetailPage() {
  const { name = "" } = useParams<{ name: string }>()
  const navigate = useNavigate()

  const [files, setFiles] = useState<ZipFile[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<ZipFile | null>(null)
  const [treeWidth, setTreeWidth] = useState(256)
  const [mkdirOpen, setMkdirOpen] = useState(false)
  const [mkdirName, setMkdirName] = useState("")
  const [mkdirDir, setMkdirDir] = useState<string>("")
  const [creating, setCreating] = useState(false)
  const [newFileOpen, setNewFileOpen] = useState(false)
  const [newFileName, setNewFileName] = useState("")
  const [newFileDir, setNewFileDir] = useState<string>("")
  const [newFileSource, setNewFileSource] = useState<File | null>(null)
  const [ctxMenu, setCtxMenu] = useState<ContextMenuState | null>(null)
  const [ctxNode, setCtxNode] = useState<TreeNode | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<TreeNode | null>(null)
  const [archive, setArchive] = useState<ZipArchive | null>(null)
  const [editOpen, setEditOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [infoNode, setInfoNode] = useState<TreeNode | null>(null)
  const replaceFileInputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement | null>(null)

  const requestId = useRef(0)

  const startDrag = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    const container = containerRef.current
    if (!container) return
    const containerLeft = container.getBoundingClientRect().left
    document.body.style.cursor = "col-resize"
    document.body.style.userSelect = "none"

    const onMove = (ev: MouseEvent) => {
      const next = ev.clientX - containerLeft
      setTreeWidth(Math.min(600, Math.max(160, next)))
    }
    const onUp = () => {
      document.body.style.cursor = ""
      document.body.style.userSelect = ""
      document.removeEventListener("mousemove", onMove)
      document.removeEventListener("mouseup", onUp)
    }

    document.addEventListener("mousemove", onMove)
    document.addEventListener("mouseup", onUp)
  }, [])

  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const ret = await zipFileApi.list(name)
      if (id !== requestId.current) return
      setFiles(ret.records ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [name])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    void zipArchiveApi
      .info(name)
      .then(setArchive)
      .catch((err) => {
        toast.error(err instanceof Error ? err.message : "加载归档信息失败")
      })
  }, [name])

  const tree = useMemo(() => buildFileTree(files), [files])

  const stats = useMemo(() => {
    let fileCount = 0
    let dirCount = 0
    let totalSize = 0
    for (const f of files) {
      if (f.is_directory) dirCount += 1
      else {
        fileCount += 1
        totalSize += f.file_size ?? 0
      }
    }
    return { fileCount, dirCount, totalSize }
  }, [files])

  const handleSelect = useCallback((node: TreeNode) => {
    if (node.file) setSelected(node.file)
  }, [])

  const targetDir = ctxNode
    ? ctxNode.children.length > 0 || ctxNode.file?.is_directory
      ? ctxNode.path
      : ctxNode.file?.directory ?? ""
    : ""

  const openMkdir = (dir: string) => {
    setMkdirDir(dir)
    setMkdirName("")
    setMkdirOpen(true)
  }

  const openNewFile = (dir: string) => {
    setNewFileDir(dir)
    setNewFileName("")
    setNewFileSource(null)
    setNewFileOpen(true)
  }

  const handleReplaceFile = async (file: File) => {
    if (!selected) return
    try {
      await zipFileApi.create(name, selected.full_name, file)
      toast.success("替换成功")
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "替换失败")
    }
  }

  const handleCreateFile = async () => {
    const fileName = newFileName.trim()
    if (!fileName) {
      toast.error("请输入文件名")
      return
    }
    const fullName = newFileDir ? `${newFileDir}/${fileName}` : fileName
    setCreating(true)
    try {
      if (newFileSource) {
        await zipFileApi.create(name, fullName, newFileSource)
      } else {
        await zipFileApi.update(name, fullName, "", fileName)
      }
      toast.success("创建成功")
      setNewFileOpen(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "创建失败")
    } finally {
      setCreating(false)
    }
  }

  const handleEditArchive = async (form: ZipArchiveForm, id?: string) => {
    setSaving(true)
    try {
      await zipArchiveApi.update(id ?? archive?.id ?? "", form)
      toast.success("更新成功")
      setEditOpen(false)
      if (archive) {
        setArchive({ ...archive, ...form })
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "更新失败")
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (node: TreeNode) => {
    const isDir = node.children.length > 0 || Boolean(node.file?.is_directory)
    if (isDir) {
      toast.error("删除目录功能存在 BUG，还在开发中...")
      setDeleteTarget(null)
      return
    }

    try {
      await zipFileApi.remove(name, node.path)
      toast.success("删除成功")
      if (selected && selected.full_name === node.path) {
        setSelected(null)
      }
      await load()
      setDeleteTarget(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除失败")
    }
  }

  const handleMkdir = async () => {
    const dirName = mkdirName.trim()
    if (!dirName) {
      toast.error("请输入目录名")
      return
    }
    const fullName = mkdirDir ? `${mkdirDir}/${dirName}` : dirName
    setCreating(true)
    try {
      await zipFileApi.mkdir(name, fullName)
      toast.success("目录创建成功")
      setMkdirOpen(false)
      setMkdirName("")
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "创建失败")
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft />
        </Button>
        <h2 className="truncate text-sm font-semibold">{name}</h2>
        <TooltipHint content="下载 ZIP">
          <a
            href={zipArchiveApi.downloadUrl(name)}
            aria-label="下载 ZIP"
            className="hover:bg-accent hover:text-accent-foreground flex size-7 items-center justify-center rounded-md"
          >
            <Download className="size-4" />
          </a>
        </TooltipHint>
        <TooltipHint content="编辑">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="编辑"
            onClick={() => setEditOpen(true)}
          >
            <Edit3 className="size-4" />
          </Button>
        </TooltipHint>

        <div className="text-muted-foreground ml-auto flex items-center gap-3 text-xs">
          <span>文件 {stats.fileCount}</span>
          <span>目录 {stats.dirCount}</span>
          <span>总大小 {formatFileSize(stats.totalSize)}</span>
        </div>
      </div>

      

      <div ref={containerRef} className="flex min-h-0 flex-1 overflow-hidden rounded-md border bg-card">
        <aside
          className="shrink-0 overflow-hidden"
          style={{ width: treeWidth }}
        >
          {loading ? (
            <div className="text-muted-foreground flex h-full items-center justify-center">
              <Loader2 className="size-5 animate-spin" />
            </div>
          ) : (
            <FileTree
              nodes={tree}
              selectedPath={selected?.full_name ?? ""}
              onSelect={handleSelect}
              onDelete={setDeleteTarget}
              onContextMenu={(e, node) => {
                setCtxNode(node)
                setCtxMenu({ x: e.clientX, y: e.clientY })
              }}
            />
          )}
        </aside>

        <div
          role="separator"
          aria-orientation="vertical"
          className={cn(
            "group relative w-1 shrink-0 cursor-col-resize border-r transition-colors hover:bg-ring/60"
          )}
          onMouseDown={startDrag}
        />

        <div className="min-w-0 flex-1">
          {selected ? (
            <ContentPreview
              key={selected.id}
              file={selected}
              source={{
                load: () => zipFileApi.content(name, selected.full_name),
                downloadUrl: zipFileApi.contentUrl(name, selected.full_name),
                save: async (content) => {
                  await zipFileApi.update(
                    name,
                    selected.full_name,
                    content,
                    selected.file_name
                  )
                },
              }}
              onChanged={() => {
                void load()
              }}
            />
          ) : (
            <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
              请选择左侧文件
            </div>
          )}
        </div>
      </div>

      <Dialog open={mkdirOpen} onOpenChange={setMkdirOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>新建目录</DialogTitle>
            <DialogDescription>
              在{mkdirDir ? `「${mkdirDir}」` : "根目录"}下创建目录
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-2">
            <Label htmlFor="mkdir-name">目录名</Label>
            <Input
              id="mkdir-name"
              value={mkdirName}
              placeholder="目录名"
              onChange={(e) => setMkdirName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleMkdir()
              }}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setMkdirOpen(false)}>
              取消
            </Button>
            <Button onClick={handleMkdir} disabled={creating}>
              {creating ? "创建中…" : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={newFileOpen} onOpenChange={setNewFileOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>新建文件</DialogTitle>
            <DialogDescription>
              在{newFileDir ? `「${newFileDir}」` : "根目录"}下创建文件
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-3">
            <div className="grid gap-2">
              <Label htmlFor="new-file-name">文件名</Label>
              <Input
                id="new-file-name"
                value={newFileName}
                placeholder="例如 index.html"
                onChange={(e) => setNewFileName(e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label>文件内容（可选）</Label>
              <label className="flex cursor-pointer flex-col items-center justify-center gap-1 rounded-md border border-dashed px-4 py-4 text-sm transition-colors hover:bg-accent/40">
                <Upload className="text-muted-foreground size-4" />
                {newFileSource ? (
                  <span className="text-xs">{newFileSource.name}</span>
                ) : (
                  <span className="text-muted-foreground text-xs">
                    留空则创建空文件，点击选择本机文件
                  </span>
                )}
                <input
                  type="file"
                  className="hidden"
                  onChange={(e) => {
                    setNewFileSource(e.target.files?.[0] ?? null)
                    const picked = e.target.files?.[0]
                    if (picked && newFileName.trim() === "") {
                      setNewFileName(picked.name)
                    }
                  }}
                />
              </label>
              {newFileSource && (
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground text-xs underline"
                  onClick={() => setNewFileSource(null)}
                >
                  移除文件
                </button>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setNewFileOpen(false)}>
              取消
            </Button>
            <Button onClick={handleCreateFile} disabled={creating}>
              {creating ? "创建中…" : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={infoNode !== null} onOpenChange={(open) => {
        if (!open) setInfoNode(null)
      }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>文件信息</DialogTitle>
          </DialogHeader>
          {infoNode && (
            <div className="grid gap-1 text-sm">
              <InfoRow label="名称" value={infoNode.name} />
              <InfoRow label="路径" value={infoNode.path} />
              {infoNode.file && !infoNode.file.is_directory && (
                <InfoRow label="类型" value={infoNode.file.mime_type || "-"} />
              )}
              {infoNode.file && !infoNode.file.is_directory && (
                <InfoRow
                  label="大小"
                  value={formatFileSize(infoNode.file.file_size)}
                />
              )}
              {infoNode.file && !infoNode.file.is_directory && (
                <InfoRow label="修改次数" value={`${infoNode.file.version} 次`} />
              )}
              {infoNode.file?.created_at && (
                <InfoRow
                  label="创建时间"
                  value={formatDateTime(infoNode.file.created_at)}
                />
              )}
              {infoNode.file?.updated_at && (
                <InfoRow
                  label="修改时间"
                  value={formatDateTime(infoNode.file.updated_at)}
                />
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setInfoNode(null)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteTarget !== null} onOpenChange={(open) => {
        if (!open) setDeleteTarget(null)
      }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除确认</DialogTitle>
            <DialogDescription>
              确认删除
              {deleteTarget?.file?.is_directory || deleteTarget?.children?.length
                ? "目录"
                : "文件"}
              「{deleteTarget?.path}」？
              {deleteTarget?.file?.is_directory || deleteTarget?.children?.length
                ? "其下所有内容将一并删除。"
                : ""}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                if (deleteTarget) handleDelete(deleteTarget)
              }}
            >
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ContextMenu
        state={ctxMenu}
        onClose={() => {
          setCtxMenu(null)
          setCtxNode(null)
        }}
      >
        {ctxNode ? null : (
          <>
            <ContextMenuItem
              onClick={() => {
                openMkdir("")
                setCtxMenu(null)
              }}
            >
              <FolderPlus /> 新建目录
            </ContextMenuItem>
            <ContextMenuItem
              onClick={() => {
                openNewFile("")
                setCtxMenu(null)
              }}
            >
              <FilePlus2 /> 新建文件
            </ContextMenuItem>
          </>
        )}

        {ctxNode && !(ctxNode.children.length > 0 || ctxNode.file?.is_directory) && (
          <ContextMenuItem
            onClick={() => {
              const node = ctxNode
              setCtxMenu(null)
              if (node?.file) setSelected(node.file)
              replaceFileInputRef.current?.click()
            }}
          >
            <Upload /> 上传替换
          </ContextMenuItem>
        )}

        {ctxNode &&
          (ctxNode.children.length > 0 || ctxNode.file?.is_directory) && (
          <>
            <ContextMenuItem
              onClick={() => {
                openMkdir(targetDir)
                setCtxMenu(null)
              }}
            >
              <FolderPlus /> 新建目录
            </ContextMenuItem>
            <ContextMenuItem
              onClick={() => {
                openNewFile(targetDir)
                setCtxMenu(null)
              }}
            >
              <FilePlus2 /> 新建文件
            </ContextMenuItem>
          </>
        )}

        {ctxNode && (
          <ContextMenuItem
            onClick={() => {
              setInfoNode(ctxNode)
              setCtxMenu(null)
            }}
          >
            <Info /> 文件信息
          </ContextMenuItem>
        )}

        {ctxNode && (
          <>
            <ContextMenuSeparator />
            <ContextMenuItem
              variant="destructive"
              onClick={() => {
                const node = ctxNode
                setCtxMenu(null)
                if (node) setDeleteTarget(node)
              }}
            >
              <Trash2 /> 删除
            </ContextMenuItem>
          </>
        )}
      </ContextMenu>

      <input
        ref={replaceFileInputRef}
        type="file"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file) void handleReplaceFile(file)
          e.target.value = ""
        }}
      />

      <ArchiveFormDialog
        open={editOpen}
        onOpenChange={(open) => {
          if (!open) setEditOpen(false)
        }}
        initial={archive ?? undefined}
        submitting={saving}
        onSubmit={handleEditArchive}
      />
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-4 border-b py-2 last:border-b-0">
      <span className="text-muted-foreground shrink-0">{label}</span>
      <span className="text-right font-medium break-all">{value}</span>
    </div>
  )
}
