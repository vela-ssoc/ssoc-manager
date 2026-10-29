import { useSyncExternalStore } from "react"

import { AUTH_PROFILE_KEY } from "@/lib/auth-session"

import { scopedKey } from "./scope"

const STORAGE_PREFIX = "event-subscription"

const listeners = new Set<() => void>()
let currentKey = ""
let enabled: boolean | null = null

function storageKey() {
  return scopedKey(STORAGE_PREFIX)
}

function syncScope() {
  const key = storageKey()
  if (key !== currentKey) {
    currentKey = key
    enabled = null
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
      enabled = null
      notify()
      return
    }
    if (event.key !== storageKey()) return
    enabled = null
    notify()
  })
}

function readEnabled(): boolean {
  const key = syncScope()
  if (enabled !== null) return enabled
  try {
    const raw = window.localStorage.getItem(key)
    enabled = raw === null ? true : raw !== "0"
  } catch {
    enabled = true
  }
  return enabled
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function useServerEventsEnabled() {
  return useSyncExternalStore(subscribe, readEnabled, readEnabled)
}

export function setServerEventsEnabled(value: boolean) {
  enabled = value
  try {
    window.localStorage.setItem(storageKey(), value ? "1" : "0")
  } catch {
    enabled = value
  }
  notify()
}