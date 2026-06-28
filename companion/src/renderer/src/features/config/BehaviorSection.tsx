import type { ReactElement } from 'react'
import { Keyboard } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import { REVERSE_MODIFIERS, type ConfigErrors } from '../../lib/config-schema'
import { capitalize } from '../../lib/text'
import { Field, Input, Section, Select, Setting, Switch } from '../../ui'

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

  return (
    <Section icon={Keyboard} title="Behavior" description="How the Globe key responds to presses.">
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
            onChange={(event) => setReverse({ modifier: event.target.value })}
          />
        )}
      </Field>
    </Section>
  )
}
