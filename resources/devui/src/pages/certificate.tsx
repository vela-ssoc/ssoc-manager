import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import type { ReactNode } from "react"
import {
  FileBadge,
  Pencil,
  Plus,
  ScanLine,
  Search,
  Trash2,
  Upload,
  X,
} from "lucide-react"
import { z } from "zod"

import { TooltipHint } from "@/components/tooltip-hint"
import { Badge } from "@/components/ui/badge"
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
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { certificateApi } from "@/features/certificate/api"
import type { Certificate } from "@/features/certificate/types"

const MAX_NAMES = 3

export function CertificatePage() {
  const [records, setRecords] = useState<Certificate[]>([])
  const [loading, setLoading] = useState(true)
  const [query, setQuery] = useState("")
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Certificate | undefined>()
  const [deleteTarget, setDeleteTarget] = useState<Certificate | null>(null)
  const [deleting, setDeleting] = useState(false)

  const requestId = useRef(0)
  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const data = await certificateApi.list()
      if (id !== requestId.current) return
      setRecords(data ?? [])
    } catch (err) {
      if (id !== requestId.current) return
      setRecords([])
      toast.error(err instanceof Error ? err.message : "加载证书失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    if (!keyword) return records
    return records.filter((cert) =>
      [
        cert.common_name,
        cert.issuer?.common_name,
        cert.issuer?.organization?.join(" "),
        cert.certificate_sha256,
        ...[...(cert.dns_names ?? []), ...(cert.ip_addresses ?? [])],
      ].some((value) => value && value.toLowerCase().includes(keyword))
    )
  }, [query, records])

  const remove = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await certificateApi.remove(deleteTarget.id)
      toast.success("证书已删除")
      setDeleteTarget(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除证书失败")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={query}
            placeholder="搜索通用名、域名或指纹"
            className="h-7 pl-9"
            onChange={(event) => setQuery(event.target.value)}
          />
        </div>
        <Button
          size="sm"
          onClick={() => {
            setEditing(undefined)
            setDialogOpen(true)
          }}
        >
          <Plus /> 新建证书
        </Button>
        <span className="text-muted-foreground ml-auto text-xs">
          {loading ? "加载中…" : `共 ${records.length} 张`}
        </span>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            加载中…
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2 text-sm">
            <FileBadge className="size-9 opacity-40" />
            <span>{records.length === 0 ? "暂无证书" : "没有匹配项"}</span>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 2xl:grid-cols-3">
            {filtered.map((cert) => (
              <CertificateCard
                key={cert.id}
                cert={cert}
                onEdit={() => {
                  setEditing(cert)
                  setDialogOpen(true)
                }}
                onDelete={() => setDeleteTarget(cert)}
              />
            ))}
          </div>
        )}
      </div>

      <CertificateDialog
        open={dialogOpen}
        editing={editing}
        onClose={() => setDialogOpen(false)}
        onSaved={async () => {
          setDialogOpen(false)
          await load()
        }}
      />

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteTarget(null)
        }}
      >
        <DialogContent showCloseButton={false} className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除证书</DialogTitle>
            <DialogDescription>
              将删除证书「{deleteTarget?.common_name || "未命名"}」，此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              disabled={deleting}
              onClick={() => setDeleteTarget(null)}
            >
              取消
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={remove}>
              {deleting ? "删除中…" : "删除"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

const formSchema = z.object({
  publicKey: z.string().trim().min(1, "请粘贴公钥证书（PEM）"),
  privateKey: z.string().trim().min(1, "请粘贴私钥（PEM）"),
})

function CertificateDialog({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean
  editing?: Certificate
  onClose: () => void
  onSaved: () => void | Promise<void>
}) {
  const [publicKey, setPublicKey] = useState("")
  const [privateKey, setPrivateKey] = useState("")
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [preview, setPreview] = useState<Certificate | null>(null)
  const [parsing, setParsing] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return
    setPublicKey("")
    setPrivateKey("")
    setErrors({})
    setPreview(null)
  }, [open])

  const parse = async () => {
    const result = formSchema.safeParse({ publicKey, privateKey })
    if (!result.success) {
      setErrors(fieldErrors(result.error))
      return
    }
    setErrors({})
    setParsing(true)
    try {
      const cert = await certificateApi.parse(
        result.data.publicKey,
        result.data.privateKey
      )
      setPreview(cert)
    } catch (err) {
      setPreview(null)
      toast.error(err instanceof Error ? err.message : "证书解析失败")
    } finally {
      setParsing(false)
    }
  }

  const submit = async () => {
    const result = formSchema.safeParse({ publicKey, privateKey })
    if (!result.success) {
      setErrors(fieldErrors(result.error))
      return
    }
    setErrors({})
    setSubmitting(true)
    try {
      const payload = {
        public_key: result.data.publicKey,
        private_key: result.data.privateKey,
      }
      if (editing) {
        await certificateApi.update(editing.id, payload)
        toast.success("证书已更新")
      } else {
        await certificateApi.create(payload)
        toast.success("证书已创建")
      }
      await onSaved()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存证书失败")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && !submitting) onClose()
      }}
    >
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{editing ? "编辑证书" : "新建证书"}</DialogTitle>
          <DialogDescription>
            {editing
              ? "列表不返回密钥内容，请重新粘贴或选择证书与私钥后保存。"
              : "粘贴 PEM 格式的证书与私钥，或选择本地文件自动填充；保存时由服务端解析并落库。"}
          </DialogDescription>
        </DialogHeader>

        <div className="grid max-h-[65vh] gap-3 overflow-y-auto pr-1">
          <PemField
            id="cert-public-key"
            label="公钥证书"
            value={publicKey}
            placeholder="-----BEGIN CERTIFICATE-----"
            accept=".pem,.crt,.cer,.txt"
            error={errors.publicKey}
            onChange={setPublicKey}
          />

          <PemField
            id="cert-private-key"
            label="私钥"
            value={privateKey}
            placeholder="-----BEGIN PRIVATE KEY-----"
            accept=".pem,.key,.txt"
            error={errors.privateKey}
            onChange={setPrivateKey}
          />

          {preview && (
            <section className="grid gap-2 rounded-md border p-3">
              <h3 className="text-muted-foreground text-[11px] font-medium">
                解析结果
              </h3>
              <div className="grid gap-x-6 gap-y-1.5 sm:grid-cols-2">
                <PreviewField label="通用名" value={preview.common_name} />
                <PreviewField
                  label="签发者"
                  value={
                    preview.issuer?.common_name ||
                    preview.issuer?.organization?.join(", ")
                  }
                />
                <PreviewField
                  label="生效时间"
                  value={preview.not_before && formatDateTime(preview.not_before)}
                />
                <PreviewField
                  label="过期时间"
                  value={preview.not_after && formatDateTime(preview.not_after)}
                />
                <PreviewField
                  label="域名 / IP"
                  value={namesOf(preview).join(", ")}
                />
                <PreviewField
                  label="签名算法"
                  value={preview.signature_algorithm}
                />
              </div>
              <PreviewField
                label="SHA256 指纹"
                value={preview.certificate_sha256}
                mono
              />
            </section>
          )}
        </div>

        <DialogFooter className="sm:justify-between">
          <Button
            variant="outline"
            disabled={parsing || submitting}
            onClick={parse}
          >
            <ScanLine /> {parsing ? "解析中…" : "解析预览"}
          </Button>
          <div className="flex flex-col-reverse gap-2 sm:flex-row">
            <Button variant="outline" disabled={submitting} onClick={onClose}>
              取消
            </Button>
            <Button disabled={submitting} onClick={submit}>
              {submitting ? "保存中…" : "保存"}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function PemField({
  id,
  label,
  value,
  placeholder,
  accept,
  error,
  onChange,
}: {
  id: string
  label: string
  value: string
  placeholder: string
  accept: string
  error?: string
  onChange: (value: string) => void
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [fileName, setFileName] = useState("")
  const [reading, setReading] = useState(false)
  const [dragging, setDragging] = useState(false)

  const read = async (file: File) => {
    if (file.size > 512 * 1024) {
      toast.error("文件过大，证书文件通常在 10 KB 以内")
      return
    }
    setReading(true)
    try {
      const text = await file.text()
      if (!text.trim()) {
        toast.error("文件内容为空")
        return
      }
      onChange(text)
      setFileName(file.name)
    } catch {
      toast.error("读取文件失败")
    } finally {
      setReading(false)
    }
  }

  const clear = () => {
    onChange("")
    setFileName("")
    if (inputRef.current) inputRef.current.value = ""
  }

  return (
    <div className="grid gap-1.5">
      <div
        className={cn(
          "overflow-hidden rounded-md border transition-colors",
          dragging && "border-primary/60 bg-primary/5",
          error && "border-destructive/50"
        )}
        onDragOver={(event) => {
          event.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => {
          event.preventDefault()
          setDragging(false)
          const file = event.dataTransfer.files?.[0]
          if (file) void read(file)
        }}
      >
        <div className="bg-muted/40 flex min-h-9 items-center gap-2 border-b px-2.5 py-1">
          <label className="shrink-0 text-xs font-medium" htmlFor={id}>
            {label}
          </label>
          <input
            ref={inputRef}
            type="file"
            accept={accept}
            className="hidden"
            onChange={(event) => {
              const file = event.target.files?.[0]
              event.target.value = ""
              if (file) void read(file)
            }}
          />
          <TooltipHint content={reading ? "读取中…" : "选择文件"}>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-6 shrink-0"
              aria-label={`选择${label}文件`}
              disabled={reading}
              onClick={() => inputRef.current?.click()}
            >
              <Upload className="size-3.5" />
            </Button>
          </TooltipHint>
          {value.length > 0 && (
            <span className="text-muted-foreground shrink-0 text-[10px]">
              {value.length} 字符
            </span>
          )}
          <div className="ml-auto flex min-w-0 items-center gap-1">
            {fileName && (
              <TooltipHint content={fileName}>
                <span className="text-muted-foreground max-w-32 truncate text-[11px]">
                  {fileName}
                </span>
              </TooltipHint>
            )}
            {value.length > 0 && (
              <TooltipHint content="清空">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6"
                  aria-label={`清空${label}`}
                  onClick={clear}
                >
                  <X className="size-3.5" />
                </Button>
              </TooltipHint>
            )}
          </div>
        </div>
        <textarea
          id={id}
          value={value}
          spellCheck={false}
          placeholder={`${placeholder}\n\n可直接粘贴，或点击「选择文件」/ 拖拽文件到此处`}
          className="bg-background placeholder:text-muted-foreground/60 block h-28 w-full resize-none p-2.5 font-mono text-[11px] leading-relaxed outline-hidden"
          onChange={(event) => onChange(event.target.value)}
        />
      </div>
      {error && (
        <p className="border-destructive/30 bg-destructive/5 text-destructive rounded-md border px-2 py-1 text-[11px]">
          {error}
        </p>
      )}
    </div>
  )
}

function PreviewField({
  label,
  value,
  mono,
}: {
  label: string
  value?: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0 text-xs">
      <span className="text-muted-foreground">{label}：</span>
      <span className={cn("break-all", mono && "font-mono")}>
        {value && value.trim() ? value : "—"}
      </span>
    </div>
  )
}

function CertificateCard({
  cert,
  onEdit,
  onDelete,
}: {
  cert: Certificate
  onEdit: () => void
  onDelete: () => void
}) {
  const names = namesOf(cert)
  const days = daysLeft(cert.not_after)
  const shown = names.slice(0, MAX_NAMES)

  return (
    <article className="group flex min-w-0 flex-col gap-3 rounded-md border bg-card p-3 transition-colors hover:bg-accent/20">
      <div className="flex min-w-0 items-start justify-between gap-2">
        <div className="flex min-w-0 flex-col gap-0.5">
          <TooltipHint content={cert.common_name || "未命名证书"}>
            <h2 className="min-w-0 truncate text-sm font-semibold">
              {cert.common_name || "未命名证书"}
            </h2>
          </TooltipHint>
          <TooltipHint content={cert.certificate_sha256 || "—"}>
            <span className="text-muted-foreground truncate font-mono text-[10px]">
              {shortHash(cert.certificate_sha256)}
            </span>
          </TooltipHint>
        </div>
        <div className="flex shrink-0 items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
          <TooltipHint content="编辑">
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              aria-label="编辑证书"
              onClick={onEdit}
            >
              <Pencil />
            </Button>
          </TooltipHint>
          <TooltipHint content="删除">
            <Button
              variant="ghost"
              size="icon"
              className="text-destructive size-7"
              aria-label="删除证书"
              onClick={onDelete}
            >
              <Trash2 />
            </Button>
          </TooltipHint>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-x-4 gap-y-2">
        <CertField
          label="签发者"
          value={
            cert.issuer?.common_name ||
            cert.issuer?.organization?.join(", ") ||
            "—"
          }
        />
        <CertField label="签名算法" value={cert.signature_algorithm || "—"} />
        <CertField
          label="生效时间"
          value={cert.not_before ? formatDateTime(cert.not_before) : "—"}
        />
        <CertField
          label="过期时间"
          value={cert.not_after ? formatDateTime(cert.not_after) : "—"}
          suffix={
            days === null ? undefined : (
              <span
                className={cn(
                  "shrink-0 rounded-full border px-1.5 py-0.5 text-[10px] leading-none font-medium",
                  expiryChipTone(days)
                )}
              >
                {days < 0 ? "已过期" : `${days} 天`}
              </span>
            )
          }
        />
      </div>

      {shown.length > 0 && (
        <div className="flex min-w-0 flex-wrap items-center gap-1">
          {shown.map((name) => (
            <Badge
              key={name}
              variant="secondary"
              className="max-w-40 truncate text-[10px] font-normal"
            >
              {name}
            </Badge>
          ))}
          {names.length > shown.length && (
            <Badge variant="outline" className="text-[10px] font-normal">
              +{names.length - shown.length}
            </Badge>
          )}
        </div>
      )}

      <div className="text-muted-foreground flex items-center justify-end gap-2 border-t pt-2 text-[11px]">
        <span className="shrink-0">
          {cert.updated_at ? formatDateTime(cert.updated_at) : "—"}
        </span>
      </div>
    </article>
  )
}

function CertField({
  label,
  value,
  suffix,
}: {
  label: string
  value: string
  suffix?: ReactNode
}) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <div className="flex min-w-0 items-center gap-1.5 text-xs">
        <TooltipHint content={value}>
          <span className="min-w-0 truncate">{value}</span>
        </TooltipHint>
        {suffix}
      </div>
    </div>
  )
}

function expiryChipTone(days: number): string {
  if (days < 0) {
    return "border-red-500/40 bg-red-500/10 text-red-600 dark:text-red-400"
  }
  if (days <= 15) {
    return "border-amber-500/40 bg-amber-500/10 text-amber-600 dark:text-amber-400"
  }
  return "border-green-500/40 bg-green-500/10 text-green-600 dark:text-green-400"
}

function namesOf(cert: Certificate): string[] {
  return [...(cert.dns_names ?? []), ...(cert.ip_addresses ?? [])].filter(
    Boolean
  )
}

function daysLeft(notAfter?: string): number | null {
  if (!notAfter) return null
  const ts = Date.parse(notAfter)
  if (Number.isNaN(ts)) return null
  return Math.ceil((ts - Date.now()) / 86_400_000)
}

function shortHash(value?: string): string {
  if (!value) return "—"
  const head = value.slice(0, 16)
  return value.length > 16 ? `${head}…` : head
}

function fieldErrors(error: z.ZodError): Record<string, string> {
  const ret: Record<string, string> = {}
  for (const issue of error.issues) {
    const key = issue.path[0]
    if (typeof key === "string" && !ret[key]) ret[key] = issue.message
  }
  return ret
}