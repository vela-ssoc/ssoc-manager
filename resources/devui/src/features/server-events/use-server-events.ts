import { useEffect, useRef } from "react"

import {
  onServerEvent,
  startServerEvents,
  stopServerEvents,
  type ServerEvent,
} from "./client"
import { appendServerEvent } from "./history"
import { useServerEventsEnabled } from "./preferences"

export function useServerEvents() {
  const enabled = useServerEventsEnabled()

  useEffect(() => {
    if (!enabled) {
      stopServerEvents()
      return
    }
    const off = onServerEvent((event) => appendServerEvent(event))
    startServerEvents()
    return () => {
      off()
      stopServerEvents()
    }
  }, [enabled])
}

export function useServerEvent(
  events: string | string[],
  handler: (event: ServerEvent) => void
) {
  const key = Array.isArray(events) ? events.join(",") : events
  const handlerRef = useRef(handler)

  useEffect(() => {
    handlerRef.current = handler
  }, [handler])

  useEffect(() => {
    const names = new Set(key.split(","))
    return onServerEvent((event) => {
      if (names.has(event.event)) handlerRef.current(event)
    })
  }, [key])
}