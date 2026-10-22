import { useEffect, useState } from "react"
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

import { ConfigSection } from "./config-editor"
import type { BrokerConfig, BrokerForm, ExposeAddress, ManagerConfig } from "./types"

const schema = z.object({
  slug: z
    .string()
    .trim()
    .min(1, "标识不能为空")
    .max(20, "标识不能超过 20 字"),
  name: z
    .string()
    .trim()
    .min(2, "名称至少 2 个字符")
    .max(20, "名称不能超过 20 字"),
})

function emptyExpose(): ExposeAddress {
  return { name: "", addr: "" }
}

export function BrokerDialog({
  open,
  onOpenChange,
  managerConfig,
  submitting,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  managerConfig: ManagerConfig | null
  submitting: boolean
  onSubmit: (form: BrokerForm) => void
}) {
  const [slug, setSlug] = useState("")
  const [name, setName] = useState("")
  const [exposes, setExposes] = useState<ExposeAddress[]>([emptyExpose()])
  const [serverConfig, setServerConfig] = useState<BrokerConfig["server"]>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!open) return
    setSlug("")
    setName("")
    setExposes([emptyExpose()])
    setServerConfig(null)
    setErrors({})
  }, [open])

  const updateExpose = (index: number, patch: Partial<ExposeAddress>) =>
    setExposes((rows) => rows.map((row, i) => (i === index ? { ...row, ...patch } : row)))

  const save = () => {
    const parsed = schema.safeParse({ slug, name })
    const validExposes = exposes.filter(
      (item) => item.name.trim() && item.addr.trim()
    )
    if (!parsed.success || validExposes.length === 0) {
      const next: Record<string, string> = {}
      if (!parsed.success) {
        for (const issue of parsed.error.issues) {
          const field = String(issue.path[0] ?? "")
          if (!next[field]) next[field] = issue.message
        }
      }
      if (validExposes.length === 0) next.exposes = "请至少填写一个暴露地址"
      setErrors(next)
      return
    }
    onSubmit({
      slug,
      name,
      exposes: validExposes,
      config: { server: serverConfig, logger: null, database: null },
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90svh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>新增中继节点</DialogTitle>
          <DialogDescription>配置节点标识、名称及暴露地址</DialogDescription>
        </DialogHeader>

        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <Label htmlFor="broker-slug">唯一标识</Label>
            <Input
              id="broker-slug"
              value={slug}
              placeholder="提交后不可修改"
              onChange={(event) => setSlug(event.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              节点的唯一标识，提交后不可修改
            </p>
            {errors.slug && <p className="text-destructive text-xs">{errors.slug}</p>}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="broker-name">名称</Label>
            <Input
              id="broker-name"
              value={name}
              placeholder="便于阅读理解的名称"
              onChange={(event) => setName(event.target.value)}
            />
            {errors.name && <p className="text-destructive text-xs">{errors.name}</p>}
          </div>

          <div className="grid gap-1.5">
            <div className="flex items-center gap-1.5">
              <Label>暴露地址</Label>
              <InfoTooltip label="暴露地址说明">
                <ul>
                  <li>该地址会下发到探针节点配置中，供探针节点连接并上线。</li>
                  <li>探针节点会按列表顺序依次尝试连接，直到成功为止。</li>
                </ul>
              </InfoTooltip>
            </div>
            <div className="grid gap-2">
              {exposes.map((row, index) => (
                <div key={index} className="grid grid-cols-[1fr_1.5fr_auto] gap-2">
                  <Input
                    aria-label={`暴露地址 ${index + 1} 名称`}
                    value={row.name}
                    placeholder="名称"
                    onChange={(event) => updateExpose(index, { name: event.target.value })}
                  />
                  <Input
                    aria-label={`暴露地址 ${index + 1} 地址`}
                    value={row.addr}
                    placeholder="例如 broker.example.com:9000"
                    onChange={(event) => updateExpose(index, { addr: event.target.value })}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-8 text-muted-foreground hover:text-destructive"
                    aria-label="移除地址"
                    disabled={exposes.length <= 1}
                    onClick={() =>
                      setExposes((rows) => rows.filter((_, i) => i !== index))
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
                onClick={() => setExposes((rows) => [...rows, emptyExpose()])}
              >
                <Plus /> 添加地址
              </Button>
            </div>
            {errors.exposes && (
              <p className="text-destructive text-xs">{errors.exposes}</p>
            )}
          </div>

          <div className="grid gap-3 border-t pt-3">
            <div className="flex items-center gap-1.5">
              <span className="text-sm font-medium">高级配置</span>
              <InfoTooltip label="高级配置说明">
                <p>修改保存后，需要重启该中继节点服务才会生效。</p>
              </InfoTooltip>
            </div>
            <p className="text-muted-foreground text-xs">
              默认继承中心端配置，更多配置项可在详情页中编辑。
            </p>
            <ConfigSection
              title="服务配置"
              enabled={Boolean(serverConfig)}
              onToggle={(on) => setServerConfig(on ? { addr: "" } : null)}
              inherit={
                managerConfig?.server?.addr ? (
                  <div className="space-y-0.5">
                    <div>继承中心端配置</div>
                    <div>监听地址 {managerConfig.server.addr}</div>
                  </div>
                ) : (
                  "继承中心端配置"
                )
              }
            >
              <div className="grid gap-1.5">
                <Label htmlFor="broker-server-addr">监听地址</Label>
                <Input
                  id="broker-server-addr"
                  value={serverConfig?.addr ?? ""}
                  placeholder="0.0.0.0:9000"
                  onChange={(event) => setServerConfig({ addr: event.target.value })}
                />
              </div>
            </ConfigSection>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={save} disabled={submitting}>
            {submitting ? "创建中…" : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}