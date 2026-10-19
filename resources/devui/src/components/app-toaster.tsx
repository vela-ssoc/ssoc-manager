import { CircleAlert } from "lucide-react"
import { Toaster } from "sonner"

import { useTheme } from "@/providers/theme-provider"

export function AppToaster() {
  const { isDark } = useTheme()

  return (
    <Toaster
      theme={isDark ? "dark" : "light"}
      position="bottom-right"
      closeButton
      icons={{
        error: <CircleAlert className="size-4 text-primary" />,
      }}
      toastOptions={{
        classNames: {
          toast: "!border !border-primary/30 !bg-[var(--toast-bg)] !text-foreground !pr-8 !shadow-xl",
          title: "!text-foreground text-sm font-medium",
          description: "!text-muted-foreground text-xs",
          closeButton:
            "!left-auto !right-2 !top-2 !transform-none !border-0 !bg-transparent !text-muted-foreground hover:!bg-accent hover:!text-foreground",
        },
      }}
    />
  )
}
