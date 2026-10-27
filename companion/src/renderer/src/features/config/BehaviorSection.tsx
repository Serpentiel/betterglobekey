import type { ReactElement } from 'react'
import { Keyboard } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import { REVERSE_MODIFIERS, type ConfigErrors } from '../../lib/config-schema'
import { capitalize } from '../../lib/text'
import { GLOBE, triggerKey, triggerName } from '../../lib/trigger-keys'
import { Field, Input, Section, Select, Setting, Switch } from '../../ui'
import { KeyPicker } from './KeyPicker'

interface Props {
  config: Config
  errors: ConfigErrors
  onChange: (config: Config) => void
}

export function BehaviorSection({ config, errors, onChange }: Props): ReactElement {
  const setDoublePress = (patch: Partial<Config['doublePress']>): void =>
    onChange({ ...config, doublePress: { ...config.doublePress, ...patch } })

  const setReverse = (patch: Partial<Config['reverse']>): void =>
    onChange({ ...config, reverse: { ...config.reverse, ...patch } })

  const { key, passThrough } = config.trigger
  const name = triggerName(key)
  const canPassThrough = triggerKey(key)?.passThrough ?? false
  const remapped = key !== GLOBE && !passThrough

  // A key that can only be remapped drops pass-through, so picking it never
  // leaves the config invalid.
  const setTriggerKey = (next: string): void =>
    onChange({
      ...config,
      trigger: { key: next, passThrough: passThrough && (triggerKey(next)?.passThrough ?? false) },
    })

  const keyHint =
    key === GLOBE
      ? 'The Globe (fn) key on Apple keyboards. Record another key for a keyboard without one.'
      : remapped
        ? `${name} becomes the Globe key while betterglobekey runs, and loses its normal function.`
        : `Pressing ${name} alone switches input sources; in combinations it still works as ${name}.`

  const passThroughHint =
    errors.trigger.passThrough ??
    (key === GLOBE
      ? 'Globe always keeps its normal function.'
      : canPassThrough
        ? `${name} keeps working as itself instead of becoming the Globe key.`
        : `${name} can only be remapped to Globe.`)

  return (
    <Section icon={Keyboard} title="Behavior" description={`How the ${name} key responds to presses.`}>
      <Field label="Trigger key" hint={keyHint} error={errors.trigger.key}>
        {({ id }) => <KeyPicker id={id} value={key} remapped={remapped} onChange={setTriggerKey} />}
      </Field>
      <Setting
        title="Keep the key's normal function"
        description={passThroughHint}
        control={
          <Switch
            checked={passThrough && canPassThrough}
            disabled={!canPassThrough}
            onChange={(enabled) => onChange({ ...config, trigger: { key, passThrough: enabled } })}
            label="Toggle keeping the key's normal function"
          />
        }
      />

      <Setting
        title="Double press to switch collections"
        description="A quick second press cycles to the next collection."
        control={
          <Switch
            checked={config.doublePress.enabled}
            onChange={(enabled) => setDoublePress({ enabled })}
            label="Toggle double press"
          />
        }
      />
      <Field
        label="Double-press maximum delay"
        hint="The longest gap between two presses still counted as a double press. A Go duration, e.g. 250ms."
        error={errors.doublePress.maximumDelay}
      >
        {({ id, invalid }) => (
          <Input
            id={id}
            invalid={invalid}
            value={config.doublePress.maximumDelay}
            placeholder="250ms"
            onChange={(event) => setDoublePress({ maximumDelay: event.target.value })}
          />
        )}
      </Field>

      <Setting
        title="Reverse modifier"
        description="Hold a modifier while pressing to go back instead of forward."
        control={
          <Switch
            checked={config.reverse.enabled}
            onChange={(enabled) => setReverse({ enabled })}
            label="Toggle reverse modifier"
          />
        }
      />
      <Field label="Modifier key" error={errors.reverse.modifier}>
        {({ id }) => (
          <Select
            id={id}
            value={config.reverse.modifier}
            disabled={!config.reverse.enabled}
            options={REVERSE_MODIFIERS.map((modifier) => ({ value: modifier, label: capitalize(modifier) }))}
            onChange={(modifier) => setReverse({ modifier })}
          />
        )}
      </Field>
    </Section>
  )
}
