export interface CMDBDepartmentNode {
  name: string
  ci_id: string
  department: string
  children?: CMDBDepartmentNode[]
}

export interface CMDBDepartmentRecords {
  records: CMDBDepartmentNode[]
}

export interface SimpleDuty {
  job_number?: string
  name?: string
  company?: string
}

export interface SSOCAgentSimple {
  id?: number | string
  status?: number
  machine_id?: string
  version?: string
  broker_slug?: string
  broker_name?: string
}

export interface CMDBAsset {
  id: string
  ci_id: string
  ci_type: string
  name: string
  department: string
  hostname?: string
  private_ip?: string[]
  host_ip?: string[]
  ipv6?: string[]
  float_ip?: string[]
  ctc_ip?: string
  cnc_ip?: string
  cmc_ip?: string
  private_cloud_ip?: string
  public_cloud_id?: string
  net_open?: string
  os_name?: string
  os_version?: string
  os_arch?: string
  kernel_version?: string
  env?: string
  cpu?: string
  cpu_count?: number
  logic_cpu_count?: number
  hyper_threading?: string
  use?: string
  ram?: string
  ram_size?: string
  harddisk?: string
  raid?: string
  image_name?: string
  image_version?: string
  idc?: string
  server_room?: string
  rack?: string
  region?: string
  zone?: string
  comment?: string
  tags?: string[]
  business?: string
  payer?: string
  status?: string
  asset_status?: string
  asset_id?: string
  sn?: string
  uuid?: string
  instance_id?: string
  host_type?: string
  trade_type?: string
  baoleiji_identity?: string
  security_info?: string
  app_name?: string
  app_duty?: SimpleDuty[]
  rd_duty?: SimpleDuty[]
  op_duty?: SimpleDuty[]
  sys_duty?: SimpleDuty[]
  ssoc_agent?: SSOCAgentSimple | null
  created_time?: string
  update_time?: string
  created_at?: string
  updated_at?: string
}

export interface CMDBAssetPage {
  page: number
  size: number
  total: number
  records: CMDBAsset[]
}

export interface CMDBAssetQuery {
  page: number
  size: number
  q?: string
  department?: string
  ciTypes?: string[]
}

export const CI_TYPE_OPTIONS = [
  { value: "server", label: "物理机" },
  { value: "vserver", label: "虚拟机" },
  { value: "fund_vserver", label: "基金虚拟机" },
  { value: "docker", label: "容器" },
  { value: "public_cloud", label: "公有云" },
] as const

const CI_TYPE_LABELS: Record<string, string> = {
  ...Object.fromEntries(
    CI_TYPE_OPTIONS.map((option) => [option.value, option.label])
  ),
  department: "部门",
}

export function ciTypeLabel(type?: string): string {
  if (!type) return "—"
  return CI_TYPE_LABELS[type] ?? type
}

export function countDepartmentNodes(nodes: CMDBDepartmentNode[]): number {
  return nodes.reduce(
    (sum, node) => sum + 1 + countDepartmentNodes(node.children ?? []),
    0
  )
}