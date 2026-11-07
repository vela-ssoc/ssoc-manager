import { useCallback, useEffect, useState } from "react"
import {
  ChevronDown,
  Eye,
  EyeOff,
  Gauge,
  MemoryStick,
  Pencil,
  Plus,
  Save,
  Server,
  Table2,
  Terminal,
  Trash2,
  X,
} from "lucide-react"
import { z } from "zod"

import {
  ConfigEmptyState,
  ConfigFieldError,
  ConfigReadonlyRow,
  ConfigStatusPill,
} from "@/components/config-panel"
import { InfoTooltip } from "@/components/info-tooltip"
import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { redisConfigApi } from "@/features/redis-config/api"
import { RedisCliPanel } from "@/features/redis-config/cli-panel"
import { RedisInfoPanel } from "@/features/redis-config/info-panel"
import { RedisKeysBrowser } from "@/features/redis-config/keys-browser"
import type {
  RedisConfig,
  RedisConfigForm,
  RedisConfigPayload,
} from "@/features/redis-config/types"

const ADDRESS_RE = /^[^\s:,]+:\d{1,5}$/

type TabKey = "info" | "keys" | "cli"

const TABS: Array<{ value: TabKey; label: string; icon: typeof Gauge }> = [
  { value: "info", label: "运行信息", icon: Gauge },
  { value: "keys", label: "数据浏览", icon: Table2 },
  { value: "cli", label: "命令行", icon: Terminal },
]

const schema = z
  .object({
    clusterMode: z.boolean(),
    addresses: z
      .array(z.string().trim())
      .min(1, "至少填写一个地址")
      .refine(
        (list) => list.every((item) => ADDRESS_RE.test(item)),
        "地址必须是 host:port 形式，例如 127.0.0.1:6379"
      ),
    masterName: z.string().max(100, "Master 名称不能超过 100 字"),
    username: z.string().max(100, "用户名不能超过 100 字"),
    password: z.string().max(100, "密码不能超过 100 字"),
    db: z.coerce
      .number()
      .int("DB 必须是整数")
      .min(0, "DB 取值范围为 0~15")
      .max(15, "DB 取值范围为 0~15"),
  })
  .superRefine((value, ctx) => {
    if (!value.clusterMode && !value.masterName.trim() && value.addresses.length > 1) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["addresses"],
        message: "单机模式只允许填写一个地址；多地址请开启集群模式或填写 Master 名称",
      })
    }
  })

function emptyForm(): RedisConfigForm {
  return {
    enabled: false,
    addresses: [""],
    username: "",
    password: "",
    clusterMode: false,
    masterName: "",
    db: 0,
  }
}

function toForm(config: RedisConfig | null): RedisConfigForm {
  if (!config) return emptyForm()
  return {
    enabled: config.enabled ?? false,
    addresses: config.addresses?.length ? config.addresses : [""],
    username: config.username ?? "",
    password: config.password ?? "",
    clusterMode: config.cluster_mode ?? false,
    masterName: config.master_name ?? "",
    db: config.db ?? 0,
  }
}

function modeLabel(config: RedisConfig | null): string {
  if (!config) return "单机"
  if (config.cluster_mode) {
    return config.master_name ? "集群 · 哨兵发现" : "集群"
  }
  return config.master_name ? "哨兵" : "单机"
}

function trimAddresses(list: string[]): string[] {
  return list.map((item) => item.trim()).filter(Boolean)
}

