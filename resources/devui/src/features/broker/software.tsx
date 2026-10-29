import { useCallback, useEffect, useState } from "react"
import { toast } from "@/lib/toast"

import { BuildInfoView } from "@/features/buildinfo/build-info"
import type { BuildInfo } from "@/features/buildinfo/types"
import { brokerApi } from "@/features/broker/api"

export function SoftwareSection({ slug }: { slug: string }) {
  const [info, setInfo] = useState<BuildInfo | null>(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setInfo(await brokerApi.buildInfo(slug))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载软件信息失败")
    } finally {
      setLoading(false)
    }
  }, [slug])

  useEffect(() => {
    void load()
  }, [load])

  return <BuildInfoView info={info} loading={loading} />
}