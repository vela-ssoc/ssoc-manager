export interface CertificatePKIXName {
  country?: string[]
  organization?: string[]
  organizational_unit?: string[]
  locality?: string[]
  province?: string[]
  street_address?: string[]
  postal_code?: string[]
  serial_number?: string
  common_name?: string
}

export interface Certificate {
  id: string
  common_name: string
  certificate_sha256: string
  public_key_sha256: string
  private_key_sha256: string
  dns_names?: string[]
  ip_addresses?: string[]
  email_addresses?: string[]
  uris?: string[]
  version?: number
  not_before?: string
  not_after?: string
  issuer?: CertificatePKIXName
  subject?: CertificatePKIXName
  signature_algorithm?: string
  created_at?: string
  updated_at?: string
}

export interface CertificatePayload {
  public_key: string
  private_key: string
}