export function RedisConfigPage() {
  const [config, setConfig] = useState<RedisConfig | null>(null)
  const [form, setForm] = useState<RedisConfigForm>(emptyForm)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [showPassword, setShowPassword] = useState(false)
  const [activeTab, setActiveTab] = useState<TabKey>("info")
  const [showConfig, setShowConfig] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await redisConfigApi.get()
      setConfig(data)
      setForm(toForm(data))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载 Redis 配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (patch: Partial<RedisConfigForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const startEdit = () => {
    setForm(toForm(config))
    setErrors({})
    setShowPassword(false)
    setShowConfig(true)
    setEditing(true)
  }

  const cancelEdit = () => {
    setForm(toForm(config))
    setErrors({})
    setShowPassword(false)
    setShowConfig(false)
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

    const payload: RedisConfigPayload = {
      enabled: form.enabled,
      addresses: trimAddresses(parsed.data.addresses),
      username: form.username,
      password: form.password,
      cluster_mode: form.clusterMode,
      master_name: parsed.data.masterName.trim(),
      db: form.clusterMode ? 0 : parsed.data.db,
    }

    setSaving(true)
    try {
      await redisConfigApi.upsert(payload)
      toast.success("配置已保存")
      setShowPassword(false)
      setShowConfig(false)
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存 Redis 配置失败")
    } finally {
      setSaving(false)
    }
  }

  if (loading && !config) {
    return (
      <div className="h-full overflow-y-auto p-4">
        <div className="mx-auto flex w-full max-w-6xl animate-pulse flex-col gap-4">
          <div className="flex items-center gap-3 rounded-md border bg-card px-4 py-3">
            <div className="bg-muted size-10 shrink-0 rounded-md" />
            <div className="grid gap-2">
              <div className="bg-muted h-3.5 w-32 rounded" />
              <div className="bg-muted h-2.5 w-44 rounded" />
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
    <div className="flex h-full flex-col overflow-hidden p-4">
      <div className="mx-auto flex min-h-0 w-full max-w-6xl flex-1 flex-col gap-3">
        <div
          className={cn(
            "relative flex shrink-0 flex-wrap items-center justify-between gap-3 overflow-hidden rounded-md border bg-card px-4 py-3 transition-colors",
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
              <MemoryStick className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-semibold">Redis</h1>
                <ConfigStatusPill state={status} />
              </div>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <span>管理中心端缓存服务连接配置</span>
                <InfoTooltip label="Redis 配置说明">
                  <p>该配置供管理中心端连接 Redis 使用，保存后由后端读取。</p>
                  <p>
                    地址必须是
                    <code className="bg-muted mx-1 rounded px-1 py-0.5 font-mono text-[10px]">
                      host:port
                    </code>
                    形式，不要带协议前缀。
                  </p>
                  <p>开启集群模式可填写多个种子节点；填写 Master 名称即按哨兵模式连接。</p>
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
          <ConfigEmptyState
            icon={MemoryStick}
            title="尚未配置 Redis"
            description="填写连接地址与 DB 后即可启用缓存服务。"
            action={
              <Button size="sm" onClick={startEdit}>
                <Pencil /> 立即配置
              </Button>
            }
          />
        ) : (
          <div className="flex min-h-0 flex-1 flex-col gap-3">
            <section className="flex shrink-0 flex-col overflow-hidden rounded-md border bg-card">
              <div className="bg-muted/30 flex flex-wrap items-center gap-2 border-b px-4 py-2">
                <span className="bg-background text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded border">
                  <Server className="size-3.5" />
                </span>
                <h2 className="shrink-0 text-sm font-semibold">连接配置</h2>
                {!editing && config && (
                  <div className="text-muted-foreground flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 text-[11px]">
                    <span className="bg-background rounded-full border px-1.5 py-0.5">
                      {modeLabel(config)}
                    </span>
                    <span className="max-w-[420px] truncate font-mono">
                      {(config.addresses ?? []).join(", ")}
                    </span>
                    {config.master_name && (
                      <span className="truncate">master {config.master_name}</span>
                    )}
                    <span>DB {config.cluster_mode ? 0 : config.db}</span>
                    {config.username && <span>user {config.username}</span>}
                  </div>
                )}
                {!editing && (
                  <button
                    type="button"
                    aria-expanded={showConfig}
                    className="text-muted-foreground hover:text-foreground ml-auto flex shrink-0 items-center gap-1 text-[11px] transition-colors"
                    onClick={() => setShowConfig((prev) => !prev)}
                  >
                    {showConfig ? "收起" : "展开"}
                    <ChevronDown
                      className={cn(
                        "size-3.5 transition-transform",
                        showConfig && "rotate-180"
                      )}
                    />
                  </button>
                )}
              </div>

              {editing ? (
                <div className="grid max-h-[42vh] min-h-0 gap-3 overflow-y-auto p-3 sm:grid-cols-2">
                  <div className="grid gap-1.5 sm:col-span-2">
                    <div className="flex items-center justify-between gap-2">
                      <Label>地址</Label>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-6 px-1.5 text-[11px]"
                        onClick={() => update({ addresses: [...form.addresses, ""] })}
                      >
                        <Plus className="size-3" /> 添加地址
                      </Button>
                    </div>
                    <div className="grid gap-1.5">
                      {form.addresses.map((address, index) => (
                        <div key={index} className="flex items-center gap-1.5">
                          <Input
                            value={address}
                            placeholder={index === 0 ? "127.0.0.1:6379" : "10.0.0.2:6379"}
                            aria-label={`地址 ${index + 1}`}
                            onChange={(event) => {
                              const next = [...form.addresses]
                              next[index] = event.target.value
                              update({ addresses: next })
                            }}
                          />
                          <TooltipHint content="移除该地址">
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="text-muted-foreground size-8 shrink-0"
                              aria-label={`移除地址 ${index + 1}`}
                              disabled={form.addresses.length <= 1}
                              onClick={() =>
                                update({
                                  addresses: form.addresses.filter((_, i) => i !== index),
                                })
                              }
                            >
                              <Trash2 className="size-3.5" />
                            </Button>
                          </TooltipHint>
                        </div>
                      ))}
                    </div>
                    <p className="text-muted-foreground text-[11px]">
                      host:port 形式，不要带 redis:// 前缀；单机填 1 个，集群 / 哨兵可填多个
                    </p>
                    <ConfigFieldError message={errors.addresses} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="redis-master">Master 名称</Label>
                    <Input
                      id="redis-master"
                      value={form.masterName}
                      placeholder="哨兵模式填写，例如 mymaster"
                      onChange={(event) => update({ masterName: event.target.value })}
                    />
                    <p className="text-muted-foreground text-[11px]">
                      留空为直连，填写则通过哨兵发现节点
                    </p>
                    <ConfigFieldError message={errors.masterName} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="redis-username">用户名</Label>
                    <Input
                      id="redis-username"
                      value={form.username}
                      placeholder="Redis 6+ ACL 用户名，留空为 default"
                      autoComplete="off"
                      onChange={(event) => update({ username: event.target.value })}
                    />
                    <ConfigFieldError message={errors.username} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="redis-db">DB</Label>
                    <Input
                      id="redis-db"
                      type="number"
                      min={0}
                      max={15}
                      disabled={form.clusterMode}
                      value={form.clusterMode ? 0 : form.db}
                      onChange={(event) =>
                        update({ db: event.target.value === "" ? 0 : Number(event.target.value) })
                      }
                    />
                    <p className="text-muted-foreground text-[11px]">
                      {form.clusterMode ? "集群模式固定使用 DB 0" : "取值范围 0~15"}
                    </p>
                    <ConfigFieldError message={errors.db} />
                  </div>

                  <div className="grid gap-1.5">
                    <Label htmlFor="redis-password">密码</Label>
                    <div className="relative">
                      <Input
                        id="redis-password"
                        type={showPassword ? "text" : "password"}
                        value={form.password}
                        placeholder="留空表示无密码"
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
                    <ConfigFieldError message={errors.password} />
                  </div>

                  <div className="grid overflow-hidden rounded-md border sm:col-span-2 sm:grid-cols-2 sm:divide-x">
                    <div className="flex items-center justify-between gap-3 px-3 py-2">
                      <div className="min-w-0">
                        <div className="text-xs font-medium">集群模式</div>
                        <p className="text-muted-foreground truncate text-[11px]">
                          多地址种子节点，固定 DB 0
                        </p>
                      </div>
                      <Switch
                        checked={form.clusterMode}
                        label="集群模式"
                        onCheckedChange={(value) => update({ clusterMode: value })}
                      />
                    </div>
                    <div className="flex items-center justify-between gap-3 border-t px-3 py-2 sm:border-t-0">
                      <div className="min-w-0">
                        <div className="text-xs font-medium">启用 Redis</div>
                        <p className="text-muted-foreground truncate text-[11px]">
                          关闭后不读取该配置
                        </p>
                      </div>
                      <Switch
                        checked={form.enabled}
                        label="启用 Redis"
                        onCheckedChange={(value) => update({ enabled: value })}
                      />
                    </div>
                  </div>
                </div>
              ) : (
                showConfig && (
                  <div className="max-h-[42vh] min-h-0 divide-y overflow-y-auto px-4">
                    <ConfigReadonlyRow label="模式" value={modeLabel(config)} />
                    <ConfigReadonlyRow
                      label="地址"
                      value={(config?.addresses ?? []).join(", ")}
                      mono
                    />
                    {config?.master_name && (
                      <ConfigReadonlyRow label="Master" value={config.master_name} mono />
                    )}
                    <ConfigReadonlyRow label="用户名" value={config?.username} mono />
                    <ConfigReadonlyRow
                      label="DB"
                      value={config?.cluster_mode ? "0（集群）" : String(config?.db ?? 0)}
                      mono
                    />
                    <ConfigReadonlyRow
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
                )
              )}

              {config && (
                <div className="text-muted-foreground flex flex-wrap items-center gap-x-4 gap-y-1 border-t px-4 py-1.5 text-[11px]">
                  <span>
                    创建时间 {config.created_at ? formatDateTime(config.created_at) : "—"}
                  </span>
                  <span>
                    更新时间 {config.updated_at ? formatDateTime(config.updated_at) : "—"}
                  </span>
                </div>
              )}
            </section>

            {config?.enabled && (
              <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border bg-card">
                <div className="flex shrink-0 items-center gap-1 border-b px-2">
                  {TABS.map((tab) => {
                    const Icon = tab.icon
                    const selected = activeTab === tab.value
                    return (
                      <button
                        key={tab.value}
                        type="button"
                        aria-pressed={selected}
                        className={cn(
                          "-mb-px flex items-center gap-1.5 border-b-2 px-3 py-2 text-xs font-medium transition-colors",
                          selected
                            ? "border-primary text-foreground"
                            : "text-muted-foreground hover:text-foreground border-transparent"
                        )}
                        onClick={() => setActiveTab(tab.value)}
                      >
                        <Icon className="size-3.5" />
                        {tab.label}
                      </button>
                    )
                  })}
                </div>

                <div className="min-h-0 flex-1">
                  {activeTab === "info" && (
                    <div className="h-full min-h-0 overflow-hidden">
                      <RedisInfoPanel />
                    </div>
                  )}
                  {activeTab === "keys" && <RedisKeysBrowser />}
                  {activeTab === "cli" && (
                    <RedisCliPanel
                      endpoint={config.addresses?.[0]}
                      db={config.db}
                      cluster={config.cluster_mode}
                      master={config.master_name}
                    />
                  )}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}