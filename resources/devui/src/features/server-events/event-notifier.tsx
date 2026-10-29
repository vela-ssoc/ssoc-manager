import { useCallback, useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { ChevronRight, X } from "lucide-react"

import { cn } from "@/lib/utils"

import { onServerEvent, type ServerEvent } from "./client"
import {
  eventToneStyles,
  renderServerEvent,
  type EventPresentation,
} from "./event-renderers"

interface EventNotification {
  id: number
  presentation: EventPresentation
}

const AUTO_CLOSE_DELAY = 6000
const HOVER_CLOSE_DELAY = 2000
const MAX_ITEMS = 3

export function EventNotifications() {
  const [items, setItems] = useState<EventNotification[]>([])
  const seq = useRef(0)

  useEffect(() => {
    return onServerEvent((event: ServerEvent) => {
      const presentation = renderServerEvent(event.event, event.data)
      const item: EventNotification = { id: seq.current, presentation }
      seq.current += 1
      setItems((prev) =>
        [
          ...prev.filter(
            (it) => !presentation.key || it.presentation.key !== presentation.key
          ),
          item,
        ].slice(-MAX_ITEMS)
      )
    })
  }, [])

  const remove = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id))
  }, [])

  return (
    <div
      role="status"
      aria-live="polite"
      className="pointer-events-none fixed top-4 right-4 z-[60] flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2"
    >
      {items.map((item) => (
        <EventNotificationCard key={item.id} item={item} onClose={remove} />
      ))}
    </div>
  )
}

function EventNotificationCard({
  item,
  onClose,
}: {
  item: EventNotification
  onClose: (id: number) => void
}) {
  const timer = useRef<number | null>(null)
  const presentation = item.presentation
  const tone = eventToneStyles[presentation.tone]
  const Icon = presentation.icon

  useEffect(() => {
    timer.current = window.setTimeout(
      () => onClose(item.id),
      presentation.duration ?? AUTO_CLOSE_DELAY
    )
    return () => {
      if (timer.current !== null) window.clearTimeout(timer.current)
    }
  }, [item.id, onClose, presentation.duration])

  const pause = () => {
    if (timer.current !== null) {
      window.clearTimeout(timer.current)
      timer.current = null
    }
  }

  const resume = () => {
    pause()
    timer.current = window.setTimeout(() => onClose(item.id), HOVER_CLOSE_DELAY)
  }

  return (
    <div
      onMouseEnter={pause}
      onMouseLeave={resume}
      className={cn(
        "event-toast-enter pointer-events-auto relative flex items-start gap-3 overflow-hidden rounded-md border bg-card p-3 pr-8 shadow-lg",
        tone.card
      )}
    >
      <span
        className={cn("absolute inset-y-0 left-0 w-0.5", tone.bar)}
      />
      <span
        className={cn(
          "flex size-8 shrink-0 items-center justify-center rounded-md",
          tone.icon
        )}
      >
        <Icon className="size-4" />
      </span>
      <div className="min-w-0 flex-1">
        <div
          className={cn(
            "flex min-w-0 items-center gap-1.5 text-[11px] font-medium",
            tone.label
          )}
        >
          <span className={cn("size-1.5 shrink-0 rounded-full", tone.dot)} />
          <span className="truncate">{presentation.title}</span>
        </div>
        {presentation.content && (
          <div className="mt-0.5">{presentation.content}</div>
        )}
        {presentation.href && (
          <Link
            to={presentation.href}
            className="group/link text-muted-foreground hover:text-foreground mt-1 inline-flex items-center gap-0.5 text-[11px] transition-colors"
            onClick={() => onClose(item.id)}
          >
            查看
            <ChevronRight className="size-3 transition-transform group-hover/link:translate-x-0.5" />
          </Link>
        )}
        {presentation.action && (
          <div className="mt-2">{presentation.action}</div>
        )}
      </div>
      <button
        type="button"
        aria-label="关闭通知"
        className="text-muted-foreground hover:bg-accent hover:text-foreground absolute top-2 right-2 flex size-5 items-center justify-center rounded"
        onClick={() => onClose(item.id)}
      >
        <X className="size-3.5" />
      </button>
    </div>
  )
}