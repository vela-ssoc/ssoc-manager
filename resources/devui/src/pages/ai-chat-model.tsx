import { useCallback, useEffect, useState } from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import {
  Bot,
  Eye,
  EyeOff,
  ListChecks,
  Loader2,
  MessageSquareText,
  Pencil,
  Plus,
  Power,
  RefreshCw,
  Save,
  Server,
  X,
} from "lucide-react"
import { z } from "zod"

import {
  ConfigEmptyState,
  ConfigFieldError,
  ConfigReadonlyRow,
  ConfigSection,
  ConfigStatusPill,
} from "@/components/config-panel"
import { InfoTooltip } from "@/components/info-tooltip"
import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { aiChatModelApi } from "@/features/ai-chat-model/api"
import type {
  AIChatModelConfig,
  AIChatModelForm,
} from "@/features/ai-chat-model/types"

const schema = z.object({
  base_url: z
    .string()
    .trim()
    .min(1, "服务地址不能为空")
    .refine(
      (value) => /^https?:\/\/[^\s]+$/i.test(value),
      "请输入有效的 HTTP/HTTPS 地址"
    ),
  api_key: z.string().trim().min(1, "API Key 不能为空"),
  prompt: z.string().max(10000, "系统提示词不能超过 10000 字"),
})

function emptyForm(): AIChatModelForm {
  return {
    enabled: false,
    base_url: "",
    api_key: "",
    models: [],
    prompt: "",
  }
}

function toForm(config: AIChatModelConfig | null): AIChatModelForm {
  if (!config) return emptyForm()
  return {
    enabled: config.enabled ?? false,
    base_url: config.base_url ?? "",
    api_key: config.api_key ?? "",
    models: [...(config.models ?? [])],
    prompt: config.prompt ?? "",
  }
}

function collectErrors(error: z.ZodError) {
  const next: Record<string, string> = {}
  for (const issue of error.issues) {
    const key = String(issue.path[0] ?? "")
    if (!next[key]) next[key] = issue.message
  }
  return next
}

