import { useCallback, useEffect, useState } from "react"
import {
  Activity,
  Eye,
  EyeOff,
  Link2,
  Pencil,
  Plus,
  Power,
  Save,
  Server,
  Tags,
  Trash2,
  X,
} from "lucide-react"
import { z } from "zod"

import { InfoTooltip } from "@/components/info-tooltip"
import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { victoriaMetricsApi } from "@/features/victoria-metrics/api"
import type {
  VictoriaMetricsConfig,
  VictoriaMetricsConfigForm,
} from "@/features/victoria-metrics/types"

interface HeaderRow {
  key: string
  value: string
}

const schema = z.object({
  address: z
    .string()
    .trim()
    .min(1, "地址不能为空")
    .refine(
      (value) => /^https?:\/\/[^\s]+$/i.test(value),
      "请输入有效的 HTTP/HTTPS 地址"
    ),
  tenant_id: z.string().trim().max(20, "租户 ID 不能超过 20 字"),
  username: z.string().max(100, "用户名不能超过 100 字"),
  password: z.string().max(100, "密码不能超过 100 字"),
})

function emptyForm(): VictoriaMetricsConfigForm {
  return {
    enabled: false,
    address: "",
    tenant_id: "",
    username: "",
    password: "",
    headers: {},
  }
}

function toForm(config: VictoriaMetricsConfig | null): VictoriaMetricsConfigForm {
  if (!config) return emptyForm()
  return {
    enabled: config.enabled ?? false,
    address: config.address ?? "",
    tenant_id: config.tenant_id ?? "",
    username: config.username ?? "",
    password: config.password ?? "",
    headers: config.headers ?? {},
  }
}

function toRows(headers: Record<string, string>): HeaderRow[] {
  return Object.entries(headers).map(([key, value]) => ({ key, value }))
}

function pushUrl(address: string, tenant: string): string {
  const base = address.trim().replace(/\/+$/, "")
  if (!base) return "—"
  const prefix = tenant.trim() ? `/insert/${tenant.trim()}/prometheus` : ""
  return `${base}${prefix}/api/v1/import/prometheus`
}

