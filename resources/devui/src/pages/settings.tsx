import { Monitor, Moon, Sun } from "lucide-react"

import { cn } from "@/lib/utils"
import { ACCENT_OPTIONS, type Theme } from "@/lib/theme"
import { useTheme } from "@/providers/theme-provider"
import { FONT_OPTIONS, getFontOption, type FontFamily } from "@/lib/fonts"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { TooltipHint } from "@/components/tooltip-hint"

const THEME_OPTIONS: Array<{
  value: Theme
  label: string
  icon: typeof Monitor
}> = [
  { value: "system", label: "系统", icon: Monitor },
  { value: "light", label: "亮色", icon: Sun },
  { value: "dark", label: "暗色", icon: Moon },
]

export function SettingsPage() {
  const { theme, setTheme, accent, setAccent, font, setFont } = useTheme()

  return (
    <div className="flex h-full flex-col gap-5 p-4">
      <section className="grid gap-3">
        <h2 className="text-sm font-semibold">主题</h2>
        <div className="flex gap-1">
          {THEME_OPTIONS.map(({ value, label, icon: Icon }) => (
            <TooltipHint key={value} content={label}>
              <button
                type="button"
                aria-label={label}
                aria-pressed={theme === value}
                className={cn(
                  "flex size-9 items-center justify-center rounded-md border border-transparent transition-colors hover:bg-accent/50",
                  theme === value && "bg-accent/40 text-primary"
                )}
                onClick={() => setTheme(value)}
              >
                <Icon className="size-4" />
              </button>
            </TooltipHint>
          ))}
        </div>
      </section>

      <section className="grid gap-3">
        <h2 className="text-sm font-semibold">强调色</h2>
        <div className="flex flex-wrap gap-2">
          {ACCENT_OPTIONS.map((option) => (
            <TooltipHint key={option.value} content={option.label}>
              <button
                type="button"
                aria-label={option.label}
                aria-pressed={accent === option.value}
                className="flex size-8 items-center justify-center rounded-full transition-colors hover:bg-accent/50"
                onClick={() => setAccent(option.value)}
              >
                <span
                  className="flex size-5 items-center justify-center rounded-full border-2"
                  style={{ borderColor: option.swatch }}
                >
                  {accent === option.value && (
                    <span
                      className="size-2 rounded-full"
                      style={{ backgroundColor: option.swatch }}
                    />
                  )}
                </span>
              </button>
            </TooltipHint>
          ))}
        </div>
      </section>

      <section className="grid gap-3">
        <h2 className="text-sm font-semibold">字体</h2>
        <Select value={font} onValueChange={(value) => setFont(value as FontFamily)}>
          <SelectTrigger
            className="w-52"
            style={{ fontFamily: getFontOption(font).monospace }}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {FONT_OPTIONS.map((option) => (
              <SelectItem
                key={option.value}
                value={option.value}
                style={{ fontFamily: option.monospace }}
              >
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </section>
    </div>
  )
}
