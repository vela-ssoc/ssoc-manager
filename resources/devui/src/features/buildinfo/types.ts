export interface BuildModule {
  Path?: string
  Version?: string
  Sum?: string
  Replace?: BuildModule
}

export interface BuildSetting {
  Key: string
  Value: string
}

export interface BuildModuleInfo {
  GoVersion: string
  Path: string
  Main: BuildModule
  Deps: BuildModule[]
  Settings: BuildSetting[]
}

export interface BuildInfo {
  goos: string
  goarch: string
  version: string
  revision: string
  username: string
  workdir: string
  module: string
  committed_at: string
  build_info?: BuildModuleInfo
}