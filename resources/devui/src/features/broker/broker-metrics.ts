import type { MonitorWindow, Series } from "@/features/central/monitoring"
import { metricsLabel, rangeQuery, type VMPoint } from "@/features/central/vm"

export interface BrokerMetricData {
  cpu: Series[]
  mem: Series[]
  net: Series[]
}

interface MetricSample {
  time: number
  value: number
}

const COLORS = [
  "#55a76a",
  "#4682fa",
  "#ffaf0f",
  "#e55765",
  "#8b5cf6",
  "#0f8b8d",
  "#ee7b18",
  "#d94b8c",
]

const BROKER_LABEL = metricsLabel("ssoc-broker")
const WINDOW_SECONDS = 6 * 60 * 60
const STEP_SECONDS = 120

export function brokerMetricWindow(): MonitorWindow {
  const end = Math.floor(Date.now() / 1000)
  return { start: end - WINDOW_SECONDS, end, step: STEP_SECONDS }
}

export async function fetchBrokerMetrics(
  window: MonitorWindow
): Promise<Record<string, BrokerMetricData>> {
  const step = `${window.step}s`
  const label = BROKER_LABEL

  const [cpuRes, memUsedRes, memTotalRes, netRecvRes, netSentRes] =
    await Promise.all([
      rangeQuery(
        `max by (instance) (psutil_cpu_usage_percent{${label}})`,
        window.start,
        window.end,
        step
      ),
      rangeQuery(
        `max by (instance) (psutil_memory_bytes_used{${label}})`,
        window.start,
        window.end,
        step
      ),
      rangeQuery(
        `max by (instance) (psutil_memory_bytes_total{${label}})`,
        window.start,
        window.end,
        step
      ),
      rangeQuery(
        `sum by (instance) (rate(psutil_network_receive_bytes_total{${label}}[5m]))`,
        window.start,
        window.end,
        step
      ),
      rangeQuery(
        `sum by (instance) (rate(psutil_network_transmit_bytes_total{${label}}[5m]))`,
        window.start,
        window.end,
        step
      ),
    ])

  const cpu = parseSeries(cpuRes.data?.result ?? [])
  const memUsed = parseSeries(memUsedRes.data?.result ?? [])
  const memTotal = parseSeries(memTotalRes.data?.result ?? [])
  const netRecv = parseSeries(netRecvRes.data?.result ?? [])
  const netSent = parseSeries(netSentRes.data?.result ?? [])

  const slugs = new Set<string>()
  for (const map of [cpu, memUsed, memTotal, netRecv, netSent]) {
    for (const slug of map.keys()) slugs.add(slug)
  }

  const result: Record<string, BrokerMetricData> = {}
  for (const slug of slugs) {
    result[slug] = {
      cpu: [
        {
          name: "CPU",
          color: COLORS[0],
          samples: cpu.get(slug) ?? [],
        },
      ],
      mem: [
        {
          name: "已用",
          color: COLORS[0],
          samples: memUsed.get(slug) ?? [],
        },
        {
          name: "总量",
          color: COLORS[1],
          samples: memTotal.get(slug) ?? [],
        },
      ],
      net: [
        {
          name: "接收",
          color: COLORS[0],
          samples: netRecv.get(slug) ?? [],
        },
        {
          name: "发送",
          color: COLORS[1],
          samples: netSent.get(slug) ?? [],
        },
      ],
    }
  }

  return result
}

function parseSeries(points: VMPoint[]): Map<string, MetricSample[]> {
  const map = new Map<string, MetricSample[]>()

  for (const point of points) {
    const instance = point.metric.instance
    if (!instance) continue
    const samples = point.values
      .map(([time, value]) => ({
        time: time * 1000,
        value: Number(value) || 0,
      }))
      .sort((a, b) => a.time - b.time)
    map.set(instance, samples)
  }

  return map
}