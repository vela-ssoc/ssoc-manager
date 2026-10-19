import {
  useEffect,
  useRef,
  useState,
  type ChangeEvent,
} from "react"
import { z } from "zod"
import { RefreshCw, UploadCloud } from "lucide-react"

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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

import type { ZipArchive, ZipArchiveForm } from "./types"
import { METHOD_OPTIONS, NO_ENCRYPTION } from "./types"

const formSchema = z
  .object({
    name: z.string().trim().min(1, "名称不能为空").max(100, "名称不能超过 100 字"),
    description: z.string().trim().max(1000, "描述不能超过 1000 字"),
    method: z.number().int().min(0, "加密方式无效").max(4, "加密方式无效"),
    password: z.string().max(100, "密码不能超过 100 字"),
  })
  .superRefine((val, ctx) => {
    if (val.method !== NO_ENCRYPTION && val.password.trim() === "") {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["password"],
        message: "选择加密方式后必须填写密码",
      })
    }
  })

function emptyForm(): ZipArchiveForm {
  return { name: "", description: "", method: NO_ENCRYPTION, password: "" }
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null
  return <p className="text-destructive text-xs">{message}</p>
}

function randomPassword(length = 10): string {
  const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
  let result = ""
  for (let i = 0; i < length; i++) {
    result += chars[Math.floor(Math.random() * chars.length)]
  }
  return result
}

function PasswordInput({
  id,
  value,
  onChange,
}: {
  id: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className="relative">
      <Input
        id={id}
        value={value}
        placeholder="加密密码"
        autoComplete="new-password"
        className="pr-9"
        onChange={(e) => onChange(e.target.value)}
      />
      <TooltipHint content="随机生成密码">
        <button
          type="button"
          aria-label="随机生成密码"
          className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 flex size-5 -translate-y-1/2 items-center justify-center rounded"
          onClick={() => onChange(randomPassword())}
        >
          <RefreshCw className="size-3.5" />
        </button>
      </TooltipHint>
    </div>
  )
}

