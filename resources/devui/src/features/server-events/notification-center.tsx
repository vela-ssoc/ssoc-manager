import { useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import {
  Bell,
  BellOff,
  Check,
  CheckCheck,
  ChevronRight,
  Circle,
  Inbox,
  Trash2,
} from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Switch } from "@/components/ui/switch"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"

import { eventToneStyles, renderServerEvent } from "./event-renderers"
import {
  markAllServerEventsRead,
  markServerEventRead,
  removeServerEvent,
  toggleServerEventRead,
  useServerEventHistory,
  type ServerEventRecord,
} from "./history"
import {
  setServerEventsEnabled,
  useServerEventsEnabled,
} from "./preferences"

export function NotificationCenter() {
  const records = useServerEventHistory()
  const enabled = useServerEventsEnabled()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const unread = records.reduce((sum, record) => sum + (record.read ? 0 : 1), 0)

  useEffect(() => {
    if (!open) return
    const onPointerDown = (event: PointerEvent) => {
      const target = event.target as Node | null
      if (containerRef.current && target && !containerRef.current.contains(target)) {
        setOpen(false)
      }
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false)
    }
    document.addEventListener("pointerdown", onPointerDown)
    document.addEventListener("keydown", onKeyDown)
    return () => {
      document.removeEventListener("pointerdown", onPointerDown)
      document.removeEventListener("keydown", onKeyDown)
    }
  }, [open])

  return (
    <div ref={containerRef} className="relative shrink-0">
      <TooltipHint content={enabled ? "通知" : "通知（已静音）"}>
        <button
          type="button"
          aria-label={
            !enabled
              ? "通知（已静音）"
              : unread > 0
                ? `通知（${unread} 条未读）`
                : "通知"
          }
          aria-expanded={open}
          className={cn(
            "relative flex size-7 items-center justify-center rounded-md transition-colors",
            open
              ? "bg-sidebar-accent text-sidebar-accent-foreground"
              : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
          )}
          onClick={() => setOpen((visible) => !visible)}
        >
          {enabled ? <Bell className="size-4" /> : <BellOff className="size-4" />}
          {unread > 0 && (
            <span className="bg-primary text-primary-foreground ring-sidebar absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[9px] leading-none font-medium ring-2">
              {unread > 99 ? "99+" : unread}
            </span>
          )}
        </button>
      </TooltipHint>

      {open && (
        <div className="bg-popover text-popover-foreground absolute top-full left-[calc(100%+0.5rem)] z-50 mt-2 flex max-h-[70vh] w-96 max-w-[calc(100vw-15rem)] flex-col overflow-hidden rounded-md border shadow-lg">
          <div className="flex items-center justify-between gap-2 border-b px-3 py-2">
            <div className="flex min-w-0 items-center gap-1.5">
              <span className="text-sm font-semibold">通知</span>
              {unread > 0 && (
                <span className="bg-primary/10 text-primary rounded-full px-1.5 text-[10px] leading-4 font-medium">
                  {unread} 条未读
                </span>
              )}
            </div>
            <button
              type="button"
              disabled={unread === 0}
              className="text-muted-foreground hover:text-foreground flex shrink-0 items-center gap-1 rounded px-1.5 py-1 text-[11px] transition-colors disabled:opacity-40"
              onClick={markAllServerEventsRead}
            >
              <CheckCheck className="size-3.5" />
              全部已读
            </button>
          </div>

          {records.length === 0 ? (
            <div className="text-muted-foreground flex flex-col items-center gap-2 px-4 py-10 text-xs">
              <Inbox className="size-6 opacity-60" />
              暂无通知
            </div>
          ) : (
            <div className="min-h-0 flex-1 divide-y overflow-y-auto">
              {records.map((record) => (
                <NotificationRow
                  key={record.id}
                  record={record}
                  onNavigate={() => setOpen(false)}
                />
              ))}
            </div>
          )}

          <div className="border-t px-3 py-2">
            <div className="flex items-center justify-between gap-2">
              <span className="text-muted-foreground text-[11px]">订阅通知</span>
              <Switch
                checked={enabled}
                label="订阅通知"
                onCheckedChange={setServerEventsEnabled}
              />
            </div>
            <p className="text-muted-foreground mt-1 text-[10px]">
              {enabled
                ? "关闭后不再连接事件流，最多保留最近 200 条通知"
                : "已关闭事件流，不再接收新通知；历史记录仍保留"}
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

function NotificationRow({
  record,
  onNavigate,
}: {
  record: ServerEventRecord
  onNavigate: () => void
}) {
  const presentation = renderServerEvent(record.event, record.data)
  const tone = eventToneStyles[presentation.tone]
  const Icon = presentation.icon

  return (
    <div className="group relative">
      <button
        type="button"
        className={cn(
          "hover:bg-muted/40 flex w-full items-start gap-2.5 px-3 py-2 pr-28 text-left transition-colors",
          !record.read && "bg-primary/5"
        )}
        onClick={() => markServerEventRead(record.id)}
      >
        <span
          className={cn(
            "mt-0.5 flex size-6 shrink-0 items-center justify-center rounded",
            tone.icon
          )}
        >
          <Icon className="size-3.5" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="flex min-w-0 items-center gap-1.5">
            {!record.read && (
              <span className="bg-primary size-1.5 shrink-0 rounded-full" />
            )}
            <span
              className={cn(
                "min-w-0 truncate text-xs",
                record.read ? "text-muted-foreground" : "font-medium"
              )}
            >
              {presentation.title}
            </span>
          </span>
          {presentation.content && (
            <span
              className={cn("mt-0.5 block", record.read && "opacity-70")}
            >
              {presentation.content}
            </span>
          )}
          <span className="text-muted-foreground mt-1 block text-[10px] tabular-nums">
            {formatDateTime(record.receivedAt)}
          </span>
        </span>
      </button>
      <div className="absolute top-1/2 right-2 flex -translate-y-1/2 items-center gap-1">
        {presentation.href && (
          <Link
            to={presentation.href}
            className="group/link text-muted-foreground hover:bg-muted hover:text-foreground flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[10px] transition-colors"
            onClick={() => {
              markServerEventRead(record.id)
              onNavigate()
            }}
          >
            查看
            <ChevronRight className="size-3 transition-transform group-hover/link:translate-x-0.5" />
          </Link>
        )}
        <TooltipHint content={record.read ? "标记未读" : "标记已读"}>
          <button
            type="button"
            aria-label={record.read ? "标记未读" : "标记已读"}
            className="text-muted-foreground hover:bg-accent hover:text-foreground flex size-5 items-center justify-center rounded opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
            onClick={() => toggleServerEventRead(record.id)}
          >
            {record.read ? <Circle className="size-3.5" /> : <Check className="size-3.5" />}
          </button>
        </TooltipHint>
        <TooltipHint content="删除通知">
          <button
            type="button"
            aria-label="删除通知"
            className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive flex size-5 items-center justify-center rounded opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
            onClick={() => removeServerEvent(record.id)}
          >
            <Trash2 className="size-3.5" />
          </button>
        </TooltipHint>
      </div>
    </div>
  )
}