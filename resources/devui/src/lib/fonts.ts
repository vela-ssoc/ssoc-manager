export const FONT_OPTIONS = [
  {
    value: "space-mono",
    label: "Space Mono",
    monospace: '"Space Mono"',
  },
  {
    value: "ibm-plex-mono",
    label: "IBM Plex Mono",
    monospace: '"IBM Plex Mono", "Space Mono"',
  },
  {
    value: "jetbrains-mono",
    label: "JetBrains Mono",
    monospace: '"JetBrains Mono", "Space Mono"',
  },
  {
    value: "maple-mono",
    label: "Maple Mono",
    monospace: '"Maple Mono", "Space Mono"',
  },
  {
    value: "ubuntu-mono",
    label: "Ubuntu Mono",
    monospace: '"Ubuntu Mono", "Space Mono"',
  },
  {
    value: "comic-mono",
    label: "Comic Mono",
    monospace: '"Comic Mono", "Space Mono"',
  },
  {
    value: "system-mono",
    label: "系统等宽",
    monospace:
      'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", "Courier New"',
  },
] as const

export type FontFamily = (typeof FONT_OPTIONS)[number]["value"]

const STORAGE_KEY = "font-family"
const DEFAULT_FONT: FontFamily = "space-mono"

export function getStoredFont(): FontFamily {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (FONT_OPTIONS.some((font) => font.value === stored)) {
      return stored as FontFamily
    }
  } catch {
    // Ignore unavailable local storage.
  }
  return DEFAULT_FONT
}

export function getFontOption(font: FontFamily) {
  return FONT_OPTIONS.find((item) => item.value === font) ?? FONT_OPTIONS[0]
}

export function getMonacoFontFamily(font: FontFamily): string {
  return `${getFontOption(font).monospace}, "NSimSun", "Microsoft YaHei", "PingFang SC", sans-serif`
}

export function persistFont(font: FontFamily) {
  try {
    localStorage.setItem(STORAGE_KEY, font)
  } catch {
    // Ignore unavailable local storage.
  }
}