export function VictoriaMetricsPage() {
  const [config, setConfig] = useState<VictoriaMetricsConfig | null>(null)
  const [form, setForm] = useState<VictoriaMetricsConfigForm>(emptyForm)
  const [headerRows, setHeaderRows] = useState<HeaderRow[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [showPassword, setShowPassword] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await victoriaMetricsApi.get()
      setConfig(data)
      setForm(toForm(data))
      setHeaderRows(toRows(data?.headers ?? {}))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载 VictoriaMetrics 配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (patch: Partial<VictoriaMetricsConfigForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const startEdit = () => {
    setForm(toForm(config))
    setHeaderRows(toRows(config?.headers ?? {}))
    setErrors({})
    setShowPassword(false)
    setEditing(true)
  }

  const cancelEdit = () => {
    setForm(toForm(config))
    setHeaderRows(toRows(config?.headers ?? {}))
    setErrors({})
    setShowPassword(false)
    setEditing(false)
  }

  const save = async () => {
    const parsed = schema.safeParse(form)
    if (!parsed.success) {
      const next: Record<string, string> = {}
      for (const issue of parsed.error.issues) {
        const key = String(issue.path[0] ?? "")
        if (!next[key]) next[key] = issue.message
      }
      setErrors(next)
      return
    }

    if (headerRows.some((row) => !row.key.trim() && row.value.trim())) {
      toast.error("Header 名称不能为空")
      return
    }

    const headers = Object.fromEntries(
      headerRows
        .filter((row) => row.key.trim())
        .map((row) => [row.key.trim(), row.value] as const)
    )
    if (Object.keys(headers).length > 100) {
      toast.error("Header 数量不能超过 100 个")
      return
    }

    setSaving(true)
    try {
      await victoriaMetricsApi.update({ ...form, ...parsed.data, headers })
      toast.success("配置已保存")
      setShowPassword(false)
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存 VictoriaMetrics 配置失败")
    } finally {
      setSaving(false)
    }
  }

  if (loading && !config) {
    return (
      <div className="h-full overflow-y-auto p-4">
        <div className="mx-auto flex w-full max-w-4xl animate-pulse flex-col gap-4">
          <div className="flex items-center gap-3 rounded-md border bg-card px-4 py-3">
            <div className="bg-muted size-10 shrink-0 rounded-md" />
            <div className="grid gap-2">
              <div className="bg-muted h-3.5 w-32 rounded" />
              <div className="bg-muted h-2.5 w-44 rounded" />
            </div>
          </div>
          <div className="flex flex-col gap-4">
            <div className="h-52 rounded-md border bg-card" />
            <div className="h-32 rounded-md border bg-card" />
          </div>
        </div>
      </div>
    )
  }

  const enabled = editing ? form.enabled : Boolean(config?.enabled)
  const status = !config && !editing ? "empty" : enabled ? "enabled" : "disabled"
  const headers = config?.headers ?? {}
  const pushAddress = pushUrl(form.address, form.tenant_id)

  return (
    <div className="h-full overflow-y-auto p-4">
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-4">
        <div
          className={cn(
            "relative flex flex-wrap items-center justify-between gap-3 overflow-hidden rounded-md border bg-card px-4 py-3 transition-colors",
            status === "enabled" && "border-primary/30"
          )}
        >
          <span
            className={cn(
              "absolute inset-y-0 left-0 w-0.5 transition-colors",
              status === "enabled"
                ? "bg-primary"
                : status === "empty"
                  ? "bg-amber-500/60"
                  : "bg-border"
            )}
          />
          <div className="flex min-w-0 items-center gap-3">
            <div
              className={cn(
                "flex size-10 shrink-0 items-center justify-center rounded-md transition-colors",
                status === "enabled"
                  ? "bg-primary/10 text-primary"
                  : status === "empty"
                    ? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                    : "bg-muted text-muted-foreground"
              )}
            >
              <Activity className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-semibold">
                  VictoriaMetrics
                </h1>
                <StatusPill state={status} />
              </div>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <span>管理中心端运行指标推送配置</span>
                <InfoTooltip label="VictoriaMetrics 配置说明">
                  <p>管理中心端会将自身运行指标推送到该 VictoriaMetrics 服务。</p>
                  <p>不填写租户 ID 时视为单机版，直接使用服务地址推送。</p>
                  <p>
                    填写租户 ID 时视为集群版，后端会自动拼接
                    <code className="bg-muted mx-1 rounded px-1 py-0.5 font-mono text-[10px]">
                      insert/&#123;tenant&#125;/prometheus
                    </code>
                    写入路径。
                  </p>
                  <p>用户名和密码会以 Basic Auth 方式加入请求，无需重复配置 Authorization Header。</p>
                </InfoTooltip>
              </div>
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-2">
            {editing ? (
              <>
                <Button variant="outline" size="sm" disabled={saving} onClick={cancelEdit}>
                  <X /> 取消
                </Button>
                <Button size="sm" disabled={saving} onClick={() => void save()}>
                  <Save /> {saving ? "保存中…" : "保存"}
                </Button>
              </>
            ) : (
              <Button variant="outline" size="sm" onClick={startEdit}>
                <Pencil /> 编辑
              </Button>
            )}
          </div>
        </div>

        {!editing && !config ? (
          <div className="flex flex-col items-center gap-3 rounded-md border border-dashed bg-card px-6 py-14 text-center">
            <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-full">
              <Activity className="size-6" />
            </div>
            <div className="grid gap-1">
              <h2 className="text-sm font-semibold">尚未配置 VictoriaMetrics</h2>
              <p className="text-muted-foreground text-xs">
                配置服务地址后，管理中心端即可推送运行指标。
              </p>
            </div>
            <div className="text-muted-foreground flex flex-wrap justify-center gap-1.5 text-[11px]">
              {["填写服务地址", "配置认证方式", "启用并保存"].map((step, index) => (
                <span
                  key={step}
                  className="bg-muted/40 flex items-center gap-1.5 rounded-full border px-2.5 py-1"
                >
                  <span className="bg-background flex size-4 items-center justify-center rounded-full border text-[10px] font-medium">
                    {index + 1}
                  </span>
                  {step}
                </span>
              ))}
            </div>
            <Button size="sm" onClick={startEdit}>
              <Pencil /> 立即配置
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="flex min-w-0 flex-col gap-4">
              <Section icon={Server} title="基础配置">
                {editing ? (
                  <div className="grid gap-4 p-4 sm:grid-cols-3">
                    <div className="grid gap-1.5 sm:col-span-3">
                      <Label htmlFor="vm-address">服务地址</Label>
                      <Input
                        id="vm-address"
                        value={form.address}
                        placeholder="http://127.0.0.1:8428"
                        onChange={(event) => update({ address: event.target.value })}
                      />
                      <p className="text-muted-foreground text-[11px]">
                        单机版填写服务地址；集群版填写 vminsert 地址
                      </p>
                      <FieldError message={errors.address} />
                    </div>

                    <div className="grid gap-1.5">
                      <Label htmlFor="vm-tenant">租户 ID</Label>
                      <Input
                        id="vm-tenant"
                        value={form.tenant_id}
                        placeholder="留空为单机版，集群版填写例如 0"
                        onChange={(event) => update({ tenant_id: event.target.value })}
                      />
                      <FieldError message={errors.tenant_id} />
                    </div>

                    <div className="grid gap-1.5">
                      <Label htmlFor="vm-username">用户名</Label>
                      <Input
                        id="vm-username"
                        value={form.username}
                        placeholder="Basic Auth 用户名（可选）"
                        autoComplete="off"
                        onChange={(event) => update({ username: event.target.value })}
                      />
                      <FieldError message={errors.username} />
                    </div>

                    <div className="grid gap-1.5">
                      <Label htmlFor="vm-password">密码</Label>
                      <div className="relative">
                        <Input
                          id="vm-password"
                          type={showPassword ? "text" : "password"}
                          value={form.password}
                          placeholder="Basic Auth 密码（可选）"
                          autoComplete="new-password"
                          className="pr-9"
                          onChange={(event) => update({ password: event.target.value })}
                        />
                        <button
                          type="button"
                          aria-label={showPassword ? "隐藏密码" : "显示密码"}
                          className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 flex size-6 -translate-y-1/2 items-center justify-center rounded"
                          onClick={() => setShowPassword((visible) => !visible)}
                        >
                          {showPassword ? (
                            <EyeOff className="size-3.5" />
                          ) : (
                            <Eye className="size-3.5" />
                          )}
                        </button>
                      </div>
                      <FieldError message={errors.password} />
                    </div>

                    <div
                      className={cn(
                        "flex items-center justify-between gap-3 rounded-md border px-3.5 py-3 transition-colors sm:col-span-3",
                        form.enabled
                          ? "border-primary/25 bg-primary/5"
                          : "bg-muted/20"
                      )}
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <span
                          className={cn(
                            "flex size-8 shrink-0 items-center justify-center rounded-md transition-colors",
                            form.enabled
                              ? "bg-primary/10 text-primary"
                              : "bg-muted text-muted-foreground"
                          )}
                        >
                          <Power className="size-4" />
                        </span>
                        <div className="min-w-0">
                          <div className="text-sm font-medium">启用推送</div>
                          <p className="text-muted-foreground mt-0.5 text-xs">
                            关闭后管理中心端不再向 VictoriaMetrics 推送运行指标
                          </p>
                        </div>
                      </div>
                      <Switch
                        checked={form.enabled}
                        label="启用 VictoriaMetrics 推送"
                        onCheckedChange={(value) => update({ enabled: value })}
                      />
                    </div>
                  </div>
                ) : (
                  <div className="divide-y px-4">
                    <ReadonlyRow label="服务地址" value={config?.address} mono />
                    <ReadonlyRow label="租户 ID" value={config?.tenant_id} mono />
                    <ReadonlyRow label="用户名" value={config?.username} />
                    <ReadonlyRow
                      label="密码"
                      value={
                        config?.password
                          ? showPassword
                            ? config.password
                            : "••••••••"
                          : undefined
                      }
                      mono
                      action={
                        config?.password ? (
                          <TooltipHint content={showPassword ? "隐藏密码" : "显示密码"}>
                            <button
                              type="button"
                              aria-label={showPassword ? "隐藏密码" : "显示密码"}
                              className="text-muted-foreground hover:text-foreground flex size-5 shrink-0 items-center justify-center rounded"
                              onClick={() => setShowPassword((visible) => !visible)}
                            >
                              {showPassword ? (
                                <EyeOff className="size-3.5" />
                              ) : (
                                <Eye className="size-3.5" />
                              )}
                            </button>
                          </TooltipHint>
                        ) : undefined
                      }
                    />
                  </div>
                )}
              </Section>

              <Section
                icon={Tags}
                title="HTTP Header"
                description="附加到 VictoriaMetrics 推送请求"
                actions={
                  editing ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setHeaderRows((rows) => [...rows, { key: "", value: "" }])
                      }
                    >
                      <Plus /> 添加
                    </Button>
                  ) : undefined
                }
              >
                {editing ? (
                  <div className="grid gap-2 p-4">
                    {headerRows.map((row, index) => (
                      <div
                        key={index}
                        className="grid grid-cols-[1rem_1fr_1.4fr_auto] items-center gap-2"
                      >
                        <span className="text-muted-foreground text-right text-[11px] tabular-nums">
                          {index + 1}
                        </span>
                        <Input
                          aria-label={`Header ${index + 1} 名称`}
                          value={row.key}
                          placeholder="Header 名称，例如 X-Scope-OrgID"
                          onChange={(event) => {
                            const next = [...headerRows]
                            next[index] = { ...row, key: event.target.value }
                            setHeaderRows(next)
                          }}
                        />
                        <Input
                          aria-label={`Header ${index + 1} 值`}
                          value={row.value}
                          placeholder="值"
                          onChange={(event) => {
                            const next = [...headerRows]
                            next[index] = { ...row, value: event.target.value }
                            setHeaderRows(next)
                          }}
                        />
                        <TooltipHint content="移除 Header">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="text-muted-foreground hover:text-destructive size-8"
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
                    {headerRows.length === 0 && (
                      <div className="flex flex-col items-center gap-2 rounded-md border border-dashed px-4 py-6 text-center">
                        <Tags className="text-muted-foreground size-5 opacity-60" />
                        <span className="text-muted-foreground text-xs">
                          暂无 Header，可按需添加多租户或网关鉴权信息
                        </span>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="p-4">
                    {Object.keys(headers).length === 0 ? (
                      <div className="flex flex-col items-center gap-2 rounded-md border border-dashed px-4 py-6 text-center">
                        <Tags className="text-muted-foreground size-5 opacity-60" />
                        <span className="text-muted-foreground text-xs">暂无 Header</span>
                      </div>
                    ) : (
                      <div className="overflow-hidden rounded-md border">
                        <div className="bg-muted/40 text-muted-foreground grid grid-cols-[1fr_1.5fr] border-b px-3 py-1.5 text-[11px] font-medium">
                          <span>名称</span>
                          <span>值</span>
                        </div>
                        {Object.entries(headers).map(([key, value]) => (
                          <div
                            key={key}
                            className="hover:bg-muted/40 grid grid-cols-[1fr_1.5fr] items-center border-b px-3 py-2 text-xs transition-colors last:border-b-0"
                          >
                            <TooltipHint content={key}>
                              <span className="truncate font-mono">{key}</span>
                            </TooltipHint>
                            <TooltipHint content={value || "—"}>
                              <span className="min-w-0 truncate font-mono">
                                {value || "—"}
                              </span>
                            </TooltipHint>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </Section>

              <Section
                icon={Link2}
                title="写入地址"
                description="指标实际推送的完整地址"
              >
                <div className="p-4">
                  <div className="bg-muted/20 overflow-hidden rounded-md border px-3 py-2">
                    <TooltipHint content={pushAddress}>
                      <div className="text-muted-foreground truncate font-mono text-xs">
                        {pushAddress}
                      </div>
                    </TooltipHint>
                  </div>
                </div>
              </Section>

              <div className="text-muted-foreground flex flex-wrap items-center gap-x-6 gap-y-1 px-1 text-[11px]">
                <span>
                  创建时间{" "}
                  {config?.created_at ? formatDateTime(config.created_at) : "—"}
                </span>
                <span>
                  更新时间{" "}
                  {config?.updated_at ? formatDateTime(config.updated_at) : "—"}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function Section({
  icon: Icon,
  title,
  description,
  actions,
  children,
}: {
  icon: typeof Server
  title: string
  description?: string
  actions?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="overflow-hidden rounded-md border bg-card">
      <div className="flex items-center justify-between gap-2 border-b bg-muted/30 px-4 py-2.5">
        <div className="flex min-w-0 items-center gap-2.5">
          <span className="bg-background text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded border">
            <Icon className="size-3.5" />
          </span>
          <div className="min-w-0">
            <h2 className="truncate text-sm font-semibold">{title}</h2>
            {description && (
              <p className="text-muted-foreground truncate text-[11px]">{description}</p>
            )}
          </div>
        </div>
        {actions}
      </div>
      {children}
    </section>
  )
}

function StatusPill({ state }: { state: "enabled" | "disabled" | "empty" }) {
  const label = state === "enabled" ? "已启用" : state === "disabled" ? "已停用" : "未配置"
  return (
    <span
      className={cn(
        "inline-flex h-5 shrink-0 items-center gap-1.5 rounded-full border px-2 text-[11px] leading-none font-medium",
        state === "enabled"
          ? "border-green-500/30 bg-green-500/10 text-green-600 dark:text-green-400"
          : state === "empty"
            ? "border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400"
            : "bg-muted/40 text-muted-foreground"
      )}
    >
      <span
        className={cn(
          "size-1.5 rounded-full",
          state === "enabled"
            ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
            : state === "disabled"
              ? "bg-muted-foreground/50"
              : "bg-amber-500"
        )}
      />
      {label}
    </span>
  )
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null
  return <p className="text-destructive text-xs">{message}</p>
}

function ReadonlyRow({
  label,
  value,
  mono = false,
  action,
}: {
  label: string
  value?: string
  mono?: boolean
  action?: React.ReactNode
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="flex items-center justify-between gap-4 py-2.5">
      <span className="text-muted-foreground w-24 shrink-0 text-xs">{label}</span>
      <div className="flex min-w-0 flex-1 items-center justify-end gap-1.5">
        <TooltipHint content={text}>
          <span
            className={cn(
              "min-w-0 truncate text-sm",
              mono && "font-mono",
              text === "—" && "text-muted-foreground"
            )}
          >
            {text}
          </span>
        </TooltipHint>
        {action}
      </div>
    </div>
  )
}