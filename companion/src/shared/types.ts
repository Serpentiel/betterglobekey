// Types mirroring the betterglobekey.control.v1 protobuf messages, as decoded by
// @grpc/proto-loader (camelCase field names, defaults applied).

export interface Logger {
  path: string
  retentionDays: number
  retentionFiles: number
}

export interface Collection {
  name: string
  sources: string[]
}

export interface Config {
  logger: Logger
  doublePressMaxDelay: string
  hud: boolean
  collections: Collection[]
}

export interface InputSource {
  id: string
  name: string
}

// Api is the surface exposed to the renderer via the preload bridge.
export interface Api {
  getConfig: () => Promise<Config>
  applyConfig: (config: Config) => Promise<void>
  listInputSources: () => Promise<InputSource[]>
  getCurrentSource: () => Promise<InputSource>
}
