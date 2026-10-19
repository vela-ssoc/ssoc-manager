import { useEffect, useState } from "react"
import { ChevronDown, ChevronLeft, ChevronRight, ChevronUp } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const PAGE_SIZES = [12, 24, 36, 48, 60, 120]

export function DataPagination({
  current,
  size,
  total,
  loading = false,
  onPageChange,
  onPageSizeChange,
}: {
  current: number
  size: number
  total: number
  loading?: boolean
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
}) {
  const pageCount = Math.max(1, Math.ceil(total / size))
  const [pageInput, setPageInput] = useState(String(current))

  useEffect(() => {
    setPageInput(String(current))
  }, [current])

  const commitPage = () => {
    const parsed = Number.parseInt(pageInput, 10)
    if (!Number.isFinite(parsed)) {
      setPageInput(String(current))
      return
    }
    onPageChange(Math.max(1, Math.min(pageCount, parsed)))
  }

  const stepPage = (step: number) => {
    const parsed = Number.parseInt(pageInput, 10)
    const base = Number.isFinite(parsed) ? parsed : current
    const next = Math.max(1, Math.min(pageCount, base + step))
    setPageInput(String(next))
    onPageChange(next)
  }

  return (
    <div className="text-muted-foreground flex w-full items-center justify-between gap-3 pr-2 text-xs">
      <span className="shrink-0">共 {total} 条</span>
      <div className="flex items-center gap-2">
        <span className="hidden sm:inline">每页</span>
        <Select
          value={String(size)}
          onValueChange={(value) => onPageSizeChange(Number(value))}
          disabled={loading}
        >
          <SelectTrigger size="sm" className="w-16 px-1.5">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {PAGE_SIZES.map((pageSize) => (
              <SelectItem key={pageSize} value={String(pageSize)}>
                {pageSize}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="icon"
          className="size-7"
          aria-label="上一页"
          disabled={current <= 1 || loading}
          onClick={() => onPageChange(Math.max(1, current - 1))}
        >
          <ChevronLeft />
        </Button>
        <div className="flex items-center gap-1">
          <div className="relative">
            <Input
              aria-label="跳转到页码"
              className="h-7 w-16 appearance-none px-1 pr-5 text-center"
              type="text"
              inputMode="numeric"
              value={pageInput}
              disabled={loading}
              onChange={(event) => setPageInput(event.target.value.replace(/\D/g, ""))}
              onBlur={commitPage}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  commitPage()
                  event.currentTarget.blur()
                }
              }}
            />
            <div className="absolute top-0 right-0 flex h-full w-4 flex-col border-l">
              <button
                type="button"
                aria-label="页码加一"
                className="text-muted-foreground hover:bg-accent hover:text-foreground flex min-h-0 flex-1 cursor-pointer items-center justify-center rounded-tr-sm"
                disabled={loading}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => stepPage(1)}
              >
                <ChevronUp className="size-2.5" />
              </button>
              <button
                type="button"
                aria-label="页码减一"
                className="text-muted-foreground hover:bg-accent hover:text-foreground flex min-h-0 flex-1 cursor-pointer items-center justify-center rounded-br-sm"
                disabled={loading}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => stepPage(-1)}
              >
                <ChevronDown className="size-2.5" />
              </button>
            </div>
          </div>
          <span>/ {pageCount}</span>
        </div>
        <Button
          variant="outline"
          size="icon"
          className="size-7"
          aria-label="下一页"
          disabled={current >= pageCount || loading}
          onClick={() => onPageChange(Math.min(pageCount, current + 1))}
        >
          <ChevronRight />
        </Button>
      </div>
    </div>
  )
}
