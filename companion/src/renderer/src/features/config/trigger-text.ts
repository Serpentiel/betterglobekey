// Wording shared by the trigger key controls.

import { MODE_DEDICATED, MODE_PASS_THROUGH, triggerName } from '../../lib/trigger-keys'
import type { KeyboardTrigger } from '../../../../shared/types'

/** MODE_LABELS name the trigger modes in menus. */
export const MODE_LABELS: Record<string, string> = {
  [MODE_DEDICATED]: 'Switch input sources only',
  [MODE_PASS_THROUGH]: "Keep the key's function",
}

/** MODES_HINT explains the trigger modes and what they need. */
export const MODES_HINT =
  `A keyboard's own keys are read straight from the keyboard, which needs Input Monitoring. ` +
  `“${MODE_LABELS[MODE_DEDICATED]}” silences the key; “${MODE_LABELS[MODE_PASS_THROUGH]}” switches only ` +
  `when the key is pressed alone.`

/** describeTrigger names a trigger key and its mode, e.g. "Caps Lock (switch input sources only)". */
export function describeTrigger(trigger: KeyboardTrigger): string {
  return `${triggerName(trigger.key)} (${(MODE_LABELS[trigger.mode] ?? trigger.mode).toLowerCase()})`
}

/** describeTriggers names a keyboard's trigger keys, or says it has none. */
export function describeTriggers(triggers: KeyboardTrigger[]): string {
  return triggers.length === 0 ? 'No trigger key' : triggers.map(describeTrigger).join(', ')
}
