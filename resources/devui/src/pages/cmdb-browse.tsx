import {
  Fragment,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react"
import { useNavigate } from "react-router-dom"
import {
  ArrowLeft,
  Building2,
  ChevronDown,
  Database,
  Loader2,
  Search,
  Server,
  Shield,
  ShieldAlert,
  ShieldCheck,
  ShieldX,
} from "lucide-react"

import { DataPagination } from "@/components/data-pagination"
import { TooltipHint } from "@/components/tooltip-hint"
import { DepartmentTree } from "@/features/cmdb/department-tree"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { cmdbApi } from "@/features/cmdb/api"
import {
  CI_TYPE_OPTIONS,
  ciTypeLabel,
  countDepartmentNodes,
  type CMDBAsset,
  type CMDBDepartmentNode,
  type SimpleDuty,
  type SSOCAgentSimple,
} from "@/features/cmdb/types"

export function CMDMBrowsePage() {
  const navigate = useNavigate()

  const [departments, setDepartments] = useState<CMDBDepartmentNode[]>([])
  const [treeLoading, setTreeLoading] = useState(true)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [selected, setSelected] = useState("")

  const [keyword, setKeyword] = useState("")
  const [query, setQuery] = useState("")
  const [ciTypes, setCiTypes] = useState<string[]>([])
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(24)
  const [assets, setAssets] = useState<CMDBAsset[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<CMDBAsset | null>(null)

  useEffect(() => {
    let cancelled = false
    cmdbApi
      .departments()
      .then((data) => {
        if (cancelled) return
        const roots = (data.records ?? []).filter(
          (node) => node.name || node.department
        )
        setDepartments(roots)
      })
      .catch((err) => {
        if (cancelled) return
        toast.error(err instanceof Error ? err.message : "加载 CMDB 部门失败")
      })
      .finally(() => {
        if (!cancelled) setTreeLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const requestId = useRef(0)
  const loadAssets = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const data = await cmdbApi.assets({
        page,
        size,
        q: query || undefined,
        department: selected || undefined,
        ciTypes,
      })
      if (id !== requestId.current) return
      setAssets(data.records ?? [])
      setTotal(data.total ?? 0)
    } catch (err) {
      if (id !== requestId.current) return
      setAssets([])
      setTotal(0)
      toast.error(err instanceof Error ? err.message : "加载 CMDB 资产失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [page, size, query, selected, ciTypes])

  useEffect(() => {
    void loadAssets()
  }, [loadAssets])

  const departmentCount = useMemo(
    () => countDepartmentNodes(departments),
    [departments]
  )
  const currentDepartment = selected || "全部部门"

  const selectDepartment = (department: string) => {
    setSelected(department)
    setPage(1)
  }

  const toggleDepartment = (department: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(department)) next.delete(department)
      else next.add(department)
      return next
    })
  }

  const search = () => {
    const next = keyword.trim()
    if (page === 1 && next === query) {
      void loadAssets()
      return
    }
    setPage(1)
    setQuery(next)
  }

  const toggleCiType = (value: string) => {
    setCiTypes((prev) =>
      prev.includes(value)
        ? prev.filter((item) => item !== value)
        : [...prev, value]
    )
    setPage(1)
  }

  const typeLabel =
    ciTypes.length === 0
      ? "全部类型"
      : ciTypes.length === 1
        ? ciTypeLabel(ciTypes[0])
        : `已选 ${ciTypes.length} 类`

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-12 items-center gap-2 border-b px-4">
        <Button
          variant="ghost"
          size="icon"
          aria-label="返回"
          onClick={() => navigate("/configs/cmdb")}
        >
          <ArrowLeft />
        </Button>
        <Database className="text-muted-foreground size-4 shrink-0" />
        <h2 className="min-w-0 truncate text-sm font-semibold">CMDB 数据浏览</h2>
        <TooltipHint content={currentDepartment}>
          <span className="text-muted-foreground max-w-64 truncate text-xs">
            {currentDepartment}
          </span>
        </TooltipHint>
      </div>

      <div className="flex min-h-0 flex-1">
        <aside className="flex w-60 shrink-0 flex-col border-r">
          <div className="flex h-12 items-center justify-between gap-2 border-b px-3">
            <div className="flex items-center gap-1.5">
              <Building2 className="text-muted-foreground size-3.5" />
              <span className="text-xs font-semibold">部门</span>
            </div>
            <span className="text-muted-foreground text-[10px]">
              {treeLoading ? "加载中…" : `${departmentCount} 个`}
            </span>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            {treeLoading ? (
              <div className="text-muted-foreground flex items-center gap-1.5 px-2 py-3 text-xs">
                <Loader2 className="size-3.5 animate-spin" />
                正在加载部门…
              </div>
            ) : departments.length === 0 ? (
              <p className="text-muted-foreground px-2 py-3 text-xs">暂无部门数据</p>
            ) : (
              <DepartmentTree
                nodes={departments}
                depth={0}
                expanded={expanded}
                selected={selected}
                onSelect={selectDepartment}
                onToggle={toggleDepartment}
              />
            )}
          </div>
        </aside>

        <section className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-12 flex-wrap items-center gap-3 border-b px-4">
            <div className="flex shrink-0 items-center gap-1.5">
              <Server className="text-muted-foreground size-3.5" />
              <span className="text-xs font-semibold">资产</span>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground flex h-7 shrink-0 items-center gap-1 rounded-md border px-2 text-xs transition-colors"
                >
                  {typeLabel}
                  <ChevronDown className="size-3.5 opacity-70" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-40">
                <DropdownMenuCheckboxItem
                  checked={ciTypes.length === 0}
                  className="text-xs"
                  onSelect={(event) => event.preventDefault()}
                  onCheckedChange={() => {
                    setCiTypes([])
                    setPage(1)
                  }}
                >
                  全部类型
                </DropdownMenuCheckboxItem>
                {CI_TYPE_OPTIONS.map((option) => (
                  <DropdownMenuCheckboxItem
                    key={option.value}
                    checked={ciTypes.includes(option.value)}
                    className="text-xs"
                    onSelect={(event) => event.preventDefault()}
                    onCheckedChange={() => toggleCiType(option.value)}
                  >
                    {option.label}
                  </DropdownMenuCheckboxItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
            <div className="flex min-w-0 flex-1 items-center justify-end">
              <div className="relative min-w-0 flex-1 sm:max-w-xs">
                <Input
                  value={keyword}
                  placeholder="搜索名称 / IP / 应用 / 部门…"
                  className="h-7 pl-3 pr-9"
                  onChange={(event) => setKeyword(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") search()
                  }}
                />
                <button
                  type="button"
                  aria-label="搜索"
                  className="text-muted-foreground hover:text-foreground absolute inset-y-0 right-0 flex w-8 items-center justify-center"
                  onClick={search}
                >
                  <Search className="size-4" />
                </button>
              </div>
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-auto">
            <Table className="table-fixed">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[10%]">类型</TableHead>
                  <TableHead className="w-[16%]">IP</TableHead>
                  <TableHead className="w-[10%]">SSOC</TableHead>
                  <TableHead className="w-[20%]">部门</TableHead>
                  <TableHead className="w-[12%]">机房</TableHead>
                  <TableHead className="w-[32%]">备注</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center">
                      <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
                        <Loader2 className="size-3.5 animate-spin" />
                        正在加载资产…
                      </span>
                    </TableCell>
                  </TableRow>
                ) : assets.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-muted-foreground h-24 text-center text-xs">
                      暂无资产数据
                    </TableCell>
                  </TableRow>
                ) : (
                  assets.map((asset) => (
                    <TableRow
                      key={asset.id || asset.ci_id}
                      tabIndex={0}
                      className="hover:bg-muted/40 cursor-pointer transition-colors"
                      onClick={() => setDetail(asset)}
                      onKeyDown={(event) => {
                        if (event.key === "Enter" || event.key === " ") {
                          event.preventDefault()
                          setDetail(asset)
                        }
                      }}
                    >
                      <TableCell>{assetCell(ciTypeLabel(asset.ci_type))}</TableCell>
                      <TableCell>{assetCell(assetIps(asset), true)}</TableCell>
                      <TableCell>{ssocCell(asset.ssoc_agent)}</TableCell>
                      <TableCell>{assetCell(asset.department)}</TableCell>
                      <TableCell>{assetCell(asset.idc)}</TableCell>
                      <TableCell>{assetCell(asset.comment)}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          <div className="border-t px-4 py-2">
            <DataPagination
              current={page}
              size={size}
              total={total}
              loading={loading}
              onPageChange={setPage}
              onPageSizeChange={(next) => {
                setSize(next)
                setPage(1)
              }}
            />
          </div>
        </section>
      </div>

      <AssetDetailDialog asset={detail} onClose={() => setDetail(null)} />
    </div>
  )
}

function AssetDetailDialog({
  asset,
  onClose,
}: {
  asset: CMDBAsset | null
  onClose: () => void
}) {
  const duties = [
    { label: "应用负责人", value: dutyText(asset?.app_duty) },
    { label: "运维负责人", value: dutyText(asset?.op_duty) },
    { label: "系统负责人", value: dutyText(asset?.sys_duty) },
  ]

  return (
    <Dialog
      open={asset !== null}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogContent className="sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle className="flex min-w-0 items-center gap-2 pr-6">
            <span className="truncate">{asset?.name || "资产详情"}</span>
            {asset && (
              <Badge variant="secondary" className="shrink-0 text-[11px]">
                {ciTypeLabel(asset.ci_type)}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>CI ID {asset?.ci_id || "—"}</DialogDescription>
        </DialogHeader>

        {asset && (
          <div className="grid max-h-[75vh] gap-4 overflow-y-auto pr-1">
            <SSOCDetailSection agent={asset.ssoc_agent} />

            {asset.ci_type === "server" ? (
              <ServerDetailSections asset={asset} />
            ) : (
              <>
                <DetailSection
                  title="基本信息"
                  items={[
                    { label: "部门", value: asset.department },
                    { label: "应用", value: asset.app_name },
                    { label: "业务状态", value: asset.status },
                    { label: "机房", value: asset.idc },
                    { label: "机柜", value: asset.rack },
                    { label: "区域", value: asset.region },
                    { label: "可用区", value: asset.zone },
                    { label: "备注", value: asset.comment },
                  ]}
                />

                <DetailSection
                  title="网络"
                  items={[
                    {
                      label: "内网 IP",
                      value: ipText(asset.private_ip),
                      mono: true,
                    },
                    {
                      label: "宿主机 IP",
                      value: ipText(asset.host_ip),
                      mono: true,
                    },
                    { label: "IPv6", value: ipText(asset.ipv6), mono: true },
                    {
                      label: "浮动 IP",
                      value: ipText(asset.float_ip),
                      mono: true,
                    },
                    { label: "电信 IP", value: asset.ctc_ip, mono: true },
                    {
                      label: "私有云 IP",
                      value: asset.private_cloud_ip,
                      mono: true,
                    },
                    {
                      label: "公有云 ID",
                      value: asset.public_cloud_id,
                      mono: true,
                    },
                  ]}
                />

                <DetailSection
                  title="硬件与系统"
                  items={[
                    { label: "主机名", value: asset.hostname },
                    {
                      label: "操作系统",
                      value: [asset.os_name, asset.os_version]
                        .filter(Boolean)
                        .join(" "),
                    },
                    { label: "内核版本", value: asset.kernel_version },
                    { label: "架构", value: asset.os_arch },
                    { label: "CPU", value: asset.cpu },
                    { label: "CPU 核数", value: cpuText(asset) },
                    { label: "内存", value: asset.ram_size || asset.ram },
                    { label: "硬盘", value: asset.harddisk },
                    {
                      label: "镜像",
                      value: [asset.image_name, asset.image_version]
                        .filter(Boolean)
                        .join(" "),
                    },
                    { label: "SN", value: asset.sn, mono: true },
                    { label: "UUID", value: asset.uuid, mono: true },
                    { label: "实例 ID", value: asset.instance_id, mono: true },
                  ]}
                />

                <DetailSection title="责任人" items={duties} />
              </>
            )}

            <DetailSection
              title="资产时间"
              items={[
                { label: "CMDB 创建时间", value: asset.created_time, mono: true },
                { label: "CMDB 更新时间", value: asset.update_time, mono: true },
                { label: "SSOC 导入时间", value: timeText(asset.created_at) },
                { label: "SSOC 更新时间", value: timeText(asset.updated_at) },
              ]}
            />
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

function ServerDetailSections({ asset }: { asset: CMDBAsset }) {
  return (
    <>
      <DetailSection
        title="基本信息"
        items={[
          { label: "SN", value: asset.sn, mono: true },
          { label: "部门", value: asset.department },
          { label: "业务", value: asset.business },
          { label: "付款方", value: asset.payer },
          { label: "业务状态", value: asset.status },
          { label: "资产状态", value: asset.asset_status },
          { label: "资产编号", value: asset.asset_id },
          { label: "作用", value: asset.comment },
          { label: "标签", value: listText(asset.tags) },
        ]}
      />

      <DetailSection
        title="位置"
        items={[
          { label: "IDC", value: asset.idc },
          { label: "机房", value: asset.server_room },
          { label: "机柜", value: asset.rack },
        ]}
      />

      <DetailSection
        title="硬件配置"
        items={[
          { label: "CPU型号", value: asset.cpu },
          { label: "CPU核数", value: asset.cpu_count },
          { label: "逻辑CPU数量", value: asset.logic_cpu_count },
          { label: "超线程", value: asset.hyper_threading },
          { label: "用途", value: asset.use },
          { label: "内存信息", value: asset.ram },
          { label: "内存大小", value: asset.ram_size },
          { label: "硬盘信息", value: asset.harddisk },
          { label: "磁盘阵列", value: asset.raid },
        ]}
      />

      <DetailSection
        title="网络"
        items={[
          { label: "内网IP", value: ipText(asset.private_ip), mono: true },
          { label: "内网IPv6", value: ipText(asset.ipv6), mono: true },
          { label: "电信IP", value: asset.ctc_ip, mono: true },
          { label: "网通IP", value: asset.cnc_ip, mono: true },
          { label: "移动IP", value: asset.cmc_ip, mono: true },
          { label: "主机名", value: asset.hostname },
          { label: "浮动IP", value: ipText(asset.float_ip), mono: true },
          { label: "网络开放状态", value: asset.net_open },
        ]}
      />

      <DetailSection
        title="系统"
        items={[
          { label: "操作系统版本", value: asset.os_version },
          { label: "内核版本", value: asset.kernel_version },
          { label: "环境", value: asset.env },
        ]}
      />

      <DetailSection
        title="负责人"
        items={[
          { label: "开发负责人", value: dutyText(asset.rd_duty) },
          { label: "运维负责人", value: dutyText(asset.op_duty) },
          { label: "系统负责人", value: dutyText(asset.sys_duty) },
          { label: "应用负责人", value: dutyText(asset.app_duty) },
        ]}
      />

      <DetailSection
        title="应用与安全"
        items={[
          { label: "应用名称", value: asset.app_name },
          { label: "业务属性", value: asset.trade_type },
          { label: "宿主机类型", value: asset.host_type },
          { label: "堡垒机账号", value: asset.baoleiji_identity },
          { label: "安全信息", value: asset.security_info },
        ]}
      />
    </>
  )
}

function SSOCDetailSection({ agent }: { agent?: SSOCAgentSimple | null }) {
  return (
    <DetailSection
      title="SSOC 探针"
      items={[
        { label: "状态", value: ssocStatusBadge(agent) },
        { label: "版本", value: agent?.version },
        {
          label: "中继节点",
          value: agent?.broker_name || agent?.broker_slug,
        },
        { label: "机器码", value: agent?.machine_id, mono: true },
      ]}
    />
  )
}

interface DetailItem {
  label: string
  value?: React.ReactNode
  mono?: boolean
}

function isDetailItemVisible(item: DetailItem) {
  return item.value !== undefined && item.value !== null && item.value !== ""
}

function DetailSection({
  title,
  items,
}: {
  title: string
  items: DetailItem[]
}) {
  const visible = items.filter(isDetailItemVisible)
  if (visible.length === 0) return null

  const rows: (DetailItem | undefined)[][] = []
  for (let i = 0; i < visible.length; i += 2) {
    rows.push([visible[i], visible[i + 1]])
  }

  return (
    <section className="grid gap-2">
      <h3 className="text-muted-foreground text-[11px] font-medium">{title}</h3>
      <div className="overflow-hidden rounded-md border">
        <table className="w-full table-fixed text-xs">
          <tbody>
            {rows.map((row, rowIndex) => (
              <tr key={rowIndex} className={cn(rowIndex > 0 && "border-t")}>
                {row.map((item, pairIndex) => (
                  <Fragment key={pairIndex}>
                    <td
                      className={cn(
                        "bg-muted/40 text-muted-foreground w-[11%] px-2 py-1.5 align-top",
                        pairIndex > 0 && "border-l"
                      )}
                    >
                      {item?.label ?? ""}
                    </td>
                    <td className="w-[39%] px-2 py-1.5 align-top">
                      {item ? (
                        <span
                          className={cn(
                            "block break-all",
                            item.mono && "font-mono"
                          )}
                        >
                          {item.value}
                        </span>
                      ) : null}
                    </td>
                  </Fragment>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}

function assetIps(asset: CMDBAsset): string {
  return (asset.private_ip ?? []).join(", ")
}

function ssocCell(agent?: SSOCAgentSimple | null) {
  const detail = agent
    ? [agent.version, agent.broker_name].filter(Boolean).join(" · ")
    : ""

  return (
    <TooltipHint
      content={agent ? detail || "已安装 SSOC 探针" : "未安装 SSOC 探针"}
    >
      {ssocStatusBadge(agent)}
    </TooltipHint>
  )
}

function ssocStatusBadge(agent?: SSOCAgentSimple | null) {
  if (!agent) {
    return (
      <span className="text-muted-foreground inline-flex h-4 items-center gap-1.5 align-middle text-xs leading-none">
        <Shield className="size-3.5 shrink-0" />
        未安装
      </span>
    )
  }

  const meta = ssocStatusMeta(agent.status)
  const Icon = meta.icon
  return (
    <span
      className={cn(
        "inline-flex h-4 items-center gap-1.5 align-middle text-xs leading-none",
        meta.className
      )}
    >
      <Icon className="size-3.5 shrink-0" />
      {meta.label}
    </span>
  )
}

function ssocStatusMeta(status?: number) {
  switch (status) {
    case 1:
      return {
        label: "在线",
        icon: ShieldCheck,
        className: "text-green-600 dark:text-green-400",
      }
    case 2:
      return {
        label: "删除",
        icon: ShieldX,
        className: "text-red-600 dark:text-red-400",
      }
    default:
      return {
        label: "离线",
        icon: ShieldAlert,
        className: "text-yellow-600 dark:text-yellow-400",
      }
  }
}

function assetCell(value?: string, mono = false) {
  const text = value && value.trim() ? value : "—"
  return (
    <TooltipHint content={text}>
      <span
        className={cn(
          "block truncate text-xs",
          mono && "font-mono",
          text === "—" && "text-muted-foreground"
        )}
      >
        {text}
      </span>
    </TooltipHint>
  )
}

function listText(values?: string[]) {
  const items = (values ?? []).map((value) => value.trim()).filter(Boolean)
  return items.length ? items.join(", ") : undefined
}

function timeText(value?: string) {
  if (!value || value.startsWith("0001-01-01")) return undefined
  return formatDateTime(value)
}

function ipText(values?: string[]) {
  const items = (values ?? []).map((value) => value.trim()).filter(Boolean)
  if (items.length === 0) return undefined
  return (
    <span className="grid gap-0.5">
      {items.map((ip, index) => (
        <span key={`${ip}-${index}`} className="block">
          {ip}
        </span>
      ))}
    </span>
  )
}

function dutyText(values?: SimpleDuty[]) {
  const items = (values ?? [])
    .map((duty) => duty.name?.trim() || duty.job_number?.trim())
    .filter(Boolean)
  return items.length ? items.join("、") : undefined
}

function cpuText(asset: CMDBAsset) {
  const physical = asset.cpu_count ?? 0
  const logical = asset.logic_cpu_count ?? 0
  if (physical && logical) return `${physical} 核 / 逻辑 ${logical} 核`
  if (physical) return `${physical} 核`
  if (logical) return `逻辑 ${logical} 核`
  return undefined
}