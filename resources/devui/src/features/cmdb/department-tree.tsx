import { ChevronDown, ChevronRight } from "lucide-react"

import { cn } from "@/lib/utils"
import type { CMDBDepartmentNode } from "./types"

export function DepartmentTree({
  nodes,
  depth,
  expanded,
  selected,
  onSelect,
  onToggle,
}: {
  nodes: CMDBDepartmentNode[]
  depth: number
  expanded: Set<string>
  selected: string
  onSelect: (department: string) => void
  onToggle: (department: string) => void
}) {
  return (
    <>
      {nodes.map((node) => {
        const children = node.children ?? []
        const hasChildren = children.length > 0
        const open = expanded.has(node.department)
        const active = selected === node.department

        return (
          <div key={node.department || node.name}>
            <button
              type="button"
              aria-expanded={hasChildren ? open : undefined}
              className={cn(
                "flex w-full items-center gap-1 rounded-md pr-1 text-left transition-colors",
                active ? "bg-accent text-accent-foreground" : "hover:bg-accent/60"
              )}
              style={{ paddingLeft: depth * 12 }}
              onClick={() => {
                if (hasChildren) onToggle(node.department)
                onSelect(node.department)
              }}
            >
              {hasChildren ? (
                <span
                  aria-hidden="true"
                  className="text-muted-foreground flex size-5 shrink-0 items-center justify-center"
                >
                  {open ? (
                    <ChevronDown className="size-3.5" />
                  ) : (
                    <ChevronRight className="size-3.5" />
                  )}
                </span>
              ) : (
                <span className="size-5 shrink-0" />
              )}
              <span
                className={cn(
                  "min-w-0 flex-1 truncate py-1.5 text-xs",
                  active ? "" : "text-muted-foreground"
                )}
              >
                {node.name || "未命名部门"}
              </span>
            </button>
            {hasChildren && open && (
              <DepartmentTree
                nodes={children}
                depth={depth + 1}
                expanded={expanded}
                selected={selected}
                onSelect={onSelect}
                onToggle={onToggle}
              />
            )}
          </div>
        )
      })}
    </>
  )
}