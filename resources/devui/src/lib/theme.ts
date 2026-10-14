export type Theme = "light" | "dark" | "system"

const STORAGE_KEY = "theme"
const MEDIA_QUERY = "(prefers-color-scheme: dark)"

export function getStoredTheme(): Theme {
  try {
    const value = localStorage.getItem(STORAGE_KEY)
    if (value === "light" || value === "dark" || value === "system") return value
  } catch {
    // ignore
  }
  return "system"
}

function systemPrefersDark(): boolean {
  return window.matchMedia(MEDIA_QUERY).matches
}

export function applyTheme(theme: Theme) {
  const dark = theme === "dark" || (theme === "system" && systemPrefersDark())
  const root = document.documentElement
  root.classList.toggle("dark", dark)
  root.style.colorScheme = dark ? "dark" : "light"
}

export function setTheme(theme: Theme) {
  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // ignore
  }
  applyTheme(theme)
}

// 监听系统主题变化，用于 "system" 模式。
export function watchSystemTheme(callback: () => void): () => void {
  const media = window.matchMedia(MEDIA_QUERY)
  media.addEventListener("change", callback)
  return () => media.removeEventListener("change", callback)
}