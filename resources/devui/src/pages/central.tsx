import { useState } from "react"

import { cn } from "@/lib/utils"
import { WebTerminal } from "@/components/web-terminal"
import { WebDavExplorer } from "@/features/broker/webdav-explorer"
import { centralDavEndpoint } from "@/features/broker/webdav"

type Tab = "files" | "terminal"

const TABS: { key: Tab; label: string }[] = [
  { key: "files", label: "文件管理" },
  { key: "terminal", label: "虚拟终端" },
]

export function CentralPage() {
  const [tab, setTab] = useState<Tab>("files")

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-1 border-b px-4">
        {TABS.map((item) => (
          <button
            key={item.key}
            type="button"
            className={cn(
              "border-b-2 px-3 py-2 text-sm transition-colors",
              tab === item.key
                ? "border-primary text-foreground font-medium"
                : "text-muted-foreground border-transparent hover:text-foreground"
            )}
            onClick={() => setTab(item.key)}
          >
            {item.label}
          </button>
        ))}
      </div>

      <div className="min-h-0 flex-1 p-4">
        {tab === "files" && (
          <WebDavExplorer {...centralDavEndpoint()} />
        )}

        {tab === "terminal" && (
          <WebTerminal path="/api/v3/system/pty" />
        )}
      </div>
    </div>
  )
}