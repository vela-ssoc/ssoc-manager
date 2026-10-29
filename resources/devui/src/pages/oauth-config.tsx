import { useCallback, useEffect, useState } from "react"
import {
  Eye,
  EyeOff,
  KeyRound,
  Link2,
  Pencil,
  Plus,
  Power,
  Route,
  Save,
  ShieldCheck,
  Tags,
  Trash2,
  X,
} from "lucide-react"
import { z } from "zod"

import { InfoTooltip } from "@/components/info-tooltip"
import { TooltipHint } from "@/components/tooltip-hint"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { oauthConfigApi } from "@/features/oauth-config/api"
import type {
  OAuthConfig,
  OAuthConfigForm,
} from "@/features/oauth-config/types"

const urlSchema = z
  .string()
  .trim()
  .min(1, "地址不能为空")
  .refine(
    (value) => /^https?:\/\/[^\s]+$/i.test(value),
    "请输入有效的 HTTP/HTTPS 地址"
  )

const schema = z.object({
  client_id: z.string().trim().min(1, "Client ID 不能为空").max(100, "不能超过 100 字"),
  client_secret: z
    .string()
    .trim()
    .min(1, "Client Secret 不能为空")
    .max(100, "不能超过 100 字"),
  authorization_endpoint: urlSchema,
  token_endpoint: urlSchema,
  userinfo_endpoint: urlSchema,
  redirect_uris: z
    .array(urlSchema)
    .min(1, "至少配置一个回调地址")
    .max(100, "回调地址不能超过 100 个"),
  scopes: z
    .array(
      z
        .string()
        .trim()
        .min(1, "授权范围不能为空")
        .max(10, "每个授权范围不能超过 10 字")
    )
    .max(10, "授权范围不能超过 10 个"),
})

const OAUTH_ENDPOINT_DEFAULTS = {
  authorization_endpoint:
    "https://dongdong-auth.eastmoney.com/dongdong-auth/login",
  token_endpoint:
    "https://dongdong-auth.eastmoney.com/dongdong-auth/oauth2/token",
  userinfo_endpoint:
    "https://dongdong-auth.eastmoney.com/dongdong-auth/userinfo",
}

function withEndpointDefaults(form: OAuthConfigForm): OAuthConfigForm {
  return {
    ...form,
    authorization_endpoint:
      form.authorization_endpoint || OAUTH_ENDPOINT_DEFAULTS.authorization_endpoint,
    token_endpoint:
      form.token_endpoint || OAUTH_ENDPOINT_DEFAULTS.token_endpoint,
    userinfo_endpoint:
      form.userinfo_endpoint || OAUTH_ENDPOINT_DEFAULTS.userinfo_endpoint,
  }
}

function withRedirectDefault(rows: string[]): string[] {
  return rows.some((item) => item.trim()) ? rows : [window.location.origin]
}

function emptyForm(): OAuthConfigForm {
  return {
    enabled: false,
    client_id: "",
    client_secret: "",
    authorization_endpoint: "",
    token_endpoint: "",
    userinfo_endpoint: "",
    redirect_uris: [""],
    scopes: [],
  }
}

function toForm(config: OAuthConfig | null): OAuthConfigForm {
  if (!config) return emptyForm()
  return {
    enabled: config.enabled ?? false,
    client_id: config.client_id ?? "",
    client_secret: config.client_secret ?? "",
    authorization_endpoint: config.authorization_endpoint ?? "",
    token_endpoint: config.token_endpoint ?? "",
    userinfo_endpoint: config.userinfo_endpoint ?? "",
    redirect_uris: config.redirect_uris?.length
      ? [...config.redirect_uris]
      : [""],
    scopes: config.scopes ?? [],
  }
}

