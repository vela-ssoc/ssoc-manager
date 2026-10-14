import { useMemo, type Ref } from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

export function MarkdownPreview({
  content,
  containerRef,
}: {
  content: string
  containerRef?: Ref<HTMLDivElement>
}) {
  return (
    <div ref={containerRef} className="markdown-body h-full overflow-auto p-4">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  )
}

function parseCsv(text: string): string[][] {
  const rows: string[][] = []
  let row: string[] = []
  let cell = ""
  let inQuotes = false

  for (let i = 0; i < text.length; i++) {
    const ch = text[i]
    if (inQuotes) {
      if (ch === '"') {
        if (text[i + 1] === '"') {
          cell += '"'
          i++
        } else {
          inQuotes = false
        }
      } else {
        cell += ch
      }
    } else if (ch === '"') {
      inQuotes = true
    } else if (ch === ",") {
      row.push(cell)
      cell = ""
    } else if (ch === "\n") {
      row.push(cell)
      rows.push(row)
      row = []
      cell = ""
    } else if (ch === "\r") {
      // 跳过 \r
    } else {
      cell += ch
    }
  }
  if (cell !== "" || row.length > 0) {
    row.push(cell)
    rows.push(row)
  }

  // 去除末尾空行
  while (rows.length && rows[rows.length - 1].length === 1 && rows[rows.length - 1][0] === "") {
    rows.pop()
  }
  return rows
}

export function CsvTable({ content }: { content: string }) {
  const rows = useMemo(() => parseCsv(content), [content])
  if (rows.length === 0) {
    return (
      <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
        空文件
      </div>
    )
  }

  const header = rows[0]
  const body = rows.slice(1)
  const hasHeader = header.length > 0

  return (
    <div className="h-full overflow-auto p-1">
      <table className="w-full border-collapse text-xs">
        <tbody>
          {hasHeader && (
            <tr>
              {header.map((cell, i) => (
                <th
                  key={i}
                  className="bg-muted/50 sticky top-0 border px-2 py-1 text-left font-semibold whitespace-nowrap"
                >
                  {cell}
                </th>
              ))}
            </tr>
          )}
          {body.map((row, ri) => (
            <tr key={ri} className="hover:bg-accent/40">
              {row.map((cell, ci) => (
                <td
                  key={ci}
                  className="max-w-80 truncate border px-2 py-1 whitespace-nowrap"
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}