import { useEffect, useState } from "react"
import { LoaderCircle } from "lucide-react"

import { useIsDark } from "@/features/zip-file/file-icon"
import { toast } from "@/lib/toast"

const OAUTH_URL =
  import.meta.env.VITE_OAUTH_URL ??
  "https://dongdong-auth.eastmoney.com/dongdong-auth/login/"
const OAUTH_CLIENT_ID = "vela-ssoc"

const QR_SIZE = 150

// iframe 内部页面以更宽视口渲染（保证二维码水平居中且不出现横向滚动条），
// 再通过负 marginLeft 把二维码对齐到可见窗口中央。
const FRAME_WIDTH = 320

// 登录系统二维码上方标题高度（`text-[16px]` 行高 24px + `mt-2` 8px + `mb-4` 16px ≈ 48px），
// 向上裁剪只露出二维码。
const TITLE_CROP = 49

// 可见窗口四周留白，避免二维码紧贴边框。
const PAD = 4

interface DongDongMessage {
  type?: string
  data?: string
}

const OAUTH_REDIRECT_URI = "https://vela-ssoc.eastmoney.com"

function extractCode(redirectUri: string): string | null {
  try {
    const url = new URL(redirectUri, window.location.href)
    return url.searchParams.get("code")
  } catch {
    return null
  }
}

export function OAuthQrLogin({
  onCode,
}: {
  onCode: (code: string) => Promise<void>
}) {
  const [frameUrl, setFrameUrl] = useState("")
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const dark = useIsDark()

  const loadFrame = () => {
    if (!OAUTH_URL || !OAUTH_CLIENT_ID) {
      setLoading(false)
      return
    }
    const url = new URL(OAUTH_URL)
    url.searchParams.set("client_id", OAUTH_CLIENT_ID)
    url.searchParams.set("redirect_uri", OAUTH_REDIRECT_URI)
    url.searchParams.set("qr", "1")
    url.searchParams.set("img_size", String(QR_SIZE))
    setFrameUrl(url.toString())
    setLoading(false)
  }

  useEffect(() => {
    loadFrame()
     
  }, [])

  useEffect(() => {
    const handler = (event: MessageEvent) => {
      let msg: DongDongMessage
      try {
        msg = JSON.parse(event.data) as DongDongMessage
      } catch {
        return
      }
      if (msg.type !== "DongDongOAuth2Login" || !msg.data) return

      const code = extractCode(msg.data)
      if (!code) {
        toast.error("扫码登录返回的地址中缺少授权码")
        return
      }
      setSubmitting(true)
      void onCode(code).finally(() => setSubmitting(false))
    }

    window.addEventListener("message", handler)
    return () => window.removeEventListener("message", handler)
  }, [onCode])

  return (
    <div className="grid justify-items-center gap-3">
      {frameUrl ? (
        <div
          className="relative overflow-hidden rounded-md border bg-white"
          style={{
            width: QR_SIZE + PAD * 2,
            height: QR_SIZE + PAD * 2,
          }}
        >
          <iframe
            key={frameUrl}
            title="扫码登录"
            src={frameUrl}
            sandbox="allow-scripts allow-same-origin"
            className="block border-0"
            style={{
              width: FRAME_WIDTH,
              height: QR_SIZE + TITLE_CROP,
              marginTop: PAD - TITLE_CROP,
              marginLeft: PAD - (FRAME_WIDTH - QR_SIZE) / 2 - 1,
            }}
          />
          {dark && (
            <div className="pointer-events-none absolute -inset-px bg-black/50" />
          )}
          {submitting && (
            <div className="absolute -inset-px flex items-center justify-center bg-background/70">
              <div className="flex items-center gap-2 text-sm">
                <LoaderCircle className="animate-spin" />
                正在登录…
              </div>
            </div>
          )}
        </div>
      ) : (
        <div
          className="flex items-center justify-center rounded-md border border-dashed"
          style={{ width: QR_SIZE + PAD * 2, height: QR_SIZE + PAD * 2 }}
        >
          {loading ? (
            <LoaderCircle className="text-muted-foreground animate-spin" />
          ) : (
            <span className="text-muted-foreground max-w-48 text-center text-xs">
              未配置扫码登录入口，请先完成 OAuth 配置
            </span>
          )}
        </div>
      )}
    </div>
  )
}