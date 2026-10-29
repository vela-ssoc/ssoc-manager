import { http } from "@/lib/http"

export interface VMSample {
  metric: Record<string, string>
  value?: [number, string]
}

export interface VMVectorResponse {
  status: string
  data: { resultType: "vector"; result: VMSample[] }
}

export interface VMPoint {
  metric: Record<string, string>
  values: [number, string][]
}

export interface VMMatrixResponse {
  status: string
  data: { resultType: "matrix"; result: VMPoint[] }
}

const DEFAULT_LABEL = 'appname="ssoc-manager"'

export function metricsLabel(appname: string, instance?: string): string {
  const base = `appname=${JSON.stringify(appname)}`
  return instance ? `${base},instance=${JSON.stringify(instance)}` : base
}

export function instantQuery(query: string, time?: number): Promise<VMVectorResponse> {
  return http.get<VMVectorResponse>("/vmp/api/v1/query", {
    query,
    time: time ? String(time) : undefined,
  })
}

export function rangeQuery(
  query: string,
  start: number,
  end: number,
  step = "30s"
): Promise<VMMatrixResponse> {
  return http.get<VMMatrixResponse>("/vmp/api/v1/query_range", {
    query,
    start: String(start),
    end: String(end),
    step,
  })
}

export function cpuQuery(label = DEFAULT_LABEL): string {
  return `psutil_cpu_usage_percent{${label}}`
}

export function memUsedQuery(label = DEFAULT_LABEL): string {
  return `psutil_memory_bytes_used{${label}}`
}

export function memTotalQuery(label = DEFAULT_LABEL): string {
  return `psutil_memory_bytes_total{${label}}`
}

export function diskTotalQuery(label = DEFAULT_LABEL): string {
  return `psutil_disk_bytes_total{${label}}`
}

export function diskUsedQuery(label = DEFAULT_LABEL): string {
  return `psutil_disk_bytes_used{${label}}`
}

export function diskReadRateQuery(label = DEFAULT_LABEL): string {
  return `sum(rate(psutil_disk_read_bytes_total{${label}}[5m]))`
}

export function diskWriteRateQuery(label = DEFAULT_LABEL): string {
  return `sum(rate(psutil_disk_written_bytes_total{${label}}[5m]))`
}

export function netRecvRateQuery(label = DEFAULT_LABEL): string {
  return `sum(rate(psutil_network_receive_bytes_total{${label}}[5m]))`
}

export function netSentRateQuery(label = DEFAULT_LABEL): string {
  return `sum(rate(psutil_network_transmit_bytes_total{${label}}[5m]))`
}

export function diskReadRateByDeviceQuery(label = DEFAULT_LABEL): string {
  return `rate(psutil_disk_read_bytes_total{${label}}[5m])`
}

export function diskWriteRateByDeviceQuery(label = DEFAULT_LABEL): string {
  return `rate(psutil_disk_written_bytes_total{${label}}[5m])`
}

export function netRecvRateByInterfaceQuery(label = DEFAULT_LABEL): string {
  return `rate(psutil_network_receive_bytes_total{${label}}[5m])`
}

export function netSentRateByInterfaceQuery(label = DEFAULT_LABEL): string {
  return `rate(psutil_network_transmit_bytes_total{${label}}[5m])`
}

export function diskUsedByDeviceQuery(label = DEFAULT_LABEL): string {
  return `psutil_disk_bytes_used{${label}}`
}

export function diskTotalByDeviceQuery(label = DEFAULT_LABEL): string {
  return `psutil_disk_bytes_total{${label}}`
}

export function goGoroutinesQuery(label = DEFAULT_LABEL): string {
  return `go_goroutines{${label}}`
}

export function goHeapInuseQuery(label = DEFAULT_LABEL): string {
  return `go_memstats_heap_inuse_bytes{${label}}`
}

export function goThreadsQuery(label = DEFAULT_LABEL): string {
  return `go_threads{${label}}`
}

export function processOpenFdsQuery(label = DEFAULT_LABEL): string {
  return `process_open_fds{${label}}`
}

const BROKER_LABEL = 'appname="ssoc-broker"'

export function tunnelReceiveRateQuery(label = BROKER_LABEL): string {
  return `rate(tunnel_receive_bytes{${label}}[5m])`
}

export function tunnelTransmitRateQuery(label = BROKER_LABEL): string {
  return `rate(tunnel_transmit_bytes{${label}}[5m])`
}

export function tunnelCumulativeRateQuery(label = BROKER_LABEL): string {
  return `rate(tunnel_cumulative_substreams{${label}}[5m])`
}

export function tunnelActiveQuery(label = BROKER_LABEL): string {
  return `tunnel_active_substreams{${label}}`
}