export type MinionStatus = 1 | 2 | 3 | 4

export interface Minion {
  id: string
  inet: string
  goos: string
  edition: string
  status: MinionStatus
  heartbeat_at: string
  cpu_core: number
  mem_total: number
  mem_free: number
  idc: string
  ibu: string
  comment: string
  broker_name: string
  unload: boolean
  uptime: string
  unstable: boolean
  customized: string
  tags: string[]
}

export interface MinionDetail extends Minion {
  inet6?: string
  mac?: string
  arch?: string
  broker_id?: string
  org_path?: string
  identity?: string
  category?: string
  op_duty?: string
  created_at?: string
  updated_at?: string
  release?: string
  swap_total?: number
  swap_free?: number
  host_id?: string
  family?: string
  version?: string
  boot_at?: number
  virtual_sys?: string
  virtual_role?: string
  proc_number?: number
  hostname?: string
  kernel_version?: string
  agent_total?: number
  agent_alloc?: number
}

export interface MinionPage {
  total: number
  current: number
  size: number
  records: Minion[]
}

export interface MinionFilter {
  key: string
  operator: string
  value: string | number
}

export const MINION_STATUSES: Record<MinionStatus, string> = {
  1: "未激活",
  2: "离线",
  3: "在线",
  4: "已删除",
}
