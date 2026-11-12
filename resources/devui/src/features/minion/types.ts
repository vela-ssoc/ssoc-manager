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

export interface MinionTag {
  id?: string
  minion_id?: string
  tag: string
  kind?: number
}

export interface MinionManualNote {
  field: string
  value: string
}

export interface MinionDetail extends Omit<Minion, "tags"> {
  tags?: MinionTag[]
  inet6?: string
  mac?: string
  arch?: string
  broker_id?: string
  broker_slug?: string
  org_path?: string
  identity?: string
  category?: string
  op_duty?: string
  manual_notes?: MinionManualNote[]
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

export interface MinionNode {
  id: string
  machine_id: string
  inet: string
  inet6?: string
  mac?: string
  goos: string
  arch?: string
  edition: string
  status: MinionStatus
  broker_id?: string
  broker_name?: string
  broker_slug?: string
  unload: boolean
  unstable: boolean
  customized?: string
  org_path?: string
  identity?: string
  category?: string
  op_duty?: string
  comment?: string
  ibu?: string
  idc?: string
  os_release?: string
  heartbeat_at?: string
  created_at?: string
  updated_at?: string
}

export interface MinionNodePage {
  page: number
  size: number
  total: number
  records: MinionNode[]
}

export interface MinionFilter {
  key: string
  operator: string
  value: string | number
}

export interface AgentExecuteStat {
  goos?: string
  goarch?: string
  version?: string
  unstable?: boolean
  customized?: string
  pid?: number
  args?: string[]
  hostname?: string
  workdir?: string
  executable?: string
}

export interface AgentTunnelStat {
  library?: string
  module?: string
  inet?: string
  connected_at?: string
  cumulative?: number
  local_addr?: string
  remote_addr?: string
  receive_bytes?: number
  transmit_bytes?: number
}

export interface AgentBrokerEndpoint {
  id?: string
  idv1?: number
  slug?: string
  name?: string
}

export interface AgentConnectHistory {
  id: string
  agent_idv1: number
  execute_stat?: AgentExecuteStat
  tunnel_stat?: AgentTunnelStat
  broker?: AgentBrokerEndpoint
  seconds: number
  disconnected_at?: string
}

export interface AgentConnectHistoryPage {
  page: number
  size: number
  total: number
  records: AgentConnectHistory[]
}

export interface AgentSubstream {
  id: string
  rx: number
  tx: number
  established_at: string
}

export interface AgentSubstreams {
  records: AgentSubstream[]
}

export const MINION_STATUSES: Record<MinionStatus, string> = {
  1: "未激活",
  2: "离线",
  3: "在线",
  4: "已删除",
}

export const MINION_STATUS_COLORS: Record<MinionStatus, string> = {
  1: "bg-amber-500",
  2: "bg-muted-foreground/50",
  3: "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]",
  4: "bg-destructive",
}

// 在线状态。探针侧接口（活跃子流等）只有在线时才可用。
export const MINION_STATUS_ONLINE: MinionStatus = 3
