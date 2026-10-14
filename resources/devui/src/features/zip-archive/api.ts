import { buildAuthUrl, http } from "@/lib/http"

import type { PageQuery, PageResult, ZipArchive, ZipArchiveForm } from "./types"

export const zipArchiveApi = {
  page(query: PageQuery) {
    return http.get<PageResult<ZipArchive>>("/zip-archives", query)
  },

  downloadUrl(name: string) {
    return buildAuthUrl("/zip-archive/download", { name })
  },

  info(name: string) {
    return http.get<ZipArchive>("/zip-archive/info", { name })
  },

  create(data: ZipArchiveForm) {
    return http.post<unknown>("/zip-archive", data)
  },

  update(id: string, data: ZipArchiveForm) {
    return http.put<unknown>("/zip-archive", { id, ...data })
  },
}