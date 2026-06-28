// Types mirroring the betterglobekey.control.v1 protobuf messages, as decoded by
// @grpc/proto-loader (camelCase field names, defaults applied).

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
}

export interface InputSource {
  id: string
  name: string
}

export interface Version {
  version: string
  commit: string
}

// Api is the surface exposed to the renderer via the preload bridge.
export interface Api {
  getConfig: () => Promise<Config>
  applyConfig: (config: Config) => Promise<void>
  listInputSources: () => Promise<InputSource[]>
  getCurrentSource: () => Promise<InputSource>
  getVersion: () => Promise<Version>
}
