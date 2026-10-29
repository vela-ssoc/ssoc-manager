import { getAuthProfile } from "@/lib/auth-session"

const LEGACY_KEYS = ["event-history", "event-subscription"]

if (typeof window !== "undefined") {
  for (const key of LEGACY_KEYS) {
    try {
      window.localStorage.removeItem(key)
    } catch {
      // ignore
    }
  }
}

function scope(): string {
  const profile = getAuthProfile()
  if (profile && Number.isFinite(profile.id)) {
    return String(profile.id)
  }
  return "anonymous"
}

export function scopedKey(prefix: string): string {
  return `${prefix}:${scope()}`
}