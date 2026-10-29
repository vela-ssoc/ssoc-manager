export interface OAuthConfig {
  enabled: boolean
  client_id: string
  client_secret: string
  authorization_endpoint: string
  token_endpoint: string
  userinfo_endpoint: string
  redirect_uris: string[]
  scopes?: string[]
  created_at: string
  updated_at: string
}

export interface OAuthConfigForm {
  enabled: boolean
  client_id: string
  client_secret: string
  authorization_endpoint: string
  token_endpoint: string
  userinfo_endpoint: string
  redirect_uris: string[]
  scopes: string[]
}