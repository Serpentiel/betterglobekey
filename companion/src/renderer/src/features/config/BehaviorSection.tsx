import type { ReactElement } from 'react'
import { Keyboard } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import { REVERSE_MODIFIERS, type ConfigErrors } from '../../lib/config-schema'
import { capitalize } from '../../lib/text'
import {
  allowedModes,
  GLOBE,
  KEYBOARDS_ALL,
  KEYBOARDS_WITHOUT_GLOBE,
  MODE_DEDICATED,
  MODE_GLOBE,
  MODE_PASS_THROUGH,
  triggerName,
} from '../../lib/trigger-keys'
import { Field, Input, Section, Select, Setting, Switch } from '../../ui'
import { KeyPicker } from './KeyPicker'

const MODE_LABELS: Record<string, string> = {
  [MODE_GLOBE]: 'Act as Globe',
  [MODE_DEDICATED]: 'Only this key (Globe untouched)',
  [MODE_PASS_THROUGH]: "Keep the key's function",
}

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

  const { key, mode, keyboards } = config.trigger
  const name = triggerName(key)
  const modes = allowedModes(key)
  const remapped = key !== GLOBE && mode !== MODE_PASS_THROUGH

  const setTrigger = (patch: Partial<Config['trigger']>): void =>
    onChange({ ...config, trigger: { ...config.trigger, ...patch } })

  // A key that cannot use the current mode falls back to Globe mode, so picking
  // it never leaves the config invalid.
  const setTriggerKey = (next: string): void =>
    setTrigger({ key: next, mode: allowedModes(next).includes(mode) ? mode : MODE_GLOBE })

  const keyHint =
    key === GLOBE
      ? 'The Globe (fn) key on Apple keyboards. Record another key for a keyboard without one.'
      : remapped
        ? `${name} loses its normal function while betterglobekey runs, on the keyboards chosen below.`
        : `Pressing ${name} alone switches input sources; in combinations it still works as ${name}.`

  const modeHint =
    errors.trigger.mode ??
    (key === GLOBE
      ? 'Record a key other than Globe to choose how it works.'
      : {
          [MODE_GLOBE]: `${name} becomes a Globe key, with Globe's shortcuts (Globe+E, Globe+arrows, …). Globe keeps switching too.`,
          [MODE_DEDICATED]: `Only ${name} switches: it is remapped to an unused key (F20). Globe keeps just its macOS functions.`,
          [MODE_PASS_THROUGH]: `${name} keeps working as itself on every keyboard and switches only when pressed alone.`,
        }[mode])

  return (
    <Section icon={Keyboard} title="Behavior" description={`How the ${name} key responds to presses.`}>
      <Field label="Trigger key" hint={keyHint} error={errors.trigger.key}>
        {({ id }) => <KeyPicker id={id} value={key} remapped={remapped} onChange={setTriggerKey} />}
      </Field>
      <Field label="Trigger mode" hint={modeHint}>
        {({ id }) => (
          <Select
            id={id}
            value={mode}
            disabled={key === GLOBE}
            options={modes.map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
            onChange={(next) => setTrigger({ mode: next })}
          />
        )}
      </Field>
      <Field
        label="Remap on"
        hint={
          errors.trigger.keyboards ??
          (remapped
            ? 'Keyboards with their own Globe key are Apple keyboards. Change single keyboards under Keyboards.'
            : 'Only a remapped key applies to chosen keyboards.')
        }
      >
        {({ id }) => (
          <Select
            id={id}
            value={keyboards.applyTo}
            disabled={!remapped}
            options={[
              { value: KEYBOARDS_WITHOUT_GLOBE, label: 'Keyboards without Globe' },
              { value: KEYBOARDS_ALL, label: 'All keyboards' },
            ]}
            onChange={(applyTo) => setTrigger({ keyboards: { ...keyboards, applyTo } })}
          />
        )}
      </Field>

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
