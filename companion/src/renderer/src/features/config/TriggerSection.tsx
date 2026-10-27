import { type ReactElement, useState } from 'react'
import { Globe } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import {
  allowedModes,
  GLOBE,
  includesKeyboard,
  MODE_DEDICATED,
  MODE_FN,
  MODE_PASS_THROUGH,
  remappedTrigger,
  triggerName,
} from '../../lib/trigger-keys'
import { Button, Field, Section, Select, Setting } from '../../ui'
import { KeyboardsSheet } from './KeyboardsSheet'
import { KeyPicker } from './KeyPicker'

interface Props {
  config: Config
  errors: ConfigErrors
  keyboards: KeyboardsState
  onChange: (config: Config) => void
}

const MODE_LABELS: Record<string, string> = {
  [MODE_FN]: 'Act as fn',
  [MODE_DEDICATED]: 'Only this key switches',
  [MODE_PASS_THROUGH]: "Keep the key's function",
}

/** keyboardsSummary describes the keyboards row: where the key is remapped. */
function keyboardsSummary(config: Config, { report, error }: KeyboardsState): string {
  if (error) {
    return `Unavailable: ${error}`
  }

  if (!report) {
    return 'Checking connected keyboards…'
  }

  const total = report.keyboards.length
  const connected = `${total} connected`

  if (!remappedTrigger(config.trigger)) {
    return total === 1 ? '1 keyboard connected.' : `${total} keyboards connected.`
  }

  const remapped = report.keyboards.filter((keyboard) =>
    includesKeyboard(config.trigger.keyboards, keyboard.id, keyboard.hasGlobe),
  ).length

  if (remapped === 0 && config.trigger.mode === MODE_DEDICATED) {
    return `Not remapped on any keyboard (${connected}), so nothing switches.`
  }

  return `Remapped on ${remapped} of ${total} ${total === 1 ? 'keyboard' : 'keyboards'}.`
}

export function TriggerSection({ config, errors, keyboards, onChange }: Props): ReactElement {
  const [sheetOpen, setSheetOpen] = useState(false)
  const { key, mode } = config.trigger
  const name = triggerName(key)
  const remapped = remappedTrigger(config.trigger)

  const setTrigger = (patch: Partial<Config['trigger']>): void =>
    onChange({ ...config, trigger: { ...config.trigger, ...patch } })

  // A key that cannot use the current mode falls back to fn mode, so picking it
  // never leaves the config invalid.
  const setTriggerKey = (next: string): void =>
    setTrigger({ key: next, mode: allowedModes(next).includes(mode) ? mode : MODE_FN })

  const keyHint =
    key === GLOBE
      ? 'The Globe (fn) key on Apple keyboards. Record another key for a keyboard without one.'
      : remapped
        ? `${name} loses its normal function while betterglobekey runs, on the chosen keyboards.`
        : `Pressing ${name} alone switches input sources; in combinations it still works as ${name}.`

  const modeHint =
    errors.trigger.mode ??
    (key === GLOBE
      ? 'Record a key other than Globe to choose how it works.'
      : {
          [MODE_FN]: `${name} becomes an fn key, with the fn shortcuts (fn+arrows, fn+Delete, fn+E, …). Globe keeps switching too.`,
          [MODE_DEDICATED]: `Only ${name} switches: it is remapped to an unused key (F20). Globe (fn) keeps only its macOS functions.`,
          [MODE_PASS_THROUGH]: `${name} keeps working as itself on every keyboard and switches only when pressed alone.`,
        }[mode])

  return (
    <Section icon={Globe} title="Trigger Key" description="The key that switches input sources.">
      <Field label="Trigger key" hint={keyHint} error={errors.trigger.key}>
        {({ id }) => (
          <KeyPicker
            id={id}
            value={key}
            remapped={remapped}
            dedicated={remapped && mode === MODE_DEDICATED}
            onChange={setTriggerKey}
          />
        )}
      </Field>
      <Field label="Trigger mode" hint={modeHint}>
        {({ id }) => (
          <Select
            id={id}
            value={mode}
            disabled={key === GLOBE}
            options={allowedModes(key).map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
            onChange={(next) => setTrigger({ mode: next })}
          />
        )}
      </Field>
      <Setting
        title="Keyboards"
        description={errors.trigger.keyboards ?? keyboardsSummary(config, keyboards)}
        control={<Button onClick={() => setSheetOpen(true)}>Keyboards…</Button>}
      />
      <KeyboardsSheet
        open={sheetOpen}
        config={config}
        keyboards={keyboards}
        onChange={onChange}
        onClose={() => setSheetOpen(false)}
      />
    </Section>
  )
}
