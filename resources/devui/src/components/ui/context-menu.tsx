import * as React from "react"

import { cn } from "@/lib/utils"

export interface ContextMenuState {
  x: number
  y: number
}

export function ContextMenu({
  state,
  onClose,
  children,
  className,
}: {
  state: ContextMenuState | null
  onClose: () => void
  children: React.ReactNode
  className?: string
}) {
  const ref = React.useRef<HTMLDivElement | null>(null)

  React.useEffect(() => {
    if (!state) return

    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose()
      }
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    const onScroll = () => onClose()
    const onResize = () => onClose()

    document.addEventListener("mousedown", onDown)
    document.addEventListener("keydown", onKey)
    window.addEventListener("scroll", onScroll, true)
    window.addEventListener("resize", onResize)

    return () => {
      document.removeEventListener("mousedown", onDown)
      document.removeEventListener("keydown", onKey)
      window.removeEventListener("scroll", onScroll, true)
      window.removeEventListener("resize", onResize)
    }
  }, [state, onClose])

  if (!state) return null

  return (
    <div
      ref={ref}
      role="menu"
      className={cn(
        "bg-popover text-popover-foreground fixed z-50 min-w-[10rem] rounded-md border p-1 shadow-md",
        className
      )}
      style={{ left: state.x, top: state.y }}
    >
      {children}
    </div>
  )
}

export function ContextMenuItem({
  className,
  variant = "default",
  ...props
}: React.ComponentProps<"button"> & { variant?: "default" | "destructive" }) {
  return (
    <button
      type="button"
      role="menuitem"
      className={cn(
        "focus:bg-accent focus:text-accent-foreground hover:bg-accent hover:text-accent-foreground flex w-full cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm outline-none select-none [&_svg]:size-4 [&_svg]:shrink-0",
        variant === "destructive" &&
          "text-destructive hover:bg-destructive/10 hover:text-destructive focus:bg-destructive/10 focus:text-destructive",
        className
      )}
      {...props}
    />
  )
}

export function ContextMenuSeparator({ className }: { className?: string }) {
  return <div className={cn("bg-border -mx-1 my-1 h-px", className)} />
}