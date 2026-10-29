export interface ProcessMemory {
  rss: number
  vms: number
  hwm: number
  data: number
  stack: number
  locked: number
  swap: number
}

export interface ProcessInfo {
  pid: number
  name: string
  cwd: string
  exe: string
  cmdline: string
  username: string
  cpu_percent: number
  num_threads: number
  num_fds: number
  created_at: string
  memory: ProcessMemory | null
}

export interface ProcessConnectionAddress {
  ip: string
  port: number
}

export interface ProcessConnection {
  fd: number
  family: number
  type: number
  localaddr: ProcessConnectionAddress
  remoteaddr: ProcessConnectionAddress
  status: string
  uids: number[]
  pid: number
}

export interface ProcessOpenFile {
  path: string
  fd: number
}

export interface ProcessRecords<T> {
  records: T[]
}