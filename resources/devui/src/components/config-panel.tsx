import type { ReactNode } from "react"
import type { LucideIcon } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { cn } from "@/lib/utils"

export function ConfigSection({
  icon: Icon,
  title,
  description,
  actions,
  children,
}: {
  icon: LucideIcon
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="overflow-hidden rounded-md border bg-card">
      <div className="flex items-center justify-between gap-2 border-b bg-muted/30 px-4 py-2.5">
        <div className="flex min-w-0 items-center gap-2.5">
          <span className="bg-background text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded border">
            <Icon className="size-3.5" />
          </span>
          <div className="min-w-0">
            <h2 className="truncate text-sm font-semibold">{title}</h2>
            {description && (
              <p className="text-muted-foreground truncate text-[11px]">{description}</p>
            )}
          </div>
        </div>
        {actions}
      </div>
      {children}
    </section>
  )
}

export function ConfigStatusPill({
  state,
}: {
  state: "enabled" | "disabled" | "empty"
}) {
  const label = state === "enabled" ? "已启用" : state === "disabled" ? "已停用" : "未配置"
  return (
    <span
      className={cn(
        "inline-flex h-5 shrink-0 items-center gap-1.5 rounded-full border px-2 text-[11px] leading-none font-medium",
        state === "enabled"
          ? "border-green-500/30 bg-green-500/10 text-green-600 dark:text-green-400"
          : state === "empty"
            ? "border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400"
            : "bg-muted/40 text-muted-foreground"
      )}
    >
      <span
        className={cn(
          "size-1.5 rounded-full",
          state === "enabled"
            ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
            : state === "disabled"
              ? "bg-muted-foreground/50"
              : "bg-amber-500"
        )}
      />
      {label}
    </span>
  )
}

export function ConfigFieldError({ message }: { message?: string }) {
  if (!message) return null
  return <p className="text-destructive text-xs">{message}</p>
}

export function ConfigReadonlyRow({
  label,
  value,
  mono = false,
  action,
}: {
  label: string
  value?: string
  mono?: boolean
  action?: ReactNode
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="flex items-center justify-between gap-4 py-2.5">
      <span className="text-muted-foreground w-24 shrink-0 text-xs">{label}</span>
      <div className="flex min-w-0 flex-1 items-center justify-end gap-1.5">
        <TooltipHint content={text}>
          <span
            className={cn(
              "min-w-0 truncate text-sm",
              mono && "font-mono",
              text === "—" && "text-muted-foreground"
            )}
          >
            {text}
          </span>
        </TooltipHint>
        {action}
      </div>
    </div>
  )
}

export function ConfigEmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: LucideIcon
  title: string
  description: string
  action?: ReactNode
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-md border border-dashed bg-card px-6 py-14 text-center">
      <div className="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-full">
        <Icon className="size-6" />
      </div>
      <div className="grid gap-1">
        <h2 className="text-sm font-semibold">{title}</h2>
        <p className="text-muted-foreground text-xs">{description}</p>
      </div>
      {action}
    </div>
  )
}