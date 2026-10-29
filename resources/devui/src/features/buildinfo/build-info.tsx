import { Loader2 } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import type { BuildInfo, BuildModuleInfo } from "./types"

export function BuildInfoView({
  info,
  loading = false,
}: {
  info: BuildInfo | null
  loading?: boolean
}) {
  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">编译信息</h3>
        {loading && !info ? (
          <div className="text-muted-foreground flex items-center gap-2 text-xs">
            <Loader2 className="size-3.5 animate-spin" />
            正在加载编译信息…
          </div>
        ) : info ? (
          <>
            <div className="grid gap-3 sm:grid-cols-2">
              <InfoField label="版本号" value={info.version} />
              <InfoField label="修订版本" value={info.revision} mono />
              <InfoField label="平台" value={[info.goos, info.goarch].filter(Boolean).join("/")} />
              <InfoField label="包路径" value={info.module} mono />
              <InfoField label="运行用户" value={info.username} />
              <InfoField label="工作目录" value={info.workdir} mono />
              <InfoField
                label="提交时间"
                value={info.committed_at ? formatDateTime(info.committed_at) : undefined}
              />
              {info.build_info?.GoVersion && (
                <InfoField label="Go 版本" value={info.build_info.GoVersion} mono />
              )}
            </div>

            <BuildInfoDetail buildInfo={info.build_info} />
          </>
        ) : (
          <div className="text-muted-foreground text-xs">无法获取编译信息</div>
        )}
      </section>
    </div>
  )
}

function BuildInfoDetail({ buildInfo }: { buildInfo?: BuildModuleInfo }) {
  if (!buildInfo) return null

  const main = buildInfo.Main
  const deps = buildInfo.Deps ?? []
  const settings = buildInfo.Settings ?? []

  return (
    <div className="grid gap-3">
      {main?.Path && (
        <div className="grid gap-2">
          <div className="text-muted-foreground text-[11px] font-medium">主模块</div>
          <div className="overflow-hidden rounded-md border">
            <div className="bg-muted/40 text-muted-foreground grid grid-cols-2 border-b px-2 py-1 text-[11px]">
              <span>模块</span>
              <span className="px-2">版本</span>
            </div>
            <div className="grid grid-cols-2 items-center px-2 py-1.5 text-xs">
              <TooltipHint content={main.Path}>
                <span className="truncate font-mono">{main.Path}</span>
              </TooltipHint>
              <span className="min-w-0 truncate px-2 font-mono">
                {main.Version || "（本地）"}
              </span>
            </div>
          </div>
        </div>
      )}

      {settings.length > 0 && (
        <div className="grid gap-2">
          <div className="text-muted-foreground text-[11px] font-medium">
            编译设置（{settings.length}）
          </div>
          <div className="overflow-hidden rounded-md border">
            <div className="bg-muted/40 text-muted-foreground grid grid-cols-[2fr_3fr] border-b px-2 py-1 text-[11px]">
              <span>键</span>
              <span>值</span>
            </div>
            {settings.map((item) => (
              <div
                key={item.Key}
                className="grid grid-cols-[2fr_3fr] items-center border-b px-2 py-1.5 text-xs last:border-b-0"
              >
                <span className="truncate pr-2 font-mono">{item.Key}</span>
                <TooltipHint content={item.Value}>
                  <span className="truncate font-mono">{item.Value || "—"}</span>
                </TooltipHint>
              </div>
            ))}
          </div>
        </div>
      )}

      {deps.length > 0 && (
        <div className="grid gap-2">
          <div className="text-muted-foreground text-[11px] font-medium">
            依赖模块（{deps.length}）
          </div>
          <div className="overflow-hidden rounded-md border">
            <div className="bg-muted/40 text-muted-foreground grid grid-cols-[3fr_2fr_3fr] border-b px-2 py-1 text-[11px]">
              <span>模块</span>
              <span className="px-2">版本</span>
              <span>校验和</span>
            </div>
            {deps.map((dep, index) => (
              <div
                key={`${dep.Path}-${index}`}
                className="grid grid-cols-[3fr_2fr_3fr] items-center border-b px-2 py-1.5 text-xs last:border-b-0"
              >
                <TooltipHint content={dep.Path ?? "—"}>
                  <a
                    href={goPkgDevUrl(dep.Path, dep.Version)}
                    target="_blank"
                    rel="noreferrer"
                    className="truncate font-mono hover:underline"
                  >
                    {dep.Path || "—"}
                  </a>
                </TooltipHint>
                <TooltipHint content={dep.Version ?? "—"}>
                  <span className="min-w-0 truncate px-2 font-mono">
                    {dep.Version || "—"}
                  </span>
                </TooltipHint>
                <TooltipHint content={dep.Sum ?? "—"}>
                  <span className="truncate font-mono">{dep.Sum || "—"}</span>
                </TooltipHint>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function goPkgDevUrl(path?: string, version?: string): string {
  if (!path) return "https://pkg.go.dev/"
  return version
    ? `https://pkg.go.dev/${path}@${version}`
    : `https://pkg.go.dev/${path}`
}

function InfoField({
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