import { createClient, type WebDAVClient } from "webdav"

import { buildAuthUrl, getToken } from "@/lib/http"

export interface DavEntry {
  name: string
  path: string
  isDir: boolean
  size: number
  lastmod: string
}

export interface DavEndpoint {
  clientUrl: string
  apiPath: string
}

const DAV_BASE_PATH = "/api/v3/webdav"

export function brokerDavEndpoint(slug: string): DavEndpoint {
  return {
    clientUrl: `/api/v3/br/${slug}/api/v3/webdav`,
    apiPath: `/br/${slug}/api/v3/webdav`,
  }
}

export function centralDavEndpoint(): DavEndpoint {
  return {
    clientUrl: "/api/v3/webdav",
    apiPath: "/webdav",
  }
}

export function createDavClient(endpoint: DavEndpoint): WebDAVClient {
  return createClient(endpoint.clientUrl, {
    headers: { Authorization: getToken() ?? "" },
    remoteBasePath: DAV_BASE_PATH,
  })
}

export async function davList(
  endpoint: DavEndpoint,
  dir: string
): Promise<DavEntry[]> {
  const self = dir.replace(/^\/+|\/+$/g, "")
  const list = await createDavClient(endpoint).getDirectoryContents(self)

  const selfName = self.split("/").filter(Boolean).pop() ?? ""
  const result: DavEntry[] = []

  for (const stat of list) {
    if (!stat.basename) continue
    if (stat.type === "directory" && stat.basename === selfName) continue

    result.push({
      name: stat.basename,
      path: self ? `${self}/${stat.basename}` : stat.basename,
      isDir: stat.type === "directory",
      size: stat.size ?? 0,
      lastmod: stat.lastmod ?? "",
    })
  }

  return result
}

export function davDownloadUrl(endpoint: DavEndpoint, fullPath: string): string {
  const encoded = fullPath
    .split("/")
    .filter(Boolean)
    .map((segment) => encodeURIComponent(segment))
    .join("/")
  return buildAuthUrl(`${endpoint.apiPath}/${encoded}`)
}

export function splitPath(path: string): Array<{ name: string; path: string }> {
  const segments = path.split("/").filter(Boolean)
  const parts: Array<{ name: string; path: string }> = []
  let acc = ""
  for (const segment of segments) {
    acc = `${acc}/${segment}`
    parts.push({ name: segment, path: acc })
  }
  return parts
}

export function joinPath(parent: string, name: string): string {
  const base = parent.replace(/\/+$/, "")
  return base ? `${base}/${name}` : name
}