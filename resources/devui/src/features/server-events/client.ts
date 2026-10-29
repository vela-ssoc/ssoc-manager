import { getToken } from "@/lib/http"

export interface ServerEvent<T = unknown> {
  event: string
  data: T
}

export interface BrokerEventData {
  slug: string
  name: string
  inet: string
  version: string
  text?: string
  connected_at?: string
  disconnected_at?: string
}

export interface ServerEventMap {
  "broker.online": BrokerEventData
  "broker.offline": BrokerEventData
  "broker.conflict": BrokerEventData
}

type ServerEventListener = (event: ServerEvent) => void

const SSE_PATH = "/api/v3/sse/event"
const MIN_RETRY_DELAY = 1000
const MAX_RETRY_DELAY = 30000

const listeners = new Set<ServerEventListener>()

let controller: AbortController | null = null
let retryTimer: number | null = null
let retryDelay = MIN_RETRY_DELAY
let started = false

export function onServerEvent(listener: ServerEventListener) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function startServerEvents() {
  if (started) return
  started = true
  void connect()
}

export function stopServerEvents() {
  started = false
  if (retryTimer !== null) {
    window.clearTimeout(retryTimer)
    retryTimer = null
  }
  retryDelay = MIN_RETRY_DELAY
  controller?.abort()
  controller = null
}

async function connect() {
  if (!started || controller) return
  const token = getToken()
  if (!token) return

  const ctrl = new AbortController()
  controller = ctrl
  let retry = true
  try {
    const response = await fetch(SSE_PATH, {
      headers: {
        Accept: "text/event-stream",
        "Cache-Control": "no-cache",
        Authorization: token,
      },
      signal: ctrl.signal,
    })
    if (response.status >= 400 && response.status < 500) {
      retry = false
      return
    }
    if (!response.ok || !response.body) {
      throw new Error(`事件流连接失败（${response.status}）`)
    }
    retryDelay = MIN_RETRY_DELAY
    await readStream(response.body)
  } catch {
    retry = started && !ctrl.signal.aborted
  } finally {
    if (controller === ctrl) controller = null
  }

  if (retry) scheduleRetry()
}

function scheduleRetry() {
  if (!started || retryTimer !== null) return
  const delay = retryDelay
  retryDelay = Math.min(retryDelay * 2, MAX_RETRY_DELAY)
  retryTimer = window.setTimeout(() => {
    retryTimer = null
    void connect()
  }, delay)
}

async function readStream(body: ReadableStream<Uint8Array>) {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""
  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let index = buffer.indexOf("\n\n")
    while (index >= 0) {
      dispatch(buffer.slice(0, index))
      buffer = buffer.slice(index + 2)
      index = buffer.indexOf("\n\n")
    }
  }
}

function dispatch(chunk: string) {
  let event = ""
  const data: string[] = []
  for (const raw of chunk.split("\n")) {
    const line = raw.replace(/\r$/, "")
    if (!line || line.startsWith(":")) continue
    const colon = line.indexOf(":")
    const field = colon < 0 ? line : line.slice(0, colon)
    let value = colon < 0 ? "" : line.slice(colon + 1)
    if (value.startsWith(" ")) value = value.slice(1)
    if (field === "event") event = value
    else if (field === "data") data.push(value)
  }
  if (!event) return

  const payload = data.join("\n")
  let parsed: unknown = payload
  if (payload) {
    try {
      parsed = JSON.parse(payload)
    } catch {
      parsed = payload
    }
  }

  const evt: ServerEvent = { event, data: parsed }
  for (const listener of [...listeners]) {
    listener(evt)
  }
}