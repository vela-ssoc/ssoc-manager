import { http } from "@/lib/http"

import type { Certificate, CertificatePayload } from "./types"

export const certificateApi = {
  list() {
    return http.get<Certificate[]>("/certificates")
  },

  parse(publicKey: string, privateKey: string) {
    return http.post<Certificate>("/certificate/parse", {
      public_key: publicKey,
      private_key: privateKey,
    })
  },

  create(payload: CertificatePayload) {
    return http.post<unknown>("/certificate", payload)
  },

  update(id: string, payload: CertificatePayload) {
    return http.put<unknown>("/certificate", { id, ...payload })
  },

  remove(id: string) {
    return http.delete<unknown>("/certificate", { id })
  },
}