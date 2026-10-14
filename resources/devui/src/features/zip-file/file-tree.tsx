import { useEffect, useRef, useState } from "react"
import { Trash2 } from "lucide-react"

import { cn } from "@/lib/utils"

import { CollapseIcon, FileIcon } from "./file-icon"
import type { TreeNode } from "./types"

function isDirNode(node: TreeNode): boolean {
  return node.children.length > 0 || Boolean(node.file?.is_directory)
}

export function FileTree({
  nodes,
  selectedPath,
  onSelect,
  onDelete,
  onContextMenu,
}: {
  nodes: TreeNode[]
  selectedPath: string
  onSelect: (node: TreeNode) => void
  onDelete?: (node: TreeNode) => void
  onContextMenu?: (e: React.MouseEvent, node: TreeNode | null) => void
}) {
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set())
  const initialized = useRef(false)

  // 首次加载时默认展开根级目录，深层目录折叠。
  useEffect(() => {
    if (!initialized.current && nodes.length > 0) {
      const roots = nodes.filter(isDirNode).map((n) => n.path)
      if (roots.length > 0) setExpanded(new Set(roots))
      initialized.current = true
    }
  }, [nodes])

  const toggleNode = (path: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  

  return (
    <div
      className="flex h-full flex-col"
      onContextMenu={(e) => {
        e.preventDefault()
        onContextMenu?.(e, null)
      }}
    >
      <div className="min-h-0 flex-1 overflow-auto">
        {nodes.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            暂无文件
          </div>
        ) : (
          <div className="space-y-0.5 p-1">
            {nodes.map((node) => (
              <TreeNodeItem
                key={node.path}
                node={node}
                selectedPath={selectedPath}
                expanded={expanded}
                onToggle={toggleNode}
                onSelect={onSelect}
                onDelete={onDelete}
                onContextMenu={onContextMenu}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function TreeNodeItem({
  node,
  selectedPath,
  expanded,
  onToggle,
  onSelect,
  onDelete,
  onContextMenu,
}: {
  node: TreeNode
  selectedPath: string
  expanded: Set<string>
  onToggle: (path: string) => void
  onSelect: (node: TreeNode) => void
  onDelete?: (node: TreeNode) => void
  onContextMenu?: (e: React.MouseEvent, node: TreeNode | null) => void
}) {
  const isDir = isDirNode(node)
  const open = expanded.has(node.path)
  const selected = node.path === selectedPath

  return (
    <div>
      <div
        role="button"
        tabIndex={0}
        className={cn(
          "group flex w-full items-center gap-1 rounded px-2 py-1 text-sm transition-colors hover:bg-accent/60",
          selected && "bg-accent text-accent-foreground"
        )}
        onClick={() => {
          if (isDir) {
            onToggle(node.path)
          } else {
            onSelect(node)
          }
        }}
        onContextMenu={(e) => {
          e.preventDefault()
          e.stopPropagation()
          if (node.file) onSelect(node)
          onContextMenu?.(e, node)
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            if (isDir) onToggle(node.path)
            else onSelect(node)
          }
        }}
      >
        <span className="flex size-4 shrink-0 items-center justify-center">
          {isDir ? <CollapseIcon open={open} className="size-3.5" /> : null}
        </span>
        <FileIcon
          name={node.name}
          isDir={isDir}
          className="size-4 shrink-0"
        />
        <span className="min-w-0 flex-1 truncate">{node.name}</span>
        {onDelete && (
          <button
            type="button"
            className="text-muted-foreground hover:text-destructive z-10 hidden size-5 shrink-0 items-center justify-center rounded group-hover:flex"
            title="删除"
            onClick={(e) => {
              e.stopPropagation()
              onDelete(node)
            }}
          >
            <Trash2 className="size-3.5" />
          </button>
        )}
      </div>

      {isDir && open && (
        <div className="ml-3.5 pl-1">
          {node.children.map((child) => (
            <TreeNodeItem
              key={child.path}
              node={child}
              selectedPath={selectedPath}
              expanded={expanded}
              onToggle={onToggle}
              onSelect={onSelect}
              onDelete={onDelete}
              onContextMenu={onContextMenu}
            />
          ))}
        </div>
      )}
    </div>
  )
}