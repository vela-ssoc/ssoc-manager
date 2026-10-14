import { buildAuthUrl, download, http } from "@/lib/http"

import type { ZipArchiveForm } from "@/features/zip-archive/types"

import type { ZipFileListResult } from "./types"

export const zipFileApi = {
  list(name: string) {
    return http.get<ZipFileListResult>("/zip-file/readdir", { name })
  },

  content(name: string, fullName: string) {
    return download("/zip-file/content", { name, full_name: fullName })
  },

  contentUrl(name: string, fullName: string) {
    return buildAuthUrl("/zip-file/content", { name, full_name: fullName })
  },

  update(name: string, fullName: string, content: string, fileName?: string) {
    const form = new FormData()
    form.set("name", name)
    form.set("full_name", fullName)
    const blob = new Blob([content], { type: "text/plain" })
    form.set("file", blob, fileName || fullName.split("/").pop() || "file")
    return http.post<unknown>("/zip-file/content", form)
  },

  create(name: string, fullName: string, file: File) {
    const form = new FormData()
    form.set("name", name)
    form.set("full_name", fullName)
    form.set("file", file)
    return http.post<unknown>("/zip-file/content", form)
  },

  unzip(form: ZipArchiveForm, file: File) {
    const data = new FormData()
    data.set("name", form.name)
    data.set("description", form.description)
    data.set("method", String(form.method))
    data.set("password", form.password)
    data.set("file", file)
    return http.post<unknown>("/zip-file/unzip", data)
  },

  mkdir(name: string, fullName: string) {
    return http.post<unknown>("/zip-file/mkdir", { name, full_name: fullName })
  },

  remove(name: string, fullName: string) {
    return http.delete<unknown>("/zip-file/content", { name, full_name: fullName })
  },
}