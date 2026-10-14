import { useEffect, useMemo, useRef, useState } from "react"
import { z } from "zod"
import { Plus, Trash2 } from "lucide-react"

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
import { InfoTooltip } from "@/components/info-tooltip"

import type { ReverseProxy, ReverseProxyForm } from "./types"

const COMMON_HEADERS = [
  "Accept",
  "Accept-Encoding",
  "Authorization",
  "Cache-Control",
  "Connection",
  "Content-Encoding",
  "Content-Type",
  "Cookie",
  "Forwarded",
  "Host",
  "Origin",
  "Referer",
  "User-Agent",
  "X-API-Key",
  "X-Forwarded-For",
  "X-Forwarded-Host",
  "X-Forwarded-Port",
  "X-Forwarded-Proto",
  "X-Real-IP",
  "X-Request-ID",
]

const schema = z.object({
  location: z.string().trim().min(1, "路径不能为空"),
  name: z.string().trim().min(1, "名称不能为空").max(20, "名称不能超过 20 字"),
  upstream: z.string().trim().url("请输入有效的 HTTP/HTTPS 上游地址"),
})

function emptyForm(): ReverseProxyForm {
  return {
    location: "",
    name: "",
    upstream: "",
    header: {},
    manager: false,
    broker: false,
  }
}