export function OAuthConfigPage() {
  const [config, setConfig] = useState<OAuthConfig | null>(null)
  const [form, setForm] = useState<OAuthConfigForm>(emptyForm)
  const [redirectRows, setRedirectRows] = useState<string[]>([""])
  const [scopeRows, setScopeRows] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [showSecret, setShowSecret] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await oauthConfigApi.get()
      setConfig(data)
      const next = toForm(data)
      setForm(next)
      setRedirectRows(next.redirect_uris.length ? [...next.redirect_uris] : [""])
      setScopeRows([...next.scopes])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载身份验证配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (patch: Partial<OAuthConfigForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const startEdit = () => {
    const base = withEndpointDefaults(toForm(config))
    const next: OAuthConfigForm = {
      ...base,
      enabled: config ? base.enabled : true,
      redirect_uris: withRedirectDefault(base.redirect_uris),
    }
    setForm(next)
    setRedirectRows([...next.redirect_uris])
    setScopeRows([...next.scopes])
    setErrors({})
    setShowSecret(false)
    setEditing(true)
  }

  const cancelEdit = () => {
    const next = toForm(config)
    setForm(next)
    setRedirectRows(next.redirect_uris.length ? [...next.redirect_uris] : [""])
    setScopeRows([...next.scopes])
    setErrors({})
    setShowSecret(false)
    setEditing(false)
  }

  const save = async () => {
    const redirectUris = redirectRows.map((item) => item.trim()).filter(Boolean)
    const scopes = scopeRows.map((item) => item.trim()).filter(Boolean)
    const candidate: OAuthConfigForm = {
      ...form,
      redirect_uris: redirectUris,
      scopes,
    }
    const parsed = schema.safeParse(candidate)
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
      await oauthConfigApi.update({ ...candidate, ...parsed.data })
      toast.success("配置已保存")
      setShowSecret(false)
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存身份验证配置失败")
    } finally {
      setSaving(false)
    }
  }

  const addRedirect = () => setRedirectRows((rows) => [...rows, ""])
  const removeRedirect = (index: number) =>
    setRedirectRows((rows) => rows.filter((_, i) => i !== index))

  const addScope = () =>
    setScopeRows((rows) => (rows.length >= 10 ? rows : [...rows, ""]))
  const removeScope = (index: number) =>
    setScopeRows((rows) => rows.filter((_, i) => i !== index))

  if (loading && !config) {
    return (
      <div className="h-full overflow-y-auto p-4">
        <div className="mx-auto flex w-full max-w-4xl animate-pulse flex-col gap-4">
          <div className="flex items-center gap-3 rounded-md border bg-card px-4 py-3">
            <div className="bg-muted size-10 shrink-0 rounded-md" />
            <div className="grid gap-2">
              <div className="bg-muted h-3.5 w-28 rounded" />
              <div className="bg-muted h-2.5 w-40 rounded" />
            </div>
          </div>
          <div className="flex flex-col gap-4">
            <div className="h-32 rounded-md border bg-card" />
            <div className="h-48 rounded-md border bg-card" />
            <div className="h-28 rounded-md border bg-card" />
            <div className="h-20 rounded-md border bg-card" />
          </div>
        </div>
      </div>
    )
  }

  const enabled = editing ? form.enabled : Boolean(config?.enabled)
  const status = !config && !editing ? "empty" : enabled ? "enabled" : "disabled"
  const redirects = config?.redirect_uris ?? []
  const scopes = config?.scopes ?? []

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
              <ShieldCheck className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-semibold">身份验证</h1>
                <StatusPill state={status} />
              </div>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <span>当前仅支持咚咚统一认证登录</span>
                <InfoTooltip label="身份验证配置说明">
                  <p>用于配置登录页使用的咚咚统一认证服务。</p>
                  <p>回调地址需要与 OAuth 应用中登记的地址保持一致。</p>
                  <p>Client Secret 仅由服务端用于换取访问令牌。</p>
                  <p>OAuth 端点需保证管理中心端可以正常访问。</p>
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
              <ShieldCheck className="size-6" />
            </div>
            <div className="grid gap-1">
              <h2 className="text-sm font-semibold">尚未配置身份验证</h2>
              <p className="text-muted-foreground text-xs">
                配置 OAuth 应用后，登录页可使用授权服务登录。
              </p>
            </div>
            <div className="text-muted-foreground flex flex-wrap justify-center gap-1.5 text-[11px]">
              {["填写应用信息", "配置回调地址", "启用并保存"].map((step, index) => (
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
              <Section icon={KeyRound} title="基础配置">
                {editing ? (
                  <div className="grid gap-4 p-4 sm:grid-cols-2">
                    <TextField
                      id="oauth-client-id"
                      label="Client ID"
                      value={form.client_id}
                      placeholder="OAuth 应用客户端 ID"
                      error={errors.client_id}
                      onChange={(value) => update({ client_id: value })}
                    />
                    <div className="grid gap-1.5">
                      <Label htmlFor="oauth-client-secret">Client Secret</Label>
                      <div className="relative">
                        <Input
                          id="oauth-client-secret"
                          type={showSecret ? "text" : "password"}
                          value={form.client_secret}
                          placeholder="OAuth 应用客户端密钥"
                          autoComplete="new-password"
                          className="pr-9"
                          onChange={(event) =>
                            update({ client_secret: event.target.value })
                          }
                        />
                        <button
                          type="button"
                          aria-label={showSecret ? "隐藏密钥" : "显示密钥"}
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
                      <FieldError message={errors.client_secret} />
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
                          <div className="text-sm font-medium">启用 OAuth 登录</div>
                          <p className="text-muted-foreground mt-0.5 text-xs">
                            关闭后登录页将隐藏 OAuth 登录入口，依赖 OAuth 登录的用户将无法登录
                          </p>
                        </div>
                      </div>
                      <Switch
                        checked={form.enabled}
                        label="启用 OAuth 登录"
                        onCheckedChange={(value) => update({ enabled: value })}
                      />
                    </div>
                  </div>
                ) : (
                  <div className="divide-y px-4">
                    <ReadonlyRow label="Client ID" value={config?.client_id} mono />
                    <ReadonlyRow
                      label="Client Secret"
                      value={
                        config?.client_secret
                          ? showSecret
                            ? config.client_secret
                            : "••••••••"
                          : undefined
                      }
                      mono
                      action={
                        config?.client_secret ? (
                          <TooltipHint content={showSecret ? "隐藏密钥" : "显示密钥"}>
                            <button
                              type="button"
                              aria-label={showSecret ? "隐藏密钥" : "显示密钥"}
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

              <Section
                icon={Link2}
                title="OAuth 端点"
                description="授权服务提供的标准 OAuth 地址"
              >
                {editing ? (
                  <div className="grid gap-4 p-4">
                    <TextField
                      id="oauth-authorization-endpoint"
                      label="授权端点"
                      value={form.authorization_endpoint}
                      placeholder="https://auth.example.com/oauth/authorize"
                      error={errors.authorization_endpoint}
                      onChange={(value) =>
                        update({ authorization_endpoint: value })
                      }
                    />
                    <TextField
                      id="oauth-token-endpoint"
                      label="令牌端点"
                      value={form.token_endpoint}
                      placeholder="https://auth.example.com/oauth/token"
                      error={errors.token_endpoint}
                      onChange={(value) => update({ token_endpoint: value })}
                    />
                    <TextField
                      id="oauth-userinfo-endpoint"
                      label="用户信息端点"
                      value={form.userinfo_endpoint}
                      placeholder="https://auth.example.com/oauth/userinfo"
                      error={errors.userinfo_endpoint}
                      onChange={(value) => update({ userinfo_endpoint: value })}
                    />
                  </div>
                ) : (
                  <div className="divide-y px-4">
                    <ReadonlyRow
                      label="授权端点"
                      value={config?.authorization_endpoint}
                      mono
                    />
                    <ReadonlyRow
                      label="令牌端点"
                      value={config?.token_endpoint}
                      mono
                    />
                    <ReadonlyRow
                      label="用户信息端点"
                      value={config?.userinfo_endpoint}
                      mono
                    />
                  </div>
                )}
              </Section>

              <Section
                icon={Route}
                title="回调地址"
                description="至少一个，需与 OAuth 应用登记的地址一致"
                actions={
                  editing ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={addRedirect}
                    >
                      <Plus /> 添加
                    </Button>
                  ) : undefined
                }
              >
                {editing ? (
                  <div className="grid gap-2 p-4">
                    {redirectRows.map((row, index) => (
                      <div key={index} className="grid grid-cols-[1fr_auto] gap-2">
                        <Input
                          aria-label={`回调地址 ${index + 1}`}
                          value={row}
                          placeholder="https://example.com/oauth/callback"
                          onChange={(event) => {
                            const next = [...redirectRows]
                            next[index] = event.target.value
                            setRedirectRows(next)
                          }}
                        />
                        <TooltipHint content="移除回调地址">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="text-muted-foreground hover:text-destructive size-8"
                            aria-label="移除回调地址"
                            onClick={() => removeRedirect(index)}
                          >
                            <Trash2 />
                          </Button>
                        </TooltipHint>
                      </div>
                    ))}
                    {redirectRows.length === 0 && (
                      <p className="text-muted-foreground text-xs">
                        暂无回调地址，请至少添加一个
                      </p>
                    )}
                    <FieldError message={errors.redirect_uris} />
                  </div>
                ) : (
                  <div className="p-4">
                    {redirects.length === 0 ? (
                      <p className="text-muted-foreground text-xs">暂无回调地址</p>
                    ) : (
                      <div className="divide-y overflow-hidden rounded-md border">
                        {redirects.map((uri) => (
                          <div
                            key={uri}
                            className="hover:bg-muted/40 px-3 py-2 transition-colors"
                          >
                            <TooltipHint content={uri}>
                              <div className="truncate font-mono text-xs">{uri}</div>
                            </TooltipHint>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </Section>

              <Section
                icon={Tags}
                title="授权范围"
                description="可选，最多 10 个，每个不超过 10 个字符"
                actions={
                  editing ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={scopeRows.length >= 10}
                      onClick={addScope}
                    >
                      <Plus /> 添加
                    </Button>
                  ) : undefined
                }
              >
                {editing ? (
                  <div className="grid gap-2 p-4">
                    {scopeRows.map((row, index) => (
                      <div key={index} className="grid grid-cols-[1fr_auto] gap-2">
                        <Input
                          aria-label={`授权范围 ${index + 1}`}
                          value={row}
                          placeholder="例如 openid"
                          onChange={(event) => {
                            const next = [...scopeRows]
                            next[index] = event.target.value
                            setScopeRows(next)
                          }}
                        />
                        <TooltipHint content="移除授权范围">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="text-muted-foreground hover:text-destructive size-8"
                            aria-label="移除授权范围"
                            onClick={() => removeScope(index)}
                          >
                            <Trash2 />
                          </Button>
                        </TooltipHint>
                      </div>
                    ))}
                    {scopeRows.length === 0 && (
                      <p className="text-muted-foreground text-xs">
                        暂无授权范围，可按需添加
                      </p>
                    )}
                    <FieldError message={errors.scopes} />
                  </div>
                ) : (
                  <div className="p-4">
                    {scopes.length === 0 ? (
                      <p className="text-muted-foreground text-xs">暂无授权范围</p>
                    ) : (
                      <div className="flex flex-wrap gap-1.5">
                        {scopes.map((scope) => (
                          <Badge
                            key={scope}
                            variant="secondary"
                            className="border-primary/20 bg-primary/10 text-primary font-mono text-[11px]"
                          >
                            {scope}
                          </Badge>
                        ))}
                      </div>
                    )}
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
  icon: typeof KeyRound
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
              <p className="text-muted-foreground truncate text-[11px]">
                {description}
              </p>
            )}
          </div>
        </div>
        {actions}
      </div>
      {children}
    </section>
  )
}

function TextField({
  id,
  label,
  value,
  placeholder,
  error,
  onChange,
}: {
  id: string
  label: string
  value: string
  placeholder?: string
  error?: string
  onChange: (value: string) => void
}) {
  return (
    <div className="grid gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
      <FieldError message={error} />
    </div>
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