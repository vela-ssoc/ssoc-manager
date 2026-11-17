import { lazy, Suspense, type ReactNode } from "react"
import { Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom"
import { Loader2 } from "lucide-react"

import { AppLayout } from "@/components/layout/app-layout"
import { getToken } from "@/lib/http"
import { useExperimentalFeatures } from "@/lib/experimental-features"
import { LoginPage } from "@/pages/login"
import { ZipArchivePage } from "@/pages/zip-archive"
import { ZipArchiveDetailPage } from "@/pages/zip-archive-detail"
import { ReverseProxiesPage } from "@/pages/reverse-proxies"
import { OAuthConfigPage } from "@/pages/oauth-config"
import { AIChatModelPage } from "@/pages/ai-chat-model"
import { CMDBConfigPage } from "@/pages/cmdb-config"
import { CMDMBrowsePage } from "@/pages/cmdb-browse"
import { CertificatePage } from "@/pages/certificate"
import { RedisConfigPage } from "@/pages/redis-config"
import { VictoriaMetricsPage } from "@/pages/victoria-metrics"
import { SettingsPage } from "@/pages/settings"
import { BrokersPage } from "@/pages/brokers"
import { BrokerDetailPage } from "@/pages/broker-detail"
import { MinionsPage } from "@/pages/minions"
import { MinionDetailPage } from "@/pages/minion-detail"
import { LegacyPage } from "@/pages/legacy"
import { HoneDashboardPage } from "@/pages/hone-dashboard"
import { RiskDashboardPage } from "@/pages/risk-dashboard"
import { Oas3Page } from "@/pages/oas3"
import { CentralPage } from "@/pages/central"

const AIChatPage = lazy(() =>
  import("@/pages/ai-chat").then((mod) => ({ default: mod.AIChatPage }))
)

function PageLoading() {
  return (
    <div className="flex h-full items-center justify-center">
      <Loader2 className="text-muted-foreground size-5 animate-spin" />
    </div>
  )
}

function ExperimentalRoute({
  id,
  children,
}: {
  id: string
  children: ReactNode
}) {
  const features = useExperimentalFeatures()
  const enabled = features.some((feature) => feature.id === id && feature.enabled)
  if (!enabled) {
    return <Navigate to="/settings" replace />
  }
  return children
}

function RequireAuth() {
  const location = useLocation()
  if (!getToken()) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }
  return <Outlet />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RequireAuth />}>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/legacy" replace />} />
          <Route
            path="/ai-chat"
            element={
              <Suspense fallback={<PageLoading />}>
                <ExperimentalRoute id="ai-chat">
                  <AIChatPage />
                </ExperimentalRoute>
              </Suspense>
            }
          />
          <Route path="/zip-archives" element={<ZipArchivePage />} />
          <Route path="/zip-archives/:name" element={<ZipArchiveDetailPage />} />
          <Route path="/reverse-proxies" element={<ReverseProxiesPage />} />
          <Route path="/configs/oauth" element={<OAuthConfigPage />} />
          <Route path="/configs/ai-chat-model" element={<AIChatModelPage />} />
          <Route path="/configs/cmdb" element={<CMDBConfigPage />} />
          <Route path="/configs/cmdb/browse" element={<CMDMBrowsePage />} />
          <Route path="/configs/certificate" element={<CertificatePage />} />
          <Route path="/configs/victoria-metrics" element={<VictoriaMetricsPage />} />
          <Route path="/configs/redis" element={<RedisConfigPage />} />
          <Route path="/brokers" element={<BrokersPage />} />
          <Route path="/brokers/:slug" element={<BrokerDetailPage />} />
          <Route path="/central" element={<CentralPage />} />
          <Route path="/minions" element={<MinionsPage />} />
          <Route path="/minions/:id" element={<MinionDetailPage />} />
          <Route path="/legacy" element={<LegacyPage />} />
          <Route path="/hone" element={<HoneDashboardPage />} />
          <Route path="/ssoc" element={<RiskDashboardPage />} />
          <Route path="/oas3" element={<Oas3Page />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