function HeaderKeyInput({
  ariaLabel,
  value,
  onChange,
}: {
  ariaLabel: string
  value: string
  onChange: (value: string) => void
}) {
  const [open, setOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const suggestions = useMemo(
    () =>
      COMMON_HEADERS.filter((header) =>
        header.toLowerCase().includes(value.trim().toLowerCase())
      ).slice(0, 8),
    [value]
  )

  useEffect(() => {
    if (!open) return
    const closeOutside = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener("mousedown", closeOutside)
    return () => document.removeEventListener("mousedown", closeOutside)
  }, [open])

  return (
    <div ref={rootRef} className="relative min-w-0">
      <Input
        aria-label={ariaLabel}
        aria-autocomplete="list"
        value={value}
        placeholder="Header 名称"
        onFocus={() => setOpen(true)}
        onChange={(event) => {
          onChange(event.target.value)
          setOpen(true)
        }}
      />
      {open && suggestions.length > 0 && (
        <div
          role="listbox"
          className="bg-popover text-popover-foreground absolute top-full left-0 z-50 mt-1 max-h-48 w-full overflow-auto rounded-md border p-1 shadow-md"
        >
          {suggestions.map((header) => (
            <button
              key={header}
              type="button"
              role="option"
              aria-selected={header.toLowerCase() === value.trim().toLowerCase()}
              className="hover:bg-accent hover:text-accent-foreground flex w-full cursor-pointer items-center rounded-sm px-2 py-1.5 text-left text-xs"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                onChange(header)
                setOpen(false)
              }}
            >
              {header}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function ReverseProxyDialog({
  open,
  onOpenChange,
  initial,
  submitting,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  initial?: ReverseProxy
  submitting: boolean
  onSubmit: (form: ReverseProxyForm) => void
}) {
  const [form, setForm] = useState<ReverseProxyForm>(emptyForm)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [headerRows, setHeaderRows] = useState<Array<{ key: string; value: string }>>([])

  useEffect(() => {
    if (!open) return
    setForm(
      initial
        ? {
          location: initial.location,
          name: initial.name,
          upstream: initial.upstream,
          header: initial.header ?? {},
          manager: initial.manager,
          broker: initial.broker,
        }
        : emptyForm()
    )
    setHeaderRows(
      initial
        ? Object.entries(initial.header ?? {}).map(([key, value]) => ({ key, value }))
        : []
    )
    setErrors({})
  }, [open, initial])

  const update = (patch: Partial<ReverseProxyForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const save = () => {
    const parsed = schema.safeParse(form)
    if (!parsed.success) {
      const next: Record<string, string> = {}
      for (const issue of parsed.error.issues) {
        const field = String(issue.path[0] ?? "")
        if (!next[field]) next[field] = issue.message
      }
      setErrors(next)
      return
    }
    const header = Object.fromEntries(
      headerRows
        .map(({ key, value }) => [key.trim(), value] as const)
        .filter(([key]) => key.length > 0)
    )
    onSubmit({ ...form, ...parsed.data, header })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{initial ? "编辑反向代理" : "新建反向代理"}</DialogTitle>
          <DialogDescription>
            配置匹配路径、上游地址及启用节点
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <Label htmlFor="proxy-location">路径</Label>
            <Input
              id="proxy-location"
              value={form.location}
              placeholder="例如 /your-service-name"
              disabled={Boolean(initial)}
              onChange={(e) => update({ location: e.target.value })}
            />
            {errors.location && (
              <p className="text-destructive text-xs">{errors.location}</p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="proxy-name">名称</Label>
            <Input
              id="proxy-name"
              value={form.name}
              placeholder="便于用户识别理解的名字"
              onChange={(e) => update({ name: e.target.value })}
            />
            {errors.name && (
              <p className="text-destructive text-xs">{errors.name}</p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="proxy-upstream">上游地址</Label>
            <Input
              id="proxy-upstream"
              value={form.upstream}
              placeholder="https://upstream.example.com"
              onChange={(e) => update({ upstream: e.target.value })}
            />
            {errors.upstream && (
              <p className="text-destructive text-xs">{errors.upstream}</p>
            )}
          </div>

          <div className="grid gap-1.5 pt-1">
            <div className="flex items-center gap-1.5">
              <Label>启用节点</Label>
              <InfoTooltip label="启用节点说明">
                <p>
                  选择需要运行此规则的节点类型。启用后，该节点收到匹配路径的请求时，会将请求转发至上游地址。
                </p>
              </InfoTooltip>
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div className="flex items-center justify-between rounded-md border px-3 py-2">
                <span className="text-sm">Manager 节点</span>
                <SwitchControl
                  label="启用 Manager 节点"
                  checked={form.manager}
                  onChange={(manager) => update({ manager })}
                />
              </div>
              <div className="flex items-center justify-between rounded-md border px-3 py-2">
                <span className="text-sm">Broker 节点</span>
                <SwitchControl
                  label="启用 Broker 节点"
                  checked={form.broker}
                  onChange={(broker) => update({ broker })}
                />
              </div>
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label>HTTP Header</Label>
            <div className="grid gap-2">
              {headerRows.map((row, index) => (
                <div key={index} className="grid grid-cols-[1fr_1.5fr_auto] gap-2">
                  <HeaderKeyInput
                    ariaLabel={`Header ${index + 1} 名称`}
                    value={row.key}
                    onChange={(key) => {
                      const next = [...headerRows]
                      next[index] = { ...row, key }
                      setHeaderRows(next)
                    }}
                  />
                  <Input
                    aria-label={`Header ${index + 1} 值`}
                    value={row.value}
                    placeholder="值"
                    onChange={(e) => {
                      const next = [...headerRows]
                      next[index] = { ...row, value: e.target.value }
                      setHeaderRows(next)
                    }}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-8 text-muted-foreground hover:text-destructive"
                    title="移除 Header"
                    onClick={() =>
                      setHeaderRows((rows) => rows.filter((_, i) => i !== index))
                    }
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="w-fit"
                onClick={() => setHeaderRows((rows) => [...rows, { key: "", value: "" }])}
              >
                <Plus /> 添加 Header
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={save} disabled={submitting}>
            {submitting ? "保存中…" : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function SwitchControl({
  label,
  checked,
  onChange,
}: {
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-label={label}
      aria-checked={checked}
      data-state={checked ? "checked" : "unchecked"}
      className="border-input data-[state=checked]:bg-primary data-[state=unchecked]:bg-input focus-visible:ring-ring inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
      onClick={() => onChange(!checked)}
    >
      <span
        className="bg-background pointer-events-none block size-4 rounded-full shadow-sm transition-transform data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0"
        data-state={checked ? "checked" : "unchecked"}
      />
    </button>
  )
}