export function AIChatModelPage() {
  const [config, setConfig] = useState<AIChatModelConfig | null>(null)
  const [form, setForm] = useState<AIChatModelForm>(emptyForm)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [showKey, setShowKey] = useState(false)
  const [modelInput, setModelInput] = useState("")
  const [upstream, setUpstream] = useState<string[]>([])
  const [fetchingModels, setFetchingModels] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await aiChatModelApi.get()
      setConfig(data)
      setForm(toForm(data))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载聊天模型配置失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (patch: Partial<AIChatModelForm>) =>
    setForm((prev) => ({ ...prev, ...patch }))

  const clearError = (key: string) =>
    setErrors((prev) => {
      if (!prev[key]) return prev
      const next = { ...prev }
      delete next[key]
      return next
    })

  const startEdit = () => {
    setForm({ ...toForm(config), enabled: config ? config.enabled : true })
    setErrors({})
    setShowKey(false)
    setModelInput("")
    setUpstream([])
    setEditing(true)
  }

  const cancelEdit = () => {
    setForm(toForm(config))
    setErrors({})
    setShowKey(false)
    setModelInput("")
    setUpstream([])
    setEditing(false)
  }

  const addModel = (value: string) => {
    const model = value.trim()
    if (!model) return
    setForm((prev) =>
      prev.models.includes(model)
        ? prev
        : { ...prev, models: [...prev.models, model] }
    )
    setModelInput("")
    clearError("models")
  }

  const removeModel = (value: string) => {
    setForm((prev) => ({
      ...prev,
      models: prev.models.filter((model) => model !== value),
    }))
  }

  const fetchUpstream = async () => {
    const parsed = schema
      .pick({ base_url: true, api_key: true })
      .safeParse({ base_url: form.base_url, api_key: form.api_key })
    if (!parsed.success) {
      setErrors(collectErrors(parsed.error))
      return
    }

    setFetchingModels(true)
    try {
      const data = await aiChatModelApi.listModels(
        parsed.data.base_url,
        parsed.data.api_key
      )
      const ids = Array.from(
        new Set((data.data ?? []).map((item) => item.id).filter(Boolean))
      ).sort((a, b) => a.localeCompare(b))
      setUpstream(ids)
      if (!ids.length) {
        toast.error("上游没有返回可用的模型")
      }
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "获取上游模型列表失败"
      )
    } finally {
      setFetchingModels(false)
    }
  }

  const save = async () => {
    const models = Array.from(
      new Set(form.models.map((model) => model.trim()).filter(Boolean))
    )
    if (!models.length) {
      setErrors({ models: "至少添加一个模型" })
      return
    }
    if (models.length > 1000) {
      setErrors({ models: "模型数量不能超过 1000 个" })
      return
    }

    const parsed = schema.safeParse(form)
    if (!parsed.success) {
      setErrors(collectErrors(parsed.error))
      return
    }

    setSaving(true)
    try {
      await aiChatModelApi.update({
        enabled: form.enabled,
        base_url: parsed.data.base_url,
        api_key: parsed.data.api_key,
        models,
        prompt: form.prompt,
      })
      toast.success("聊天模型配置已保存")
      setShowKey(false)
      setUpstream([])
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存聊天模型配置失败")
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
              <div className="bg-muted h-3.5 w-28 rounded" />
              <div className="bg-muted h-2.5 w-48 rounded" />
            </div>
          </div>
          <div className="h-44 rounded-md border bg-card" />
          <div className="h-32 rounded-md border bg-card" />
        </div>
      </div>
    )
  }

  const enabled = editing ? form.enabled : Boolean(config?.enabled)
  const status = !config && !editing ? "empty" : enabled ? "enabled" : "disabled"
  const key = config?.api_key ?? ""

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
              <Bot className="size-5" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-base font-semibold">聊天模型</h1>
                <ConfigStatusPill state={status} />
              </div>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-1 text-xs">
                <span>AI 助手使用的对话模型与默认提示词</span>
                <InfoTooltip label="聊天模型配置说明">
                  <p>兼容 OpenAI 的 chat completions 接口，服务地址填写到版本前缀，例如 /v1。</p>
                  <p>模型列表决定 AI 助手可使用的模型，至少需要配置一个。</p>
                  <p>系统提示词会作为默认 Prompt 下发给模型，留空时使用系统内置提示词。</p>
                  <p>关闭后聊天模型配置保留，但 AI 助手不可使用。</p>
                </InfoTooltip>
              </div>
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-2">
            {editing ? (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={saving}
                  onClick={cancelEdit}
                >
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
            icon={Bot}
            title="尚未配置聊天模型"
            description="配置 OpenAI 兼容接口后，AI 助手才能使用模型进行对话。"
            action={
              <Button size="sm" onClick={startEdit}>
                <Pencil /> 立即配置
              </Button>
            }
          />
        ) : (
          <div className="flex flex-col gap-4">
            <ConfigSection
              icon={Server}
              title="连接配置"
              description="OpenAI 兼容接口地址与访问凭证"
            >
              {editing ? (
                <div className="grid gap-4 p-4 sm:grid-cols-2">
                  <div className="grid gap-1.5 sm:col-span-2">
                    <Label htmlFor="ai-base-url">服务地址</Label>
                    <Input
                      id="ai-base-url"
                      value={form.base_url}
                      placeholder="https://api.openai.com/v1"
                      onChange={(event) => {
                        update({ base_url: event.target.value })
                        clearError("base_url")
                      }}
                    />
                    <p className="text-muted-foreground text-[11px]">
                      填写到版本前缀，不包含 /chat/completions
                    </p>
                    <ConfigFieldError message={errors.base_url} />
                  </div>

                  <div className="grid gap-1.5 sm:col-span-2">
                    <Label htmlFor="ai-api-key">API Key</Label>
                    <div className="relative">
                      <Input
                        id="ai-api-key"
                        type={showKey ? "text" : "password"}
                        value={form.api_key}
                        placeholder="sk-…"
                        autoComplete="new-password"
                        className="pr-9"
                        onChange={(event) => {
                          update({ api_key: event.target.value })
                          clearError("api_key")
                        }}
                      />
                      <button
                        type="button"
                        aria-label={showKey ? "隐藏 API Key" : "显示 API Key"}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 flex size-6 -translate-y-1/2 items-center justify-center rounded"
                        onClick={() => setShowKey((visible) => !visible)}
                      >
                        {showKey ? (
                          <EyeOff className="size-3.5" />
                        ) : (
                          <Eye className="size-3.5" />
                        )}
                      </button>
                    </div>
                    <ConfigFieldError message={errors.api_key} />
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
                        <div className="text-sm font-medium">启用聊天模型</div>
                        <p className="text-muted-foreground mt-0.5 text-xs">
                          关闭后保留配置，但 AI 助手不可使用
                        </p>
                      </div>
                    </div>
                    <Switch
                      checked={form.enabled}
                      label="启用聊天模型"
                      onCheckedChange={(value) => update({ enabled: value })}
                    />
                  </div>
                </div>
              ) : (
                <div className="divide-y px-4">
                  <ConfigReadonlyRow
                    label="服务地址"
                    value={config?.base_url}
                    mono
                  />
                  <ConfigReadonlyRow
                    label="API Key"
                    value={key ? (showKey ? key : "••••••••") : undefined}
                    mono
                    action={
                      key ? (
                        <TooltipHint
                          content={showKey ? "隐藏 API Key" : "显示 API Key"}
                        >
                          <button
                            type="button"
                            aria-label={showKey ? "隐藏 API Key" : "显示 API Key"}
                            className="text-muted-foreground hover:text-foreground flex size-5 shrink-0 items-center justify-center rounded"
                            onClick={() => setShowKey((visible) => !visible)}
                          >
                            {showKey ? (
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
            </ConfigSection>

            <ConfigSection
              icon={ListChecks}
              title="可用模型"
              description={
                editing ? "至少添加一个模型" : "AI 助手可使用的模型"
              }
              actions={
                editing ? (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={fetchingModels}
                    onClick={() => void fetchUpstream()}
                  >
                    <RefreshCw
                      className={cn(fetchingModels && "animate-spin")}
                    />
                    {fetchingModels ? "获取中…" : "获取上游模型"}
                  </Button>
                ) : undefined
              }
            >
              {editing ? (
                <div className="grid gap-3 p-4">
                  <div className="flex gap-2">
                    <Input
                      value={modelInput}
                      placeholder="模型 ID，例如 deepseek-chat"
                      onChange={(event) => setModelInput(event.target.value)}
                      onKeyDown={(event) => {
                        if (
                          event.key === "Enter" &&
                          !event.nativeEvent.isComposing
                        ) {
                          event.preventDefault()
                          addModel(modelInput)
                        }
                      }}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => addModel(modelInput)}
                    >
                      <Plus /> 添加
                    </Button>
                  </div>
                  <ConfigFieldError message={errors.models} />

                  <div className="flex max-h-32 flex-wrap gap-1.5 overflow-y-auto">
                    {form.models.map((model) => (
                      <span
                        key={model}
                        className="bg-muted/60 inline-flex h-6 items-center gap-1 rounded-md border pr-1 pl-2 font-mono text-[11px]"
                      >
                        {model}
                        <button
                          type="button"
                          aria-label={`移除模型 ${model}`}
                          className="text-muted-foreground hover:text-destructive flex size-4 items-center justify-center rounded"
                          onClick={() => removeModel(model)}
                        >
                          <X className="size-3" />
                        </button>
                      </span>
                    ))}
                    {form.models.length === 0 && (
                      <span className="text-muted-foreground text-xs">
                        暂无模型
                      </span>
                    )}
                  </div>

                  {upstream.length > 0 && (
                    <div className="grid gap-2 rounded-md border bg-muted/20 p-3">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-xs font-medium">
                          上游可用模型（{upstream.length}）
                        </span>
                        <button
                          type="button"
                          className="text-muted-foreground hover:text-foreground text-[11px]"
                          onClick={() => setUpstream([])}
                        >
                          关闭
                        </button>
                      </div>
                      <div className="flex max-h-40 flex-wrap gap-1.5 overflow-y-auto">
                        {upstream.map((model) => {
                          const added = form.models.includes(model)
                          return (
                            <button
                              key={model}
                              type="button"
                              disabled={added}
                              className={cn(
                                "rounded-md border px-2 py-1 font-mono text-[11px] transition-colors",
                                added
                                  ? "bg-muted/40 text-muted-foreground cursor-default opacity-60"
                                  : "hover:border-primary/40 hover:bg-primary/5"
                              )}
                              onClick={() => addModel(model)}
                            >
                              {model}
                              {added && (
                                <span className="ml-1 font-sans">已添加</span>
                              )}
                            </button>
                          )
                        })}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="p-4">
                  {config?.models?.length ? (
                    <div className="flex max-h-60 flex-wrap gap-1.5 overflow-y-auto">
                      {config.models.map((model) => (
                        <span
                          key={model}
                          className="bg-muted/60 inline-flex h-6 items-center rounded-md border px-2 font-mono text-[11px]"
                        >
                          {model}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <div className="flex flex-col items-center gap-2 rounded-md border border-dashed px-4 py-6 text-center">
                      <ListChecks className="text-muted-foreground size-5 opacity-60" />
                      <span className="text-muted-foreground text-xs">
                        暂无可用模型
                      </span>
                    </div>
                  )}
                </div>
              )}
            </ConfigSection>

            <ConfigSection
              icon={MessageSquareText}
              title="系统提示词"
              description="可选，作为默认 Prompt 下发给模型"
            >
              {editing ? (
                <div className="grid gap-1.5 p-4">
                  <Textarea
                    value={form.prompt}
                    maxLength={10000}
                    rows={6}
                    placeholder="留空则使用系统内置提示词"
                    className="resize-y"
                    onChange={(event) => {
                      update({ prompt: event.target.value })
                      clearError("prompt")
                    }}
                  />
                  <div className="text-muted-foreground flex items-center justify-between gap-2 text-[11px]">
                    <span>留空则使用系统内置提示词</span>
                    <span className="tabular-nums">
                      {form.prompt.length}/10000
                    </span>
                  </div>
                  <ConfigFieldError message={errors.prompt} />
                </div>
              ) : config?.prompt ? (
                <div className="markdown-body max-h-96 overflow-y-auto p-4">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {config.prompt}
                  </ReactMarkdown>
                </div>
              ) : (
                <div className="p-4">
                  <div className="flex flex-col items-center gap-2 rounded-md border border-dashed px-4 py-6 text-center">
                    <MessageSquareText className="text-muted-foreground size-5 opacity-60" />
                    <span className="text-muted-foreground text-xs">
                      未配置，使用系统内置提示词
                    </span>
                  </div>
                </div>
              )}
            </ConfigSection>

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

        {loading && config && (
          <div className="text-muted-foreground flex items-center justify-center gap-2 text-xs">
            <Loader2 className="size-3.5 animate-spin" />
            刷新中…
          </div>
        )}
      </div>
    </div>
  )
}