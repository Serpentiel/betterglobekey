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

export interface KeyboardTrigger {
  key: string
  mode: string
}

/** Trigger holds the triggers of single keyboards by vendor:product ID; any
 * other keyboard uses the Globe key. */
/** Trigger turns the Globe key on or off for every keyboard that has one, and lists keyboards' own keys, each different. */
export interface Trigger {
  globe: boolean
  keyboards: Record<string, KeyboardTrigger[]>
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
  /** triggers are the keyboard's own trigger keys, the Globe key aside, and listed whether it is in Trigger.keyboards. */
  triggers: KeyboardTrigger[]
  listed: boolean
  /** triggerMappings are the daemon's own mappings, one per dedicated trigger key in place. */
  triggerMappings: KeyMapping[]
  otherMappings: KeyMapping[]
}

export interface KeyboardReport {
  keyboards: Keyboard[]
  systemMappings: KeyMapping[]
}

export interface RecordedKey {
  /** key is the trigger key pressed, or empty when the key cannot be a trigger. */
  key: string
  keyName: string
  keyboardId: string
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
  recordKey: (keyboardId: string) => Promise<RecordedKey>
  getVersion: () => Promise<Version>
  getCompanionVersion: () => Promise<string>
}
