import { Monitor, Moon, Sun } from "lucide-react"

import { cn } from "@/lib/utils"
import { ACCENT_OPTIONS, type Theme } from "@/lib/theme"
import { useTheme } from "@/providers/theme-provider"
import { FONT_OPTIONS, getFontOption, type FontFamily } from "@/lib/fonts"
import {
  setExperimentalEnabled,
  useExperimentalFeatures,
} from "@/lib/experimental-features"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { InfoTooltip } from "@/components/info-tooltip"
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
  const features = useExperimentalFeatures()

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

      <section className="grid gap-3">
        <div className="flex items-center gap-1.5">
          <h2 className="text-sm font-semibold">实验功能</h2>
          <InfoTooltip label="实验功能说明">
            <ul>
              <li>这些功能仍在开发中，可能尚未经过完整测试。</li>
              <li>可能存在功能不完善、数据不准确或异常 BUG。</li>
              <li>开启后入口显示在侧边栏「实验功能」中，可随时关闭。</li>
            </ul>
          </InfoTooltip>
        </div>
        <div className="grid max-w-xs gap-2">
          {features.map((feature) => {
            const Icon = feature.icon
            return (
              <div
                key={feature.id}
                className="bg-card flex items-center gap-3 rounded-md border p-3"
              >
                <span className="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-md">
                  <Icon className="size-4" />
                </span>
                <span className="min-w-0 flex-1 truncate text-sm font-medium">
                  {feature.name}
                </span>
                <Switch
                  checked={feature.enabled}
                  label={`${feature.enabled ? "关闭" : "开启"}${feature.name}`}
                  onCheckedChange={(checked) =>
                    setExperimentalEnabled(feature.id, checked)
                  }
                />
              </div>
            )
          })}
        </div>
      </section>
    </div>
  )
}
