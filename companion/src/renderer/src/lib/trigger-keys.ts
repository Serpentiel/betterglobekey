// The keys that can switch input sources, mirroring the daemon's TriggerKeys.

import type { KeyboardTrigger, Trigger } from '../../../shared/types'

export interface TriggerKey {
  /** value is the config name, e.g. right_command. */
  value: string
  /** label is how the key is shown on its own, e.g. "Right ⌘ Command". */
  label: string
  /** name is how the key is named inside a sentence, e.g. "Right Command". */
  name: string
  /** passThrough reports whether the key can keep its normal function while it triggers. */
  passThrough: boolean
}

export const GLOBE = 'globe'

/** DEDICATED_USAGE is the HID usage a dedicated trigger is silenced to (F24), mirroring the daemon's DedicatedUsage. */
export const DEDICATED_USAGE = 0x700000073

/** Trigger modes, mirroring the daemon's TriggerModes. */
export const MODE_DEDICATED = 'dedicated'
export const MODE_PASS_THROUGH = 'pass_through'
export const TRIGGER_MODES = [MODE_DEDICATED, MODE_PASS_THROUGH] as const

/** GLOBE_TRIGGER is the trigger key of a keyboard with a Globe key of its own, unless it is listed. */
export const GLOBE_TRIGGER: KeyboardTrigger = { key: GLOBE, mode: MODE_DEDICATED }

/** KEYBOARD_ID matches a vendor:product keyboard ID. */
export const KEYBOARD_ID = /^[0-9a-f]{4}:[0-9a-f]{4}$/

const functionKeys: TriggerKey[] = [13, 14, 15, 16, 17, 18, 19, 20].map((number) => ({
  value: `f${number}`,
  label: `F${number}`,
  name: `F${number}`,
  passThrough: true,
}))

export const TRIGGER_KEYS: readonly TriggerKey[] = [
  { value: GLOBE, label: '🌐 Globe', name: 'Globe', passThrough: false },
  { value: 'caps_lock', label: '⇪ Caps Lock', name: 'Caps Lock', passThrough: false },
  { value: 'right_command', label: 'Right ⌘ Command', name: 'Right Command', passThrough: true },
  { value: 'right_option', label: 'Right ⌥ Option', name: 'Right Option', passThrough: true },
  { value: 'right_control', label: 'Right ⌃ Control', name: 'Right Control', passThrough: true },
  { value: 'right_shift', label: 'Right ⇧ Shift', name: 'Right Shift', passThrough: true },
  { value: 'left_command', label: 'Left ⌘ Command', name: 'Left Command', passThrough: true },
  { value: 'left_option', label: 'Left ⌥ Option', name: 'Left Option', passThrough: true },
  { value: 'left_control', label: 'Left ⌃ Control', name: 'Left Control', passThrough: true },
  { value: 'left_shift', label: 'Left ⇧ Shift', name: 'Left Shift', passThrough: true },
  ...functionKeys,
  { value: 'menu', label: '▤ Menu', name: 'Menu', passThrough: true },
  { value: 'insert', label: 'Insert', name: 'Insert', passThrough: true },
  { value: 'print_screen', label: 'Print Screen', name: 'Print Screen', passThrough: true },
  { value: 'scroll_lock', label: 'Scroll Lock', name: 'Scroll Lock', passThrough: true },
  { value: 'pause', label: 'Pause', name: 'Pause', passThrough: true },
]

export const TRIGGER_KEY_VALUES = TRIGGER_KEYS.map((key) => key.value) as [string, ...string[]]

/** triggerKey returns the trigger key with the given config name. */
export function triggerKey(value: string): TriggerKey | undefined {
  return TRIGGER_KEYS.find((key) => key.value === value)
}

/** triggerLabel names the key, falling back to its config name. */
export function triggerLabel(value: string): string {
  return triggerKey(value)?.label ?? value
}

/** triggerName names the key inside a sentence, falling back to its config name. */
export function triggerName(value: string): string {
  return triggerKey(value)?.name ?? value
}

/** allowedModes lists the modes a key can use: every mode but pass-through for
 * keys that cannot keep their function. */
export function allowedModes(value: string): string[] {
  return TRIGGER_MODES.filter((mode) => !(mode === MODE_PASS_THROUGH && triggerKey(value)?.passThrough === false))
}

/** remapped reports whether a trigger key is remapped at the HID level, mirroring the daemon's KeyboardTrigger.Remapped. */
export function remapped(trigger: KeyboardTrigger): boolean {
  return trigger.key !== GLOBE && trigger.mode === MODE_DEDICATED
}

/** defaultTriggers are the trigger keys of a keyboard that is not listed: its Globe key, if it has one. */
export function defaultTriggers(hasGlobe: boolean): KeyboardTrigger[] {
  return hasGlobe ? [GLOBE_TRIGGER] : []
}

/** triggersFor returns a keyboard's trigger keys and whether they are listed, mirroring the daemon's Trigger.For. */
export function triggersFor(trigger: Trigger, id: string, hasGlobe: boolean): [KeyboardTrigger[], boolean] {
  const listed = trigger.keyboards[id]
  if (listed) {
    return [listed, true]
  }

  return [defaultTriggers(hasGlobe), false]
}

/** appleKeyboard reports whether a keyboard ID has Apple's USB or Bluetooth vendor ID, so the keyboard has a Globe
 * key, mirroring the daemon's Keyboard.HasGlobe for a keyboard that is not connected. */
export function appleKeyboard(id: string): boolean {
  return id.startsWith('05ac:') || id.startsWith('004c:')
}

/** withGlobe puts the Globe key first among trigger keys, unless they have it. */
export function withGlobe(triggers: KeyboardTrigger[]): KeyboardTrigger[] {
  return switchesWithGlobe(triggers) ? triggers : [GLOBE_TRIGGER, ...triggers]
}

/** withoutGlobe drops the Globe key from trigger keys. */
export function withoutGlobe(triggers: KeyboardTrigger[]): KeyboardTrigger[] {
  return triggers.filter((trigger) => trigger.key !== GLOBE)
}

/** switchesWithGlobe reports whether trigger keys include the Globe key. */
export function switchesWithGlobe(triggers: KeyboardTrigger[]): boolean {
  return triggers.some((trigger) => trigger.key === GLOBE)
}

/** withKey sets a trigger's key, falling back to dedicated mode for a key that cannot use its mode. */
export function withKey(trigger: KeyboardTrigger, key: string): KeyboardTrigger {
  return { key, mode: allowedModes(key).includes(trigger.mode) ? trigger.mode : MODE_DEDICATED }
}
