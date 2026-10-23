import { Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom"

import { AppLayout } from "@/components/layout/app-layout"
import { getToken } from "@/lib/http"
import { LoginPage } from "@/pages/login"
import { ZipArchivePage } from "@/pages/zip-archive"
import { ZipArchiveDetailPage } from "@/pages/zip-archive-detail"
import { ReverseProxiesPage } from "@/pages/reverse-proxies"
import { SettingsPage } from "@/pages/settings"
import { BrokersPage } from "@/pages/brokers"
import { BrokerDetailPage } from "@/pages/broker-detail"
import { MinionsPage } from "@/pages/minions"
import { LegacyPage } from "@/pages/legacy"
import { HoneDashboardPage } from "@/pages/hone-dashboard"
import { RiskDashboardPage } from "@/pages/risk-dashboard"
import { Oas3Page } from "@/pages/oas3"
import { CentralPage } from "@/pages/central"

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
          <Route path="/zip-archives" element={<ZipArchivePage />} />
          <Route path="/zip-archives/:name" element={<ZipArchiveDetailPage />} />
          <Route path="/reverse-proxies" element={<ReverseProxiesPage />} />
          <Route path="/brokers" element={<BrokersPage />} />
          <Route path="/brokers/:slug" element={<BrokerDetailPage />} />
          <Route path="/central" element={<CentralPage />} />
          <Route path="/minions" element={<MinionsPage />} />
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