function MethodSelect({
  value,
  onChange,
}: {
  value: number
  onChange: (method: number) => void
}) {
  return (
    <Select
      value={String(value)}
      onValueChange={(val) => onChange(Number(val))}
    >
      <SelectTrigger className="w-full">
        <SelectValue placeholder="选择加密方式" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={String(NO_ENCRYPTION)}>不加密</SelectItem>
        {METHOD_OPTIONS.map((opt) => (
          <SelectItem key={opt.value} value={String(opt.value)}>
            {opt.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

type CreateMode = "upload" | "manual"

export function CreateArchiveDialog({
  open,
  onOpenChange,
  submitting,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  submitting: boolean
  onSubmit: (form: ZipArchiveForm, file?: File) => void
}) {
  const [mode, setMode] = useState<CreateMode>("upload")
  const [form, setForm] = useState<ZipArchiveForm>(emptyForm)
  const [file, setFile] = useState<File | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    setMode("upload")
    setForm(emptyForm())
    setFile(null)
    setErrors({})
    if (fileInputRef.current) fileInputRef.current.value = ""
  }, [open])

  const update = (patch: Partial<ZipArchiveForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const handlePickFile = (e: ChangeEvent<HTMLInputElement>) => {
    const picked = e.target.files?.[0] ?? null
    setFile(picked)
    if (picked) {
      const base = picked.name.replace(/\.zip$/i, "")
      if (base && form.name.trim() === "") {
        setForm((prev) => ({ ...prev, name: base }))
      }
    }
  }

  const handleSave = () => {
    const parsed = formSchema.safeParse(form)
    if (!parsed.success) {
      const next: Record<string, string> = {}
      for (const issue of parsed.error.issues) {
        const key = String(issue.path[0] ?? "")
        if (!next[key]) next[key] = issue.message
      }
      setErrors(next)
      return
    }
    if (mode === "upload" && !file) {
      setErrors((prev) => ({ ...prev, file: "请选择要上传的 ZIP 文件" }))
      return
    }
    onSubmit(parsed.data, mode === "upload" ? file ?? undefined : undefined)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>新建归档</DialogTitle>
        </DialogHeader>

        <div className="bg-muted flex gap-1 rounded-md p-1">
          <button
            type="button"
            className={`flex-1 rounded px-3 py-1.5 text-sm transition-colors ${
              mode === "upload"
                ? "bg-card font-medium shadow-sm"
                : "text-muted-foreground"
            }`}
            onClick={() => setMode("upload")}
          >
            上传 ZIP
          </button>
          <button
            type="button"
            className={`flex-1 rounded px-3 py-1.5 text-sm transition-colors ${
              mode === "manual"
                ? "bg-card font-medium shadow-sm"
                : "text-muted-foreground"
            }`}
            onClick={() => setMode("manual")}
          >
            手动创建
          </button>
        </div>

        <div className="grid gap-4">
          {mode === "upload" && (
            <div className="grid gap-2">
              <Label>ZIP 文件</Label>
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="flex flex-col items-center justify-center gap-1 rounded-md border border-dashed px-4 py-5 text-sm transition-colors hover:bg-accent/40"
              >
                <UploadCloud className="text-muted-foreground size-5" />
                {file ? (
                  <span className="text-xs">{file.name}</span>
                ) : (
                  <span className="text-muted-foreground text-xs">
                    点击选择 .zip 文件
                  </span>
                )}
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept=".zip"
                className="hidden"
                onChange={handlePickFile}
              />
              <FieldError message={errors.file} />
            </div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="za-name">名称</Label>
            <Input
              id="za-name"
              value={form.name}
              placeholder="归档名称（不含 .zip）"
              onChange={(e) => update({ name: e.target.value })}
            />
            <FieldError message={errors.name} />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="za-desc">描述</Label>
            <Textarea
              id="za-desc"
              value={form.description}
              placeholder="说明描述信息"
              rows={3}
              onChange={(e) => update({ description: e.target.value })}
            />
            <FieldError message={errors.description} />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="za-method">加密方式</Label>
            <MethodSelect
              value={form.method}
              onChange={(method) =>
                update(method === NO_ENCRYPTION ? { method, password: "" } : { method })
              }
            />
            <FieldError message={errors.method} />
          </div>

          {form.method !== NO_ENCRYPTION && (
            <div className="grid gap-2">
              <Label htmlFor="za-pass">密码</Label>
              <PasswordInput
                id="za-pass"
                value={form.password}
                onChange={(value) => update({ password: value })}
              />
              <FieldError message={errors.password} />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={submitting}>
            {submitting ? "创建中…" : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function ArchiveFormDialog({
  open,
  onOpenChange,
  initial,
  submitting,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  initial?: ZipArchive
  submitting: boolean
  onSubmit: (form: ZipArchiveForm, id?: string) => void
}) {
  const [form, setForm] = useState<ZipArchiveForm>(emptyForm)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!open) return
    setForm(
      initial
        ? {
          name: initial.name,
          description: initial.description,
          method: initial.method,
          password: initial.password,
        }
        : emptyForm()
    )
    setErrors({})
  }, [open, initial])

  const update = (patch: Partial<ZipArchiveForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const handleSave = () => {
    const parsed = formSchema.safeParse(form)
    if (!parsed.success) {
      const next: Record<string, string> = {}
      for (const issue of parsed.error.issues) {
        const key = String(issue.path[0] ?? "")
        if (!next[key]) next[key] = issue.message
      }
      setErrors(next)
      return
    }
    onSubmit(parsed.data, initial?.id)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑归档</DialogTitle>
          <DialogDescription>修改归档的基本信息（名称、描述、加密配置）</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="za-name">名称</Label>
            <Input
              id="za-name"
              value={form.name}
              placeholder="归档名称（不含 .zip）"
              disabled={!!initial}
              onChange={(e) => update({ name: e.target.value })}
            />
            <FieldError message={errors.name} />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="za-desc">描述</Label>
            <Textarea
              id="za-desc"
              value={form.description}
              placeholder="说明描述信息"
              rows={3}
              onChange={(e) => update({ description: e.target.value })}
            />
            <FieldError message={errors.description} />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="za-method">加密方式</Label>
            <MethodSelect
              value={form.method}
              onChange={(method) =>
                update(method === NO_ENCRYPTION ? { method, password: "" } : { method })
              }
            />
            <FieldError message={errors.method} />
          </div>

          {form.method !== NO_ENCRYPTION && (
            <div className="grid gap-2">
              <Label htmlFor="za-pass">密码</Label>
              <PasswordInput
                id="za-pass"
                value={form.password}
                onChange={(value) => update({ password: value })}
              />
              <FieldError message={errors.password} />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={submitting}>
            {submitting ? "保存中…" : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
