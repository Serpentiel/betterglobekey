// The keys that can switch input sources, mirroring the daemon's TriggerKeys.

export interface TriggerKey {
  /** value is the config name, e.g. right_command. */
  value: string
  /** label is how the key is shown on its own, e.g. "Right ⌘ Command". */
  label: string
  /** name is how the key is named inside a sentence, e.g. "Right Command". */
  name: string
  /** code is the KeyboardEvent.code a key press reports, or null for Globe, which web views never see. */
  code: string | null
  /** passThrough reports whether the key can keep its normal function while it triggers. */
  passThrough: boolean
}

export const GLOBE = 'globe'

/** DEDICATED_USAGE is the HID usage a dedicated trigger is silenced to (F24), mirroring the daemon's DedicatedUsage. */
export const DEDICATED_USAGE = 0x700000073

/** Trigger modes, mirroring the daemon's TriggerModes. */
export const MODE_FN = 'fn'
export const MODE_DEDICATED = 'dedicated'
export const MODE_PASS_THROUGH = 'pass_through'
export const TRIGGER_MODES = [MODE_FN, MODE_DEDICATED, MODE_PASS_THROUGH] as const

/** Keyboard scopes, mirroring the daemon's KeyboardScopes. */
export const KEYBOARDS_WITHOUT_GLOBE = 'without_globe'
export const KEYBOARDS_ALL = 'all'
export const KEYBOARD_SCOPES = [KEYBOARDS_WITHOUT_GLOBE, KEYBOARDS_ALL] as const

/** KEYBOARD_ID matches a vendor:product keyboard ID. */
export const KEYBOARD_ID = /^[0-9a-f]{4}:[0-9a-f]{4}$/

const functionKeys: TriggerKey[] = [13, 14, 15, 16, 17, 18, 19, 20].map((number) => ({
  value: `f${number}`,
  label: `F${number}`,
  name: `F${number}`,
  code: `F${number}`,
  passThrough: true,
}))

export const TRIGGER_KEYS: readonly TriggerKey[] = [
  { value: GLOBE, label: '🌐 Globe', name: 'Globe', code: null, passThrough: false },
  { value: 'caps_lock', label: '⇪ Caps Lock', name: 'Caps Lock', code: 'CapsLock', passThrough: false },
  { value: 'right_command', label: 'Right ⌘ Command', name: 'Right Command', code: 'MetaRight', passThrough: true },
  { value: 'right_option', label: 'Right ⌥ Option', name: 'Right Option', code: 'AltRight', passThrough: true },
  { value: 'right_control', label: 'Right ⌃ Control', name: 'Right Control', code: 'ControlRight', passThrough: true },
  { value: 'right_shift', label: 'Right ⇧ Shift', name: 'Right Shift', code: 'ShiftRight', passThrough: true },
  { value: 'left_command', label: 'Left ⌘ Command', name: 'Left Command', code: 'MetaLeft', passThrough: true },
  { value: 'left_option', label: 'Left ⌥ Option', name: 'Left Option', code: 'AltLeft', passThrough: true },
  { value: 'left_control', label: 'Left ⌃ Control', name: 'Left Control', code: 'ControlLeft', passThrough: true },
  { value: 'left_shift', label: 'Left ⇧ Shift', name: 'Left Shift', code: 'ShiftLeft', passThrough: true },
  ...functionKeys,
  { value: 'menu', label: '▤ Menu', name: 'Menu', code: 'ContextMenu', passThrough: true },
]

export const TRIGGER_KEY_VALUES = TRIGGER_KEYS.map((key) => key.value) as [string, ...string[]]

/** triggerKey returns the trigger key with the given config name. */
export function triggerKey(value: string): TriggerKey | undefined {
  return TRIGGER_KEYS.find((key) => key.value === value)
}

/** triggerKeyForCode returns the trigger key a KeyboardEvent.code belongs to. */
export function triggerKeyForCode(code: string): TriggerKey | undefined {
  return TRIGGER_KEYS.find((key) => key.code === code)
}

/** triggerLabel names the key, falling back to its config name. */
export function triggerLabel(value: string): string {
  return triggerKey(value)?.label ?? value
}

/** allowedModes lists the modes a key can use: every mode but pass-through for
 * keys that cannot keep their function. */
export function allowedModes(value: string): string[] {
  return TRIGGER_MODES.filter((mode) => !(mode === MODE_PASS_THROUGH && triggerKey(value)?.passThrough === false))
}

/** remappedTrigger reports whether the trigger key is remapped at the HID level,
 * mirroring the daemon's Trigger.Remapped. */
export function remappedTrigger(trigger: { key: string; mode: string }): boolean {
  return trigger.key !== GLOBE && trigger.mode !== MODE_PASS_THROUGH
}

/** includesKeyboard reports whether a keyboard is in scope, mirroring the daemon's Keyboards.Includes. */
export function includesKeyboard(
  keyboards: { applyTo: string; overrides: Record<string, boolean> },
  id: string,
  hasGlobe: boolean,
): boolean {
  const override = keyboards.overrides[id]
  if (override !== undefined) {
    return override
  }

  return keyboards.applyTo === KEYBOARDS_ALL || !hasGlobe
}

/** triggerName names the key inside a sentence, falling back to its config name. */
export function triggerName(value: string): string {
  return triggerKey(value)?.name ?? value
}
