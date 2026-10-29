import { useCallback, useEffect, useState } from "react"
import { Info, Power, TriangleAlert } from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import { BuildInfoView } from "@/features/buildinfo/build-info"
import type { BuildInfo } from "@/features/buildinfo/types"
import { useConfirmCountdown } from "@/features/broker/use-confirm-countdown"
import { centralApi } from "./api"

const EXIT_PRESETS = [
  { label: "立即", seconds: 1 },
  { label: "10 秒后", seconds: 10 },
  { label: "30 秒后", seconds: 30 },
  { label: "1 分钟后", seconds: 60 },
  { label: "1 小时后", seconds: 3600 },
]

export function CentralBasicInfo() {
  const [info, setInfo] = useState<BuildInfo | null>(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setInfo(await centralApi.buildInfo())
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载编译信息失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <BuildInfoView info={info} loading={loading} />

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">运维操作</h3>
        <div className="flex flex-wrap items-center gap-2">
          <RestartItselfButton />
        </div>
      </section>
    </div>
  )
}

function RestartItselfButton() {
  const [open, setOpen] = useState(false)
  const [deadline, setDeadline] = useState("")
  const [seconds, setSeconds] = useState(10)
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState("")
  const remaining = useConfirmCountdown(open)

  const loadDeadline = useCallback(async () => {
    try {
      const data = await centralApi.exitDeadline()
      const value = typeof data === "string" ? data : data?.data
      setDeadline(typeof value === "string" && value ? value : "")
    } catch {
      setDeadline("")
    }
  }, [])

  useEffect(() => {
    void loadDeadline()
  }, [loadDeadline])

  useEffect(() => {
    if (!deadline) {
      setCountdown("")
      return
    }
    const update = () => {
      const diff = new Date(deadline).getTime() - Date.now()
      if (diff <= 0) {
        setCountdown("即将执行重启")
        return
      }
      const total = Math.floor(diff / 1000)
      const h = Math.floor(total / 3600)
      const m = Math.floor((total % 3600) / 60)
      const s = total % 60
      setCountdown(
        [h > 0 ? `${h} 时` : "", m > 0 || h > 0 ? `${m} 分` : "", `${s} 秒`]
          .filter(Boolean)
          .join(" ")
      )
    }
    update()
    const timer = setInterval(update, 1000)
    return () => clearInterval(timer)
  }, [deadline])

  const exec = async () => {
    setLoading(true)
    try {
      await centralApi.exit(seconds)
      toast.success("已提交重启，程序退出后由服务管理器拉起")
      setOpen(false)
      await loadDeadline()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "执行重启失败")
    } finally {
      setLoading(false)
    }
  }

  const cancel = async () => {
    setLoading(true)
    try {
      await centralApi.exitCancel()
      toast.success("已取消重启")
      setDeadline("")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "取消重启失败")
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      {deadline ? (
        <Button
          size="sm"
          variant="destructive"
          disabled={loading}
          onClick={() => void cancel()}
        >
          <Power /> 取消重启{countdown ? ` (${countdown})` : ""}
        </Button>
      ) : (
        <Button size="sm" variant="outline" onClick={() => setOpen(true)}>
          <Power /> 重启
        </Button>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>重启系统本体</DialogTitle>
          </DialogHeader>
          <div className="text-muted-foreground flex items-start gap-2 rounded-md border p-2.5 text-xs leading-relaxed">
            <Info className="mt-px size-4 shrink-0" />
            <span>
              此处重启的是<strong className="text-primary font-semibold">管理中心端程序本体</strong>，并非操作系统，程序退出后需要操作系统的服务管理器拉起（需要先将程序注册为系统服务）。
            </span>
          </div>

          {deadline ? (
            <div className="grid gap-3">
              <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
                <TriangleAlert className="size-4 shrink-0" />
                存在定时重启任务，倒计时 {countdown}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" disabled={loading} onClick={() => void cancel()}>
                  取消重启
                </Button>
              </div>
            </div>
          ) : (
            <div className="grid gap-3">
              <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
                <TriangleAlert className="size-4 shrink-0" />
                重启会短暂中断管理中心端服务，可能导致正在进行中的业务异常，请谨慎操作。
              </div>
              <div className="flex flex-wrap justify-center gap-1.5">
                {EXIT_PRESETS.map((preset) => (
                  <button
                    key={preset.seconds}
                    type="button"
                    className={cn(
                      "rounded-md border px-3 py-1.5 text-xs transition-colors",
                      seconds === preset.seconds
                        ? "border-primary bg-primary/10 text-foreground font-medium"
                        : "text-muted-foreground hover:bg-accent/50"
                    )}
                    onClick={() => setSeconds(preset.seconds)}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  disabled={loading}
                  onClick={() => setOpen(false)}
                >
                  取消
                </Button>
                <Button disabled={loading || remaining > 0} onClick={() => void exec()}>
                  {loading ? "处理中…" : remaining > 0 ? `确定 (${remaining})` : "确定"}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}