export function formatDateTime(value?: string | null): string {
  if (!value) return "-"
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const p = (n: number) => String(n).padStart(2, "0")
  return `${date.getFullYear()}-${p(date.getMonth() + 1)}-${p(date.getDate())} ` +
    `${p(date.getHours())}:${p(date.getMinutes())}:${p(date.getSeconds())}`
}

export function formatDuration(totalSeconds: number): string {
  const total = Math.max(0, Math.floor(totalSeconds))
  const d = Math.floor(total / 86400)
  const h = Math.floor((total % 86400) / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (d > 0) return `${d} 天 ${h} 时`
  if (h > 0) return `${h} 时 ${m} 分`
  if (m > 0) return `${m} 分 ${s} 秒`
  return `${s} 秒`
}

export function formatTimeout(ns: number): string {
  if (!ns || ns <= 0) return "不限时"
  const ms = ns / 1e6
  if (ms < 1000) return `${ms.toFixed(0)} 毫秒`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)} 秒`
  const m = s / 60
  if (m < 60) return `${m.toFixed(1)} 分钟`
  const h = m / 60
  return `${h.toFixed(1)} 小时`
}