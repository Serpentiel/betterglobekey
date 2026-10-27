// Types mirroring the betterglobekey.control.v1 protobuf messages, as the Tauri
// commands serialize them (camelCase field names, every field present).

export interface Logger {
  path: string
  level: string
  retentionDays: number
  retentionFiles: number
}

export interface DoublePress {
  enabled: boolean
  maximumDelay: string
}

export interface Reverse {
  enabled: boolean
  modifier: string
}

export interface Hud {
  enabled: boolean
  duration: string
  showCollection: boolean
}

export interface Trigger {
  key: string
  passThrough: boolean
}

export interface Collection {
  name: string
  sources: string[]
}

export interface Config {
  logger: Logger
  doublePress: DoublePress
  reverse: Reverse
  hud: Hud
  collections: Collection[]
  trigger: Trigger
}

export interface InputSource {
  id: string
  name: string
}

export interface Version {
  version: string
  commit: string
}

// Api is the surface the UI calls, implemented by lib/api.ts: the daemon, plus
// the companion's own version.
export interface Api {
  getConfig: () => Promise<Config>
  applyConfig: (config: Config) => Promise<void>
  listInputSources: () => Promise<InputSource[]>
  getVersion: () => Promise<Version>
  getCompanionVersion: () => Promise<string>
}
