import { DetailField } from "@/components/detail-field"
import { Badge } from "@/components/ui/badge"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { formatBytes } from "@/features/broker/types"
import { MINION_STATUSES, MINION_STATUS_COLORS } from "@/features/minion/types"
import type { MinionDetail } from "@/features/minion/types"

export function MinionBasicInfo({ detail }: { detail: MinionDetail }) {
  const status = MINION_STATUSES[detail.status] ?? "未知"
  const statusColor =
    MINION_STATUS_COLORS[detail.status] ?? "bg-muted-foreground/50"
  const platform = [detail.goos, detail.arch].filter(Boolean).join("/")

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <section className="grid gap-3">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">节点信息</h3>
          <div className="flex items-center gap-2">
            <TooltipHint content={status}>
              <span
                role="status"
                aria-label={status}
                className="inline-flex shrink-0 items-center"
              >
                <span className={cn("size-2 rounded-full", statusColor)} />
              </span>
            </TooltipHint>
            <span className="text-xs">{status}</span>
          </div>
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <DetailField label="IPv4" value={detail.inet} mono />
          <DetailField label="IPv6" value={detail.inet6} mono />
          <DetailField label="Mac" value={detail.mac} mono />
          <DetailField label="主机名" value={detail.hostname} />
          <DetailField label="主机 ID" value={detail.host_id} mono />
          <DetailField label="Agent 版本" value={detail.edition} />
          <DetailField
            label="中继节点"
            value={
              detail.broker_slug
                ? `${detail.broker_name || detail.broker_slug}（${detail.broker_slug}）`
                : detail.broker_name
            }
          />
          <DetailField label="最近上线时间" value={formatDateTime(detail.uptime)} />
          <DetailField
            label="首次安装时间"
            value={formatDateTime(detail.created_at)}
          />
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">系统信息</h3>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <DetailField label="系统" value={detail.release} />
          <DetailField label="系统发行版本" value={detail.kernel_version} />
          <DetailField label="平台" value={platform} />
          <DetailField label="虚拟化" value={detail.virtual_sys} />
          <DetailField label="虚拟化角色" value={detail.virtual_role} />
          <DetailField
            label="启动时间"
            value={formatDateTime(
              detail.boot_at ? new Date(detail.boot_at * 1000).toISOString() : ""
            )}
          />
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">资源信息</h3>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <DetailField
            label="CPU 核数"
            value={detail.cpu_core ? String(detail.cpu_core) : undefined}
          />
          <DetailField
            label="内存"
            value={detail.mem_total ? formatBytes(detail.mem_total) : undefined}
          />
          <DetailField
            label="空闲内存"
            value={detail.mem_total ? formatBytes(detail.mem_free) : undefined}
          />
          <DetailField
            label="交换分区"
            value={detail.swap_total ? formatBytes(detail.swap_total) : undefined}
          />
          <DetailField
            label="空闲交换分区"
            value={
              detail.swap_total ? formatBytes(detail.swap_free ?? 0) : undefined
            }
          />
          <DetailField
            label="进程数量"
            value={detail.proc_number ? String(detail.proc_number) : undefined}
          />
          <DetailField
            label="守护进程占用"
            value={detail.agent_total ? formatBytes(detail.agent_total) : undefined}
          />
          <DetailField
            label="守护进程分配"
            value={
              detail.agent_total ? formatBytes(detail.agent_alloc ?? 0) : undefined
            }
          />
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">归属信息</h3>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <DetailField label="部门" value={detail.org_path || detail.ibu} />
          <DetailField label="机房" value={detail.idc} />
          <DetailField label="运维负责人" value={detail.op_duty} />
          <DetailField label="部门信息" value={detail.category} />
          <DetailField label="堡垒机用户" value={detail.identity} />
          <DetailField label="定制版本" value={detail.customized} />
        </div>
        <div className="text-muted-foreground flex flex-wrap gap-x-6 gap-y-1 text-xs">
          <span>创建时间 {formatDateTime(detail.created_at)}</span>
          <span>更新时间 {formatDateTime(detail.updated_at)}</span>
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">标签</h3>
        <div className="flex flex-wrap items-center gap-1.5">
          {detail.tags?.length ? (
            detail.tags.map((item) => (
              <Badge key={item.id ?? item.tag} variant="secondary">
                {item.tag}
              </Badge>
            ))
          ) : (
            <span className="text-muted-foreground text-xs">暂无标签</span>
          )}
          {detail.unload && <Badge variant="secondary">静默模式</Badge>}
          {detail.unstable && <Badge variant="outline">内测版本</Badge>}
        </div>
      </section>

      {detail.manual_notes?.length ? (
        <section className="grid gap-3">
          <h3 className="text-sm font-semibold">人工备注</h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {detail.manual_notes.map((note) => (
              <DetailField
                key={note.field}
                label={note.field}
                value={note.value}
              />
            ))}
          </div>
        </section>
      ) : null}

      {detail.comment ? (
        <section className="grid gap-2">
          <h3 className="text-sm font-semibold">节点描述</h3>
          <p className="text-muted-foreground text-xs break-all whitespace-pre-wrap">
            {detail.comment}
          </p>
        </section>
      ) : null}
    </div>
  )
}