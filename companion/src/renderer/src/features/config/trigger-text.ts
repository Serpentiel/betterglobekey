// Wording shared by the trigger key controls.

import { GLOBE, MODE_DEDICATED, MODE_PASS_THROUGH, triggerName } from '../../lib/trigger-keys'
import type { KeyboardTrigger } from '../../../../shared/types'

/** MODE_LABELS name the trigger modes in menus. */
export const MODE_LABELS: Record<string, string> = {
  [MODE_DEDICATED]: 'Switch input sources only',
  [MODE_PASS_THROUGH]: "Keep the key's function",
}

/** modeHint explains what a trigger's mode does. */
export function modeHint(trigger: KeyboardTrigger): string {
  const name = triggerName(trigger.key)

  if (trigger.key === '') {
    return 'Record a key to switch input sources from this keyboard.'
  }

  if (trigger.key === GLOBE) {
    return 'Record a key other than the Globe key to choose how it works.'
  }

  return trigger.mode === MODE_PASS_THROUGH
    ? `${name} keeps working as itself and switches only when pressed alone. It is read straight from the keyboard, which needs Input Monitoring.`
    : `${name} only switches input sources: it is silenced and read straight from the keyboard, which needs Input Monitoring.`
}

/** describeTrigger names a trigger and its mode, e.g. "Caps Lock, switch input sources only". */
export function describeTrigger(trigger: KeyboardTrigger): string {
  if (trigger.key === '') {
    return 'No trigger key'
  }

  if (trigger.key === GLOBE) {
    return 'Globe key'
  }

  return `${triggerName(trigger.key)}, ${(MODE_LABELS[trigger.mode] ?? trigger.mode).toLowerCase()}`
}
