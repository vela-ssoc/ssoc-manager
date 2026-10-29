import type { ReactNode } from "react"
import {
  BellRing,
  RadioTower,
  TriangleAlert,
  Unplug,
  type LucideIcon,
} from "lucide-react"

import { formatDuration } from "@/lib/format"

import type { BrokerEventData, ServerEventMap } from "./client"

export type EventTone = "success" | "warning" | "danger" | "info" | "neutral"

export interface EventPresentation {
  tone: EventTone
  icon: LucideIcon
  title: string
  content?: ReactNode
  action?: ReactNode
  href?: string
  key?: string
  duration?: number
}

export const eventToneStyles: Record<
  EventTone,
  { card: string; bar: string; icon: string; label: string; dot: string }
> = {
  success: {
    card: "border-green-500/25",
    bar: "bg-green-500",
    icon: "bg-green-500/10 text-green-600 dark:text-green-400",
    label: "text-green-600 dark:text-green-400",
    dot: "bg-green-500",
  },
  warning: {
    card: "border-amber-500/25",
    bar: "bg-amber-500",
    icon: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    label: "text-amber-600 dark:text-amber-400",
    dot: "bg-amber-500",
  },
  danger: {
    card: "border-red-500/30",
    bar: "bg-red-500",
    icon: "bg-red-500/10 text-red-600 dark:text-red-400",
    label: "text-red-600 dark:text-red-400",
    dot: "bg-red-500",
  },
  info: {
    card: "border-primary/25",
    bar: "bg-primary",
    icon: "bg-primary/10 text-primary",
    label: "text-primary",
    dot: "bg-primary",
  },
  neutral: {
    card: "",
    bar: "bg-muted-foreground/50",
    icon: "bg-muted text-muted-foreground",
    label: "text-muted-foreground",
    dot: "bg-muted-foreground/60",
  },
}

type EventRenderer = (data: unknown) => EventPresentation

interface BrokerBase {
  slug: string
  name: string
  meta: string
  text: string
  duration: string
  href?: string
}

function brokerBase(data: unknown): BrokerBase {
  const info = (data ?? {}) as Partial<BrokerEventData>
  const slug = info.slug?.trim() ?? ""
  const name = info.name?.trim() || slug || "未知节点"
  const version = info.version?.trim() ?? ""
  const meta = [
    slug && slug !== name ? slug : "",
    info.inet?.trim() ?? "",
    version,
  ]
    .filter(Boolean)
    .join(" · ")

  const start = info.connected_at ? Date.parse(info.connected_at) : NaN
  const end = info.disconnected_at ? Date.parse(info.disconnected_at) : NaN
  const duration =
    Number.isFinite(start) && Number.isFinite(end) && end > start
      ? formatDuration((end - start) / 1000)
      : ""

  return {
    slug,
    name,
    meta,
    text: info.text?.trim() ?? "",
    duration,
    href: slug ? `/brokers/${encodeURIComponent(slug)}` : undefined,
  }
}

function brokerContent(meta: string, note?: string) {
  if (!meta && !note) return undefined

  return (
    <div className="min-w-0">
      {meta && (
        <div className="text-muted-foreground truncate font-mono text-[11px]">
          {meta}
        </div>
      )}
      {note && (
        <div className="text-muted-foreground mt-0.5 text-[11px]">{note}</div>
      )}
    </div>
  )
}

function brokerOnlineRenderer(data: unknown): EventPresentation {
  const base = brokerBase(data)

  return {
    tone: "success",
    icon: RadioTower,
    title: `[中继节点] ${base.name} 上线了`,
    key: base.slug || base.name,
    href: base.href,
    content: brokerContent(base.meta),
  }
}

function brokerOfflineRenderer(data: unknown): EventPresentation {
  const base = brokerBase(data)

  return {
    tone: "warning",
    icon: Unplug,
    title: `[中继节点] ${base.name} 下线了`,
    key: base.slug || base.name,
    href: base.href,
    content: brokerContent(
      base.meta,
      base.duration ? `在线时长 ${base.duration}` : undefined
    ),
  }
}

function brokerConflictRenderer(data: unknown): EventPresentation {
  const base = brokerBase(data)

  return {
    tone: "danger",
    icon: TriangleAlert,
    title: `[中继节点] ${base.name} 连接冲突`,
    key: base.slug || base.name,
    href: base.href,
    duration: 8000,
    content: brokerContent(base.meta, base.text || "节点已在其它位置登录"),
  }
}

const eventRenderers: Record<keyof ServerEventMap, EventRenderer> = {
  "broker.online": brokerOnlineRenderer,
  "broker.offline": brokerOfflineRenderer,
  "broker.conflict": brokerConflictRenderer,
}

function fallbackRenderer(event: string, data: unknown): EventPresentation {
  const text = typeof data === "string" ? data : JSON.stringify(data) ?? ""

  return {
    tone: "info",
    icon: BellRing,
    title: event,
    content: text ? (
      <div className="text-muted-foreground line-clamp-2 font-mono text-[11px] break-all">
        {text}
      </div>
    ) : undefined,
  }
}

export function renderServerEvent(
  event: string,
  data: unknown
): EventPresentation {
  if (event in eventRenderers) {
    const render = eventRenderers[event as keyof ServerEventMap]
    return render(data)
  }

  return fallbackRenderer(event, data)
}