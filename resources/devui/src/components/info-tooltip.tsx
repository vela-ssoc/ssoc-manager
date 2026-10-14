import type { ReactNode } from "react"
import { CircleHelp } from "lucide-react"

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"

export function InfoTooltip({
  children,
  label = "更多信息",
}: {
  children: ReactNode
  label?: string
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          aria-label={label}
          className="text-muted-foreground hover:text-foreground inline-flex size-4 shrink-0 items-center justify-center rounded-full transition-colors"
        >
          <CircleHelp className="size-3.5" />
        </button>
      </TooltipTrigger>
      <TooltipContent className="max-w-80 whitespace-normal leading-relaxed">
        <div className="space-y-1.5 [&_ol]:list-decimal [&_ol]:space-y-1 [&_ol]:pl-4 [&_ul]:list-disc [&_ul]:space-y-1 [&_ul]:pl-4">
          {children}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}
