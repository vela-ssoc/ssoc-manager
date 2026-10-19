export type Theme = "light" | "dark" | "system"
export type AccentColor =
  | "green"
  | "teal"
  | "blue"
  | "indigo"
  | "purple"
  | "yellow"
  | "orange"
  | "red"
  | "pink"

export const ACCENT_OPTIONS: Array<{
  value: AccentColor
  label: string
  swatch: string
}> = [
  { value: "green", label: "绿色", swatch: "#238636" },
  { value: "teal", label: "青色", swatch: "#0f8b8d" },
  { value: "blue", label: "蓝色", swatch: "#3478d4" },
  { value: "indigo", label: "靛蓝色", swatch: "#5856d6" },
  { value: "purple", label: "紫色", swatch: "#8b5cf6" },
  { value: "yellow", label: "黄色", swatch: "#b77908" },
  { value: "orange", label: "橙色", swatch: "#ee7b18" },
  { value: "red", label: "红色", swatch: "#dc3e43" },
  { value: "pink", label: "粉色", swatch: "#d94b8c" },
]

const STORAGE_THEME_KEY = "theme"
const STORAGE_ACCENT_KEY = "accent-color"
const MEDIA_QUERY = "(prefers-color-scheme: dark)"
const DEFAULT_ACCENT: AccentColor = "orange"

const ACCENT_HUES: Record<AccentColor, number> = {
  green: 142,
  teal: 176,
  blue: 216,
  indigo: 239,
  purple: 271,
  yellow: 42,
  orange: 25,
  red: 0,
  pink: 330,
}

export function getStoredTheme(): Theme {
  try {
    const value = localStorage.getItem(STORAGE_THEME_KEY)
    if (value === "light" || value === "dark" || value === "system") return value
  } catch {
    // Ignore unavailable local storage.
  }
  return "system"
}

export function getStoredAccent(): AccentColor {
  try {
    const value = localStorage.getItem(STORAGE_ACCENT_KEY)
    if (ACCENT_OPTIONS.some((option) => option.value === value)) {
      return value as AccentColor
    }
  } catch {
    // Ignore unavailable local storage.
  }
  return DEFAULT_ACCENT
}

function systemPrefersDark(): boolean {
  return window.matchMedia(MEDIA_QUERY).matches
}

export function applyTheme(theme: Theme, accent = getStoredAccent()) {
  const dark = theme === "dark" || (theme === "system" && systemPrefersDark())
  const hue = ACCENT_HUES[accent]
  const root = document.documentElement
  root.classList.toggle("dark", dark)
  root.style.colorScheme = dark ? "dark" : "light"

  const primary = dark
    ? accent === "yellow"
      ? `hsl(${hue} 76% 55%)`
      : `hsl(${hue} 68% 48%)`
    : accent === "yellow"
      ? `hsl(${hue} 90% 42%)`
      : `hsl(${hue} 78% 48%)`
  const primaryForeground = accent === "yellow" ? "hsl(24 55% 10%)" : "hsl(0 0% 100%)"
  root.style.setProperty(
    "--toast-bg",
    dark ? `hsl(${hue} 22% 12%)` : `hsl(${hue} 70% 98%)`
  )
  root.style.setProperty("--primary", primary)
  root.style.setProperty("--primary-foreground", primaryForeground)
  root.style.setProperty("--destructive", primary)
  root.style.setProperty("--destructive-foreground", primaryForeground)
  root.style.setProperty("--ring", primary)
  root.style.setProperty("--sidebar-primary", primary)
  root.style.setProperty("--sidebar-primary-foreground", primaryForeground)
  root.style.setProperty("--sidebar-ring", primary)

  if (dark) {
    root.style.setProperty("--accent", "hsl(0 0% 15%)")
    root.style.setProperty("--accent-foreground", "hsl(0 0% 98%)")
    root.style.setProperty("--sidebar-accent", "hsl(0 0% 15%)")
    root.style.setProperty("--sidebar-accent-foreground", "hsl(0 0% 98%)")
  } else {
    root.style.setProperty("--accent", `hsl(${hue} 95% 96%)`)
    root.style.setProperty("--accent-foreground", "hsl(240 5.9% 10%)")
    root.style.setProperty("--sidebar-accent", `hsl(${hue} 95% 96%)`)
    root.style.setProperty("--sidebar-accent-foreground", "hsl(240 5.9% 10%)")
  }
}

export function setTheme(theme: Theme) {
  try {
    localStorage.setItem(STORAGE_THEME_KEY, theme)
  } catch {
    // Ignore unavailable local storage.
  }
  applyTheme(theme)
}

export function setAccent(accent: AccentColor) {
  try {
    localStorage.setItem(STORAGE_ACCENT_KEY, accent)
  } catch {
    // Ignore unavailable local storage.
  }
  applyTheme(getStoredTheme(), accent)
}

export function watchSystemTheme(callback: () => void): () => void {
  const media = window.matchMedia(MEDIA_QUERY)
  media.addEventListener("change", callback)
  return () => media.removeEventListener("change", callback)
}
