import { Navigate, Route, Routes } from "react-router-dom"

import { AppLayout } from "@/components/layout/app-layout"
import { ZipArchivePage } from "@/pages/zip-archive"
import { ZipArchiveDetailPage } from "@/pages/zip-archive-detail"
import { ReverseProxiesPage } from "@/pages/reverse-proxies"

export default function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<Navigate to="/zip-archives" replace />} />
        <Route path="/zip-archives" element={<ZipArchivePage />} />
        <Route path="/zip-archives/:name" element={<ZipArchiveDetailPage />} />
        <Route path="/reverse-proxies" element={<ReverseProxiesPage />} />
      </Route>
    </Routes>
  )
}
