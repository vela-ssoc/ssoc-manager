import { useSyncExternalStore } from "react"

import { AUTH_PROFILE_KEY } from "@/lib/auth-session"

import { scopedKey } from "./scope"

export interface ServerEventRecord {
  id: string
  event: string
  data: unknown
  receivedAt: string
  read: boolean
}

const STORAGE_PREFIX = "event-history"
const MAX_RECORDS = 200

const listeners = new Set<() => void>()
let currentKey = ""
let records: ServerEventRecord[] | null = null

function storageKey() {
  return scopedKey(STORAGE_PREFIX)
}

function syncScope() {
  const key = storageKey()
  if (key !== currentKey) {
    currentKey = key
    records = null
  }
  return key
}

function notify() {
  for (const listener of [...listeners]) listener()
}

if (typeof window !== "undefined") {
  window.addEventListener("storage", (event) => {
    if (event.key === AUTH_PROFILE_KEY) {
      currentKey = ""
      records = null
      notify()
      return
    }
    if (event.key !== storageKey()) return
    records = null
    notify()
  })
}

function isRecord(value: unknown): value is ServerEventRecord {
  if (!value || typeof value !== "object") return false
  const record = value as Partial<ServerEventRecord>
  return (
    typeof record.id === "string" &&
    typeof record.event === "string" &&
    typeof record.receivedAt === "string" &&
    typeof record.read === "boolean"
  )
}

function snapshot(): ServerEventRecord[] {
  const key = syncScope()
  if (records) return records
  try {
    const raw = window.localStorage.getItem(key)
    const parsed: unknown = raw ? JSON.parse(raw) : []
    records = Array.isArray(parsed)
      ? parsed.filter(isRecord).slice(0, MAX_RECORDS)
      : []
  } catch {
    records = []
  }
  return records
}

function commit(next: ServerEventRecord[]) {
  records = next
  const key = storageKey()
  let persisted = next
  for (;;) {
    try {
      window.localStorage.setItem(key, JSON.stringify(persisted))
      break
    } catch {
      if (persisted.length === 0) break
      persisted = persisted.slice(0, Math.floor(persisted.length / 2))
    }
  }
  notify()
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function createId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

export function useServerEventHistory() {
  return useSyncExternalStore(subscribe, snapshot, snapshot)
}

export function appendServerEvent(event: { event: string; data: unknown }) {
  const record: ServerEventRecord = {
    id: createId(),
    event: event.event,
    data: event.data,
    receivedAt: new Date().toISOString(),
    read: false,
  }
  commit([record, ...snapshot()].slice(0, MAX_RECORDS))
}

export function markServerEventRead(id: string) {
  commit(
    snapshot().map((record) =>
      record.id === id && !record.read ? { ...record, read: true } : record
    )
  )
}

export function toggleServerEventRead(id: string) {
  commit(
    snapshot().map((record) =>
      record.id === id ? { ...record, read: !record.read } : record
    )
  )
}

export function removeServerEvent(id: string) {
  commit(snapshot().filter((record) => record.id !== id))
}

export function markAllServerEventsRead() {
  commit(snapshot().map((record) => (record.read ? record : { ...record, read: true })))
}

export function clearServerEventHistory() {
  commit([])
}