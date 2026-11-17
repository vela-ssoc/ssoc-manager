import {
  createComponentImplementation,
  type ReactCatalogComponent,
} from "@a2ui/react"
import { Catalog } from "@a2ui/web_core"
import {
  basicCatalog,
  IconApi,
} from "@a2ui/web_core/catalogs/basic/v1"
import { CommonSchemas } from "@a2ui/web_core/v1_0"
import {
  CircleCheck,
  CircleHelp,
  CircleX,
  Info,
  Lock,
  Pause,
  Play,
  Trash2,
  TriangleAlert,
  X,
  type LucideIcon,
} from "lucide-react"
import { z } from "zod3"

import { cn } from "@/lib/utils"
import { ChartView } from "./chart-view"

export const CHAT_CATALOG_ID = "https://ssoc.local/a2ui/catalog/v1"

const ICON_MAP: Record<string, { icon: LucideIcon; className: string }> = {
  check: {
    icon: CircleCheck,
    className: "text-green-600 dark:text-green-400",
  },
  close: { icon: X, className: "text-muted-foreground" },
  delete: { icon: Trash2, className: "text-red-600 dark:text-red-400" },
  error: { icon: CircleX, className: "text-red-600 dark:text-red-400" },
  info: { icon: Info, className: "text-muted-foreground" },
  lock: { icon: Lock, className: "text-muted-foreground" },
  pause: { icon: Pause, className: "text-muted-foreground" },
  play: { icon: Play, className: "text-muted-foreground" },
  warning: {
    icon: TriangleAlert,
    className: "text-amber-600 dark:text-amber-400",
  },
}

const IconComponent = createComponentImplementation(IconApi, ({ props }) => {
  const name = typeof props.name === "string" ? props.name : ""
  const entry = ICON_MAP[name] ?? {
    icon: CircleHelp,
    className: "text-muted-foreground",
  }
  const Icon = entry.icon
  return <Icon className={cn("size-4 shrink-0", entry.className)} />
})

const chartApi = {
  name: "Chart",
  schema: z.object({
    chart_type: z.enum(["line", "bar", "pie", "metric_card"]),
    title: CommonSchemas.DynamicString,
    x_axis: CommonSchemas.DynamicStringList.optional(),
    series: z.array(
      z.object({
        name: CommonSchemas.DynamicString,
        data: z.array(CommonSchemas.DynamicNumber),
      })
    ),
    summary: CommonSchemas.DynamicString.optional(),
    weight: z.number().optional(),
  }),
}

const ChartComponent = createComponentImplementation(
  chartApi,
  ({ props }) => (
    <ChartView
      chartType={props.chart_type}
      title={props.title}
      xAxis={props.x_axis}
      series={props.series.map((item, index) => ({
        name: item.name ?? `系列 ${index + 1}`,
        data: item.data as number[],
      }))}
      summary={props.summary}
    />
  )
)

const baseComponents = Array.from(basicCatalog.components.values()).map(
  (component) => (component.name === "Icon" ? IconComponent : component)
)

export const chatCatalog = new Catalog<ReactCatalogComponent>(
  CHAT_CATALOG_ID,
  "1.0",
  [...baseComponents, ChartComponent],
  [...basicCatalog.functions.values()],
)