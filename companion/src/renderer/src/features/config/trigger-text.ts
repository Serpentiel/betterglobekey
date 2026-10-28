// Wording shared by the trigger key controls.

import { GLOBE, MODE_DEDICATED, MODE_FN, MODE_PASS_THROUGH, triggerName } from '../../lib/trigger-keys'
import type { KeyboardTrigger } from '../../../../shared/types'

/** MODE_LABELS name the trigger modes in menus. */
export const MODE_LABELS: Record<string, string> = {
  [MODE_FN]: 'Act as fn',
  [MODE_DEDICATED]: 'Only this key switches',
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

  switch (trigger.mode) {
    case MODE_DEDICATED:
      return `Only ${name} switches: it is silenced and read straight from the keyboard, which needs Input Monitoring.`
    case MODE_PASS_THROUGH:
      return `${name} keeps working as itself and switches only when pressed alone. It is read straight from the keyboard, which needs Input Monitoring.`
    default:
      return `${name} switches input sources and works as fn in shortcuts (fn+arrows, fn+Delete, fn+E, …).`
  }
}

/** describeTrigger names a trigger and its mode, e.g. "Caps Lock, act as fn". */
export function describeTrigger(trigger: KeyboardTrigger): string {
  if (trigger.key === '') {
    return 'No trigger key'
  }

  if (trigger.key === GLOBE) {
    return 'Globe key'
  }

  return `${triggerName(trigger.key)}, ${(MODE_LABELS[trigger.mode] ?? trigger.mode).toLowerCase()}`
}
