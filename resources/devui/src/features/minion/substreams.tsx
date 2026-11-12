import { useCallback } from "react"

import { SubstreamsPanel } from "@/features/broker/substreams"
import type { BrokerSubstream } from "@/features/broker/types"
import { minionApi } from "@/features/minion/api"

export function AgentSubstreamsSection({
  agentId,
  brokerSlug,
}: {
  agentId: string
  brokerSlug: string
}) {
  const load = useCallback(
    async (): Promise<BrokerSubstream[]> => {
      const ret = await minionApi.substreams(brokerSlug, agentId)
      return ret.records ?? []
    },
    [agentId, brokerSlug]
  )

  const kill = useCallback(
    (streamId: string) =>
      minionApi.killSubstream(brokerSlug, agentId, streamId),
    [agentId, brokerSlug]
  )

  return <SubstreamsPanel load={load} kill={kill} />
}