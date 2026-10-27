import { type ReactElement, useState } from 'react'
import { Globe } from 'lucide-react'

import type { Config, KeyboardTrigger } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { allowedModes, GLOBE, withKey } from '../../lib/trigger-keys'
import { Button, Field, Section, Select, Setting } from '../../ui'
import { KeyboardsSheet } from './KeyboardsSheet'
import { KeyPicker } from './KeyPicker'
import { MODE_LABELS, modeHint } from './trigger-text'

interface Props {
  config: Config
  errors: ConfigErrors
  keyboards: KeyboardsState
  onChange: (config: Config) => void
}

/** keyboardsSummary describes the keyboards row: how many there are, and how many set their own trigger. */
function keyboardsSummary(config: Config, { report, error }: KeyboardsState): string {
  if (error) {
    return `Unavailable: ${error}`
  }

  if (!report) {
    return 'Checking connected keyboards…'
  }

  const total = report.keyboards.length
  const listed = Object.keys(config.trigger.keyboards).length
  const connected = total === 1 ? '1 keyboard connected' : `${total} keyboards connected`

  return listed === 0 ? `${connected}.` : `${connected}; ${listed} set their own trigger key.`
}

export function TriggerSection({ config, errors, keyboards, onChange }: Props): ReactElement {
  const [sheetOpen, setSheetOpen] = useState(false)
  const { trigger } = config
  const defaultTrigger: KeyboardTrigger = { key: trigger.key, mode: trigger.mode }

  const setDefault = (next: KeyboardTrigger): void => onChange({ ...config, trigger: { ...trigger, ...next } })

  return (
    <Section icon={Globe} title="Trigger Key" description="The key that switches input sources, on each keyboard.">
      <Field
        label="Keyboards without a Globe key"
        hint={
          trigger.key === GLOBE
            ? 'Keyboards with a Globe key use it. Record a key for keyboards without one.'
            : 'Keyboards with a Globe key keep using it. Choose single keyboards under Keyboards.'
        }
        error={errors.trigger.key}
      >
        {({ id }) => (
          <KeyPicker id={id} value={trigger.key} onChange={(key) => setDefault(withKey(defaultTrigger, key))} />
        )}
      </Field>
      <Field label="Trigger mode" hint={errors.trigger.mode ?? modeHint(defaultTrigger)}>
        {({ id }) => (
          <Select
            id={id}
            value={trigger.mode}
            disabled={trigger.key === GLOBE}
            options={allowedModes(trigger.key).map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
            onChange={(mode) => setDefault({ ...defaultTrigger, mode })}
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
