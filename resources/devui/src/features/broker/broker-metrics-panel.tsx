import { Cpu, MemoryStick, Network, TriangleAlert } from "lucide-react"

import { MonitorChart, type MonitorWindow } from "@/features/central/monitoring"

import type { BrokerMetricData } from "./broker-metrics"

export function BrokerMetricsPanel({
  data,
  window,
  loading,
  error,
}: {
  data?: BrokerMetricData
  window: MonitorWindow
  loading: boolean
  error: boolean
}) {
  if (loading) {
    return (
      <div className="grid gap-2">
        {[0, 1, 2].map((index) => (
          <div
            key={index}
            className="h-[4.75rem] animate-pulse rounded-md border bg-muted/30"
          />
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-muted-foreground flex min-h-20 flex-col items-center justify-center gap-2 rounded-md border border-dashed bg-card text-xs">
        <TriangleAlert className="size-4 opacity-60" />
        指标加载失败
      </div>
    )
  }

  return (
    <div className="grid gap-2">
      <MonitorChart
        icon={Cpu}
        title="CPU 使用率"
        unit="percent"
        series={data?.cpu ?? []}
        window={window}
        mini
      />
      <MonitorChart
        icon={MemoryStick}
        title="内存"
        unit="bytes"
        series={data?.mem ?? []}
        window={window}
        mini
      />
      <MonitorChart
        icon={Network}
        title="网络收发"
        unit="rate"
        series={data?.net ?? []}
        window={window}
        mini
      />
    </div>
  )
}