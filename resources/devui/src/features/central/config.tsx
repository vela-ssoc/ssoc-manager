import { useCallback, useEffect, useState } from "react"
import { Check, Loader2, Pencil, Plus, Save, Trash2, TriangleAlert, X } from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { TooltipHint } from "@/components/tooltip-hint"
import { InfoTooltip } from "@/components/info-tooltip"
import { cn } from "@/lib/utils"
import { centralApi } from "./api"
import type {
  CentralConfig,
  CentralConfigForm,
  CentralConfigResponse,
  CentralLoggerConfig,
  CentralStaticConfig,
} from "./types"

const LOGGER_LEVELS = ["DEBUG", "INFO", "WARN", "ERROR"]

function emptyLogger(): CentralLoggerConfig {
  return {
    level: "INFO",
    console: false,
    filename: "",
    maxsize: 100,
    maxage: 180,
    maxbackups: 10000,
    localtime: true,
    compress: true,
  }
}

function toForm(config: CentralConfig | null): CentralConfigForm {
  return {
    server: {
      addr: config?.server?.addr ?? "",
      static: { ...(config?.server?.static ?? {}) },
    },
    logger: config?.logger ?? emptyLogger(),
  }
}

interface DiffRow {
  label: string
  saved: string
  effect: string
}

interface DiffSection {
  title: string
  rows: DiffRow[]
}

function onOff(value: boolean): string {
  return value ? "开" : "关"
}

function staticText(item?: CentralStaticConfig): string {
  if (!item) return "—"
  return `${item.dir || "—"}（${item.spa ? "SPA" : "非 SPA"}）`
}

function buildDiff(form: CentralConfigForm, effect: CentralConfigForm): DiffSection[] {
  const serverRows: DiffRow[] = [
    { label: "监听地址", saved: form.server.addr, effect: effect.server.addr },
  ]

  const paths = Array.from(
    new Set([...Object.keys(form.server.static), ...Object.keys(effect.server.static)])
  ).sort()
  for (const path of paths) {
    serverRows.push({
      label: `静态映射 ${path}`,
      saved: staticText(form.server.static[path]),
      effect: staticText(effect.server.static[path]),
    })
  }

  const loggerRows: DiffRow[] = [
    { label: "级别", saved: form.logger.level, effect: effect.logger.level },
    { label: "控制台输出", saved: onOff(form.logger.console), effect: onOff(effect.logger.console) },
    { label: "日志输出文件", saved: form.logger.filename, effect: effect.logger.filename },
    { label: "日志文件分割大小 MB", saved: String(form.logger.maxsize), effect: String(effect.logger.maxsize) },
    { label: "保留天数", saved: String(form.logger.maxage), effect: String(effect.logger.maxage) },
    { label: "文件数上限", saved: String(form.logger.maxbackups), effect: String(effect.logger.maxbackups) },
    { label: "使用本地时间", saved: onOff(form.logger.localtime), effect: onOff(effect.logger.localtime) },
    { label: "压缩历史日志", saved: onOff(form.logger.compress), effect: onOff(effect.logger.compress) },
  ]

  return [
    { title: "服务配置", rows: serverRows },
    { title: "日志配置", rows: loggerRows },
  ]
}

function hasDiff(sections: DiffSection[]): boolean {
  return sections.some((section) =>
    section.rows.some((row) => row.saved !== row.effect)
  )
}

