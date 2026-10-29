import { http } from "@/lib/http"

import type {
  ProcessConnection,
  ProcessInfo,
  ProcessOpenFile,
  ProcessRecords,
} from "./types"

function systemPath(slug?: string): string {
  return slug ? `/br/${slug}/api/v3/system` : "/system"
}

export const systemProcessApi = {
  list(slug?: string) {
    return http.get<ProcessRecords<ProcessInfo>>(`${systemPath(slug)}/processes`)
  },

  connections(slug: string | undefined, pid: number) {
    return http.get<ProcessRecords<ProcessConnection>>(
      `${systemPath(slug)}/process/connections`,
      { pid }
    )
  },

  openFiles(slug: string | undefined, pid: number) {
    return http.get<ProcessRecords<ProcessOpenFile>>(
      `${systemPath(slug)}/process/openfiles`,
      { pid }
    )
  },
}