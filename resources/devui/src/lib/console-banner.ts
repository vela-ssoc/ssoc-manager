let printed = false

export function printConsoleBanner(version: string) {
  if (printed) return
  printed = true

  console.log(
    "%c SSOC %c  Smart Security Operations Center  %c " + version + " ",
    "color:#ffffff;background:#f97316;font-weight:700;font-size:14px;line-height:26px;padding:2px 10px;",
    "color:#fb923c;background:#111827;font-size:12px;line-height:26px;padding:2px 10px;",
    "color:#e5e7eb;background:#111827;font-size:12px;line-height:26px;padding:2px 10px;"
  )
  console.log(
    "%c SSOC 安全运营中心 · 控制台仅供调试，谨防信息泄露 ",
    "color:#6b7280;font-size:12px;line-height:22px;padding:4px 2px;"
  )
}