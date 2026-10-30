import { useCallback, useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import {
  Database,
  Eye,
  EyeOff,
  Link2,
  Pencil,
  Power,
  Save,
  Table2,
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
import { cmdbConfigApi } from "@/features/cmdb-config/api"
import type { CMDBConfig, CMDBConfigForm } from "@/features/cmdb-config/types"

const CMDB_ENDPOINT_DEFAULT = "https://yunwei.eastmoney.com/cmdb2/api/v0.1/ci/s"

const schema = z.object({
  endpoint: z
    .string()
    .trim()
    .min(1, "地址不能为空")
    .refine(
      (value) => /^https?:\/\/[^\s]+$/i.test(value),
      "请输入有效的 HTTP/HTTPS 地址"
    ),
  key: z.string().trim().min(1, "Key 不能为空").max(100, "不能超过 100 字"),
  secret: z.string().trim().min(1, "Secret 不能为空").max(100, "不能超过 100 字"),
})

function emptyForm(): CMDBConfigForm {
  return {
    endpoint: "",
    enabled: false,
    key: "",
    secret: "",
  }
}

function toForm(config: CMDBConfig | null): CMDBConfigForm {
  if (!config) return emptyForm()
  return {
    endpoint: config.endpoint ?? "",
    enabled: config.enabled ?? false,
    key: config.key ?? "",
    secret: config.secret ?? "",
  }
}

export function CMDBConfigPage() {
  const navigate = useNavigate()
  const [config, setConfig] = useState<CMDBConfig | null>(null)
  const [form, setForm] = useState<CMDBConfigForm>(emptyForm)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [showSecret, setShowSecret] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await cmdbConfigApi.get()
      setConfig(data)
      setForm(toForm(data))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载 CMDB 配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (patch: Partial<CMDBConfigForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const startEdit = () => {
    const next = toForm(config)
    setForm({
      ...next,
      enabled: config ? next.enabled : true,
      endpoint: next.endpoint || CMDB_ENDPOINT_DEFAULT,
    })
    setErrors({})
    setShowSecret(false)
    setEditing(true)
  }

  const cancelEdit = () => {
    setForm(toForm(config))
    setErrors({})
    setShowSecret(false)
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

    setSaving(true)
    try {
      await cmdbConfigApi.update({ ...form, ...parsed.data })
      toast.success("配置已保存")
      setShowSecret(false)
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存 CMDB 配置失败")
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
              <div className="bg-muted h-3.5 w-24 rounded" />
              <div className="bg-muted h-2.5 w-40 rounded" />
            </div>
          </div>
          <div className="h-52 rounded-md border bg-card" />
        </div>
      </div>
    )
  }

  const enabled = editing ? form.enabled : Boolean(config?.enabled)
  const status = !config && !editing ? "empty" : enabled ? "enabled" : "disabled"

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
              <Database className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-semibold">CMDB</h1>
                <StatusPill state={status} />
              </div>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <span>CMDB 资产数据接入配置</span>
                <InfoTooltip label="CMDB 配置说明">
                  <p>用于配置管理中心端访问 CMDB 资产数据的地址与凭证。</p>
                  <p>Key 与 Secret 仅由服务端使用，不会下发到浏览器以外的系统。</p>
                  <p>地址需保证管理中心端可以正常访问。</p>
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
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate("/configs/cmdb/browse")}
                >
                  <Table2 /> 数据浏览
                </Button>
                <Button variant="outline" size="sm" onClick={startEdit}>
                  <Pencil /> {config ? "编辑" : "配置"}
                </Button>
              </>
            )}
          </div>
        </div>

        {!editing && !config ? (
          <div className="flex flex-col items-center gap-3 rounded-md border border-dashed bg-card px-6 py-14 text-center">
            <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-full">
              <Database className="size-6" />
            </div>
            <div className="grid gap-1">
              <h2 className="text-sm font-semibold">尚未配置 CMDB</h2>
              <p className="text-muted-foreground text-xs">
                配置 CMDB 服务地址与访问凭证后，管理中心端即可读取资产数据。
              </p>
            </div>
            <div className="text-muted-foreground flex flex-wrap justify-center gap-1.5 text-[11px]">
              {["填写服务地址", "配置访问凭证", "保存配置"].map((step, index) => (
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
            <Section icon={Link2} title="连接配置" description="CMDB 服务地址与访问凭证">
              {editing ? (
                <div className="grid gap-4 p-4 sm:grid-cols-2">
                  <div className="grid gap-1.5 sm:col-span-2">
                    <Label htmlFor="cmdb-endpoint">服务地址</Label>
                    <Input
                      id="cmdb-endpoint"
                      value={form.endpoint}
                      placeholder="https://cmdb.example.com"
                      onChange={(event) => update({ endpoint: event.target.value })}
                    />
                    <p className="text-muted-foreground text-[11px]">
                      需保证管理中心端可以正常访问
                    </p>
                    <FieldError message={errors.endpoint} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="cmdb-key">Key</Label>
                    <Input
                      id="cmdb-key"
                      value={form.key}
                      placeholder="CMDB 访问 Key"
                      autoComplete="off"
                      onChange={(event) => update({ key: event.target.value })}
                    />
                    <FieldError message={errors.key} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="cmdb-secret">Secret</Label>
                    <div className="relative">
                      <Input
                        id="cmdb-secret"
                        type={showSecret ? "text" : "password"}
                        value={form.secret}
                        placeholder="CMDB 访问 Secret"
                        autoComplete="new-password"
                        className="pr-9"
                        onChange={(event) => update({ secret: event.target.value })}
                      />
                      <button
                        type="button"
                        aria-label={showSecret ? "隐藏 Secret" : "显示 Secret"}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 flex size-6 -translate-y-1/2 items-center justify-center rounded"
                        onClick={() => setShowSecret((visible) => !visible)}
                      >
                        {showSecret ? (
                          <EyeOff className="size-3.5" />
                        ) : (
                          <Eye className="size-3.5" />
                        )}
                      </button>
                    </div>
                    <FieldError message={errors.secret} />
                  </div>

                  <div
                    className={cn(
                      "flex items-center justify-between gap-3 rounded-md border px-3.5 py-3 transition-colors sm:col-span-2",
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
                        <div className="text-sm font-medium">启用 CMDB</div>
                        <p className="text-muted-foreground mt-0.5 text-xs">
                          关闭后管理中心端不再读取 CMDB 资产数据
                        </p>
                      </div>
                    </div>
                    <Switch
                      checked={form.enabled}
                      label="启用 CMDB"
                      onCheckedChange={(value) => update({ enabled: value })}
                    />
                  </div>
                </div>
              ) : (
                <div className="divide-y px-4">
                  <ReadonlyRow label="服务地址" value={config?.endpoint} mono />
                  <ReadonlyRow label="Key" value={config?.key} mono />
                  <ReadonlyRow
                    label="Secret"
                    value={
                      config?.secret
                        ? showSecret
                          ? config.secret
                          : "••••••••"
                        : undefined
                    }
                    mono
                    action={
                      config?.secret ? (
                        <TooltipHint content={showSecret ? "隐藏 Secret" : "显示 Secret"}>
                          <button
                            type="button"
                            aria-label={showSecret ? "隐藏 Secret" : "显示 Secret"}
                            className="text-muted-foreground hover:text-foreground flex size-5 shrink-0 items-center justify-center rounded"
                            onClick={() => setShowSecret((visible) => !visible)}
                          >
                            {showSecret ? (
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
        )}
      </div>
    </div>
  )
}

function Section({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: typeof Link2
  title: string
  description?: string
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
              <p className="text-muted-foreground truncate text-[11px]">
                {description}
              </p>
            )}
          </div>
        </div>
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
            : state === "empty"
              ? "bg-amber-500"
              : "bg-muted-foreground/50"
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