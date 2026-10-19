import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useState,
  type ReactNode,
} from "react"

import {
  applyTheme,
  getStoredAccent,
  getStoredTheme,
  setAccent as persistAccent,
  setTheme as persistTheme,
  watchSystemTheme,
  type AccentColor,
  type Theme,
} from "@/lib/theme"
import {
  getStoredFont,
  persistFont,
  type FontFamily,
} from "@/lib/fonts"

interface ThemeContextValue {
  theme: Theme
  accent: AccentColor
  isDark: boolean
  font: FontFamily
  setTheme: (theme: Theme) => void
  setAccent: (accent: AccentColor) => void
  setFont: (font: FontFamily) => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => getStoredTheme())
  const [accent, setAccentState] = useState<AccentColor>(() => getStoredAccent())
  const [font, setFontState] = useState<FontFamily>(() => getStoredFont())
  const [isDark, setIsDark] = useState(() =>
    document.documentElement.classList.contains("dark")
  )

  useLayoutEffect(() => {
    applyTheme(theme, accent)
    setIsDark(
      theme === "dark" ||
        (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
    )
  }, [theme, accent])

  useEffect(() => {
    if (theme !== "system") return
    return watchSystemTheme(() => {
      applyTheme("system", accent)
      setIsDark(window.matchMedia("(prefers-color-scheme: dark)").matches)
    })
  }, [theme, accent])

  const setTheme = useCallback((next: Theme) => {
    persistTheme(next)
    setThemeState(next)
  }, [])

  const setAccent = useCallback((next: AccentColor) => {
    persistAccent(next)
    setAccentState(next)
  }, [])

  const setFont = useCallback((next: FontFamily) => {
    persistFont(next)
    setFontState(next)
  }, [])

  return (
    <ThemeContext.Provider value={{
      theme,
      accent,
      isDark,
      font,
      setTheme,
      setAccent,
      setFont,
    }}>
      {children}
    </ThemeContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useTheme() {
  const context = useContext(ThemeContext)
  if (!context) throw new Error("useTheme must be used within ThemeProvider")
  return context
}
