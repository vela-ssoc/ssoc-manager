import { useSyncExternalStore } from "react"
import { Sparkles, type LucideIcon } from "lucide-react"

export interface ExperimentalFeature {
  id: string
  name: string
  href: string
  icon: LucideIcon
}

export const EXPERIMENTAL_FEATURES: ExperimentalFeature[] = [
  {
    id: "ai-chat",
    name: "AI 助手",
    href: "/ai-chat",
    icon: Sparkles,
  },
]

const STORAGE_PREFIX = "devui.experimental."
const listeners = new Set<() => void>()
let version = 0

function storageKey(id: string) {
  return `${STORAGE_PREFIX}${id}`
}

export function isExperimentalEnabled(id: string) {
  try {
    return localStorage.getItem(storageKey(id)) === "1"
  } catch {
    return false
  }
}

export function setExperimentalEnabled(id: string, enabled: boolean) {
  try {
    localStorage.setItem(storageKey(id), enabled ? "1" : "0")
  } catch {
    return
  }
  version += 1
  for (const listener of [...listeners]) listener()
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function getSnapshot() {
  return version
}

export function useExperimentalFeatures() {
  useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
  return EXPERIMENTAL_FEATURES.map((feature) => ({
    ...feature,
    enabled: isExperimentalEnabled(feature.id),
  }))
}