import { useEffect, useMemo, useRef, useState } from "react"
import { z } from "zod"
import { ExternalLink, Plus, Trash2 } from "lucide-react"

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
import { TooltipHint } from "@/components/tooltip-hint"
import { cn } from "@/lib/utils"
import { toast } from "@/lib/toast"

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
  location: z
    .string()
    .trim()
    .min(1, "路径不能为空")
    .regex(/^[^/]+$/, "路径不能包含 /"),
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

  const savedLocation = initial?.location ?? ""
  const savedManager = initial?.manager ?? false

  const managerActive = savedManager && Boolean(savedLocation)
  const managerUrl = `${typeof window !== "undefined" ? window.location.origin : ""}/api/v3/reverse/${form.location}`
  const brokerPrefix = `/api/v3/reverse/${form.location}`

  const guardVisit = () => {
    if (managerActive) {
      window.open(managerUrl, "_blank", "noopener,noreferrer")
    } else {
      toast.error("该配置尚未提交生效，请先保存")
    }
  }

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
              placeholder="例如 your-service-name"
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

            <div className="overflow-hidden rounded-md border">
              <table className="w-full table-fixed text-sm">
                <thead>
                  <tr className="bg-muted/40 text-muted-foreground border-b text-xs">
                    <th className="w-20 px-3 py-2 text-left font-medium">服务</th>
                    <th className="w-14 px-3 py-2 text-left font-medium">启用</th>
                    <th className="px-3 py-2 text-left font-medium">访问地址</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b last:border-b-0">
                    <td className="px-3 py-2">中心端</td>
                    <td className="px-3 py-2">
                      <SwitchControl
                        label="启用 Manager 节点"
                        checked={form.manager}
                        onChange={(manager) => update({ manager })}
                      />
                    </td>
                    <td className="px-3 py-2">
                      {form.manager && form.location ? (
                        <div className="flex items-center gap-1.5">
                          <TooltipHint content={managerUrl}>
                            <span className="text-muted-foreground min-w-0 flex-1 truncate text-xs">
                              {managerUrl}
                            </span>
                          </TooltipHint>
                          <button
                            type="button"
                            aria-label="访问"
                            onClick={guardVisit}
                            className={cn(
                              "flex size-6 shrink-0 items-center justify-center rounded",
                              managerActive
                                ? "text-primary hover:bg-accent hover:text-accent-foreground cursor-pointer"
                                : "text-muted-foreground/50 cursor-not-allowed"
                            )}
                          >
                            <ExternalLink className="size-3.5" />
                          </button>
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-xs">不可访问</span>
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td className="px-3 py-2">中继节点</td>
                    <td className="px-3 py-2">
                      <SwitchControl
                        label="启用 Broker 节点"
                        checked={form.broker}
                        onChange={(broker) => update({ broker })}
                      />
                    </td>
                    <td className="px-3 py-2">
                      {form.broker && form.location ? (
                        <TooltipHint content={`虚拟通道内访问，路径前缀 ${brokerPrefix}`}>
                          <span className="text-muted-foreground block truncate text-xs">
                            虚拟通道内访问，路径前缀 {brokerPrefix}
                          </span>
                        </TooltipHint>
                      ) : (
                        <span className="text-muted-foreground text-xs">不可访问</span>
                      )}
                    </td>
                  </tr>
                </tbody>
              </table>
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
                  <TooltipHint content="移除 Header">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-8 text-muted-foreground hover:text-destructive"
                      aria-label="移除 Header"
                      onClick={() =>
                        setHeaderRows((rows) => rows.filter((_, i) => i !== index))
                      }
                    >
                      <Trash2 />
                    </Button>
                  </TooltipHint>
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
