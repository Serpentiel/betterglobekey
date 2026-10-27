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

export interface Keyboards {
  applyTo: string
  /** overrides include (true) or exclude (false) keyboards by vendor:product ID. */
  overrides: Record<string, boolean>
}

export interface Trigger {
  key: string
  mode: string
  keyboards: Keyboards
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

export interface KeyMapping {
  src: number
  dst: number
  srcName: string
  dstName: string
}

export interface Keyboard {
  id: string
  name: string
  transport: string
  builtIn: boolean
  hasGlobe: boolean
  inScope: boolean
  /** triggerMapping is the daemon's own mapping, or absent when it has none. */
  triggerMapping?: KeyMapping | null
  otherMappings: KeyMapping[]
}

export interface KeyboardReport {
  keyboards: Keyboard[]
  systemMappings: KeyMapping[]
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
  listKeyboards: () => Promise<KeyboardReport>
  getVersion: () => Promise<Version>
  getCompanionVersion: () => Promise<string>
}