export function CentralConfigSection() {
  const [effect, setEffect] = useState<CentralConfigResponse | null>(null)
  const [latest, setLatest] = useState<CentralConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<CentralConfigForm | null>(null)
  const [saving, setSaving] = useState(false)
  const [editing, setEditing] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [eff, lat] = await Promise.all([
        centralApi.effectConfig(),
        centralApi.latestConfig().catch(() => null),
      ])
      setEffect(eff)
      setLatest(lat)
      setForm(toForm(lat))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载配置信息失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const updateStatic = (path: string, patch: Partial<CentralStaticConfig>) => {
    if (!form) return
    setForm({
      ...form,
      server: {
        ...form.server,
        static: {
          ...form.server.static,
          [path]: { ...form.server.static[path], ...patch },
        },
      },
    })
  }

  const renameStatic = (oldPath: string, nextPath: string) => {
    if (!form || oldPath === nextPath || !nextPath) return
    const next = { ...form.server.static }
    const item = next[oldPath]
    if (!item) return
    delete next[oldPath]
    next[nextPath] = item
    setForm({ ...form, server: { ...form.server, static: next } })
  }

  const removeStatic = (path: string) => {
    if (!form) return
    const next = { ...form.server.static }
    delete next[path]
    setForm({ ...form, server: { ...form.server, static: next } })
  }

  const addStatic = () => {
    if (!form) return
    let key = "/"
    let i = 1
    while (key in form.server.static) key = `/path-${i++}`
    setForm({
      ...form,
      server: {
        ...form.server,
        static: { ...form.server.static, [key]: { spa: false, dir: "" } },
      },
    })
  }

  const updateLogger = (patch: Partial<CentralLoggerConfig>) => {
    if (!form) return
    setForm({ ...form, logger: { ...form.logger, ...patch } })
  }

  const cancelEdit = () => {
    setForm(toForm(latest))
    setEditing(false)
  }

  const save = async () => {
    if (!form) return
    const staticsValid = Object.entries(form.server.static).every(
      ([path, item]) => path.trim() && item.dir.trim()
    )
    if (!staticsValid) {
      toast.error("静态资源映射的请求路径和目录不能为空")
      return
    }
    setSaving(true)
    try {
      await centralApi.updateConfig(form)
      toast.success("配置已保存，重启后生效")
      setEditing(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存配置失败")
    } finally {
      setSaving(false)
    }
  }

  if (loading && !form) {
    return (
      <div className="text-muted-foreground flex items-center gap-2 text-xs">
        <Loader2 className="size-3.5 animate-spin" />
        正在加载配置…
      </div>
    )
  }

  if (!effect || !form) {
    return (
      <div className="text-muted-foreground text-xs">无法获取配置信息</div>
    )
  }

  const boot = effect.boot
  const effectForm = toForm(effect.config)
  const latestForm = toForm(latest)
  const readonlyStatics = Object.entries(form.server.static ?? {})

  const diffSections = buildDiff(latestForm, effectForm)
  const changed = hasDiff(diffSections)

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">配置信息</h3>
        {editing ? (
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={saving} onClick={cancelEdit}>
              <X /> 取消
            </Button>
            <Button size="sm" disabled={saving} onClick={() => void save()}>
              <Save /> {saving ? "保存中…" : "保存"}
            </Button>
          </div>
        ) : (
          <TooltipHint content="编辑配置">
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              aria-label="编辑配置"
              onClick={() => setEditing(true)}
            >
              <Pencil className="size-3.5" />
            </Button>
          </TooltipHint>
        )}
      </div>

      <section className="grid gap-3">
        <div className="flex items-center gap-1.5">
          <h3 className="text-sm font-semibold">数据库配置</h3>
          <InfoTooltip label="数据库配置说明">
            数据库连接为启动参数，不可在线修改。
          </InfoTooltip>
        </div>
        <div className="grid gap-3 sm:grid-cols-2">
          <ReadonlyField label="MongoDB 地址" value={boot.database?.uri} mono />
          <ReadonlyField label="关系数据库地址" value={boot.database?.dsn} mono />
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">服务配置</h3>
        {editing ? (
          <>
            <div className="grid gap-1.5">
              <Label htmlFor="central-server-addr">监听地址</Label>
              <Input
                id="central-server-addr"
                value={form.server.addr}
                placeholder="0.0.0.0:8081"
                onChange={(event) =>
                  setForm({
                    ...form,
                    server: { ...form.server, addr: event.target.value },
                  })
                }
              />
            </div>

            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-[11px] font-medium">
                静态资源映射
              </span>
              <Button type="button" variant="outline" size="sm" onClick={addStatic}>
                <Plus /> 添加
              </Button>
            </div>
            <div className="grid gap-2">
              {Object.entries(form.server.static).map(([path, item]) => (
                <div
                  key={path}
                  className="grid grid-cols-[1fr_auto_1.5fr_auto] items-center gap-2"
                >
                  <StaticInput
                    value={path}
                    placeholder="请求路径，如 /devui"
                    onCommit={(value) => renameStatic(path, value)}
                  />
                  <Select
                    value={item.spa ? "on" : "off"}
                    onValueChange={(value) => updateStatic(path, { spa: value === "on" })}
                  >
                    <SelectTrigger className="w-28">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="off">非 SPA</SelectItem>
                      <SelectItem value="on">SPA</SelectItem>
                    </SelectContent>
                  </Select>
                  <Input
                    value={item.dir}
                    placeholder="目录，如 resources/static/devui/"
                    onChange={(event) => updateStatic(path, { dir: event.target.value })}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="text-muted-foreground hover:text-destructive size-8"
                    aria-label="移除映射"
                    onClick={() => removeStatic(path)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))}
              {Object.keys(form.server.static).length === 0 && (
                <p className="text-muted-foreground text-xs">暂无静态资源映射</p>
              )}
            </div>
          </>
        ) : (
          <>
            <div className="grid gap-3 sm:grid-cols-2">
              <ReadonlyField label="监听地址" value={form.server.addr} mono />
            </div>
            {readonlyStatics.length > 0 && (
              <div className="overflow-hidden rounded-md border">
                <div className="bg-muted/40 text-muted-foreground grid grid-cols-[3fr_2fr_5fr] border-b px-2 py-1 text-[11px]">
                  <span>请求路径</span>
                  <span className="px-2">单页面应用（SPA）</span>
                  <span>目录</span>
                </div>
                {readonlyStatics.map(([path, item]) => (
                  <div
                    key={path}
                    className="grid grid-cols-[3fr_2fr_5fr] items-center border-b px-2 py-1.5 text-xs last:border-b-0"
                  >
                    <TooltipHint content={path}>
                      <span className="truncate font-mono">{path}</span>
                    </TooltipHint>
                    <span className="text-muted-foreground px-2">
                      {item.spa ? "是" : "否"}
                    </span>
                    <TooltipHint content={item.dir}>
                      <span className="truncate font-mono">{item.dir || "—"}</span>
                    </TooltipHint>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">日志配置</h3>
        {editing ? (
          <>
            <div className="grid grid-cols-2 gap-3">
              <div className="grid gap-1.5">
                <Label>级别</Label>
                <Select
                  value={form.logger.level || "INFO"}
                  onValueChange={(value) => updateLogger({ level: value })}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {LOGGER_LEVELS.map((level) => (
                      <SelectItem key={level} value={level}>
                        {level}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-1.5">
                <Label>控制台输出</Label>
                <Select
                  value={form.logger.console ? "on" : "off"}
                  onValueChange={(value) => updateLogger({ console: value === "on" })}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="off">关</SelectItem>
                    <SelectItem value="on">开</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid gap-1.5">
              <Label htmlFor="central-log-filename">日志输出文件</Label>
              <Input
                id="central-log-filename"
                value={form.logger.filename}
                placeholder="留空表示不输出日志文件"
                onChange={(event) => updateLogger({ filename: event.target.value })}
              />
            </div>

            <div className="grid gap-3 rounded-md border p-3">
              <span className="text-xs font-medium">日志轮转</span>
              <div className="grid grid-cols-3 gap-3">
                <div className="grid gap-1.5">
                  <Label htmlFor="central-log-maxsize">日志文件分割大小 MB</Label>
                  <Input
                    id="central-log-maxsize"
                    type="number"
                    min={0}
                    value={String(form.logger.maxsize)}
                    onChange={(event) =>
                      updateLogger({ maxsize: Number(event.target.value) || 0 })
                    }
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="central-log-maxage">保留天数</Label>
                  <Input
                    id="central-log-maxage"
                    type="number"
                    min={0}
                    value={String(form.logger.maxage)}
                    onChange={(event) =>
                      updateLogger({ maxage: Number(event.target.value) || 0 })
                    }
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="central-log-maxbackups">文件数上限</Label>
                  <Input
                    id="central-log-maxbackups"
                    type="number"
                    min={0}
                    value={String(form.logger.maxbackups)}
                    onChange={(event) =>
                      updateLogger({ maxbackups: Number(event.target.value) || 0 })
                    }
                  />
                </div>
              </div>

              <div className="flex flex-wrap gap-x-5 gap-y-2">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    className="size-4 accent-primary"
                    checked={form.logger.localtime}
                    onChange={(event) =>
                      updateLogger({ localtime: event.target.checked })
                    }
                  />
                  使用本地时间
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    className="size-4 accent-primary"
                    checked={form.logger.compress}
                    onChange={(event) =>
                      updateLogger({ compress: event.target.checked })
                    }
                  />
                  压缩历史日志
                </label>
              </div>
            </div>
          </>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            <ReadonlyField label="级别" value={form.logger.level} />
            <ReadonlyField
              label="控制台输出"
              value={form.logger.console ? "开" : "关"}
            />
            <ReadonlyField label="日志输出文件" value={form.logger.filename} mono />
            <ReadonlyField
              label="日志文件分割大小 MB"
              value={String(form.logger.maxsize)}
              mono
            />
            <ReadonlyField label="保留天数" value={String(form.logger.maxage)} mono />
            <ReadonlyField
              label="文件数上限"
              value={String(form.logger.maxbackups)}
              mono
            />
            <ReadonlyField
              label="使用本地时间"
              value={form.logger.localtime ? "是" : "否"}
            />
            <ReadonlyField
              label="压缩历史日志"
              value={form.logger.compress ? "是" : "否"}
            />
          </div>
        )}
      </section>

      <ConfigDiff sections={diffSections} changed={changed} />
    </div>
  )
}

function StaticInput({
  value,
  placeholder,
  onCommit,
}: {
  value: string
  placeholder: string
  onCommit: (value: string) => void
}) {
  const [local, setLocal] = useState(value)

  useEffect(() => {
    setLocal(value)
  }, [value])

  return (
    <Input
      value={local}
      placeholder={placeholder}
      onChange={(event) => setLocal(event.target.value)}
      onBlur={() => onCommit(local.trim())}
      onKeyDown={(event) => {
        if (event.key === "Enter") (event.target as HTMLInputElement).blur()
      }}
    />
  )
}

function ConfigDiff({
  sections,
  changed,
}: {
  sections: DiffSection[]
  changed: boolean
}) {
  return (
    <div className="grid gap-3 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium">配置生效验证</span>
        {changed ? (
          <TooltipHint content="存在未生效的变更，保存并重启管理中心端后才会生效">
            <span className="text-destructive inline-flex">
              <TriangleAlert className="size-3.5" />
            </span>
          </TooltipHint>
        ) : (
          <TooltipHint content="当前保存的配置与运行中配置一致">
            <span className="inline-flex text-green-600 dark:text-green-400">
              <Check className="size-3.5" />
            </span>
          </TooltipHint>
        )}
      </div>

      {changed && (
        <p className="text-muted-foreground text-xs">
          当前保存的配置与运行中配置不一致，需重启管理中心端后才会生效。
        </p>
      )}

      <div className="grid gap-3">
        {sections.map((section) => (
          <div key={section.title} className="grid gap-2">
            <div className="text-muted-foreground text-[11px] font-medium">
              {section.title}
            </div>
            <div className="overflow-hidden rounded-md border">
              <div className="bg-muted/40 text-muted-foreground grid grid-cols-[0.7fr_1fr_1fr] border-b px-2 py-1 text-[11px]">
                <span>配置项</span>
                <span>已保存配置</span>
                <span>生效配置</span>
              </div>
              {section.rows.map((row) => {
                const diff = row.saved !== row.effect
                return (
                  <div
                    key={row.label}
                    className="grid grid-cols-[0.7fr_1fr_1fr] border-b px-2 py-1.5 text-xs last:border-b-0"
                  >
                    <span className="text-muted-foreground truncate pr-2">
                      {row.label}
                    </span>
                    <TooltipHint content={row.saved || "—"}>
                      <span className="truncate pr-2 font-mono">{row.saved || "—"}</span>
                    </TooltipHint>
                    <span
                      className={cn(
                        "truncate font-mono",
                        diff &&
                          "bg-destructive/15 rounded px-1 font-semibold text-destructive"
                      )}
                    >
                      {row.effect || "—"}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function ReadonlyField({
  label,
  value,
  mono = false,
}: {
  label: string
  value?: string
  mono?: boolean
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="grid content-start gap-0.5">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={text}>
        <div className={cn("truncate text-xs", mono && "font-mono")}>{text}</div>
      </TooltipHint>
    </div>
  )
}