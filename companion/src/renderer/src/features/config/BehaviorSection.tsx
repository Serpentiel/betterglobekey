import { Keyboard } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import type { ConfigErrors } from '../../lib/config-schema'
import { Field, Input, Section, Setting, Switch } from '../../ui'

interface Props {
  config: Config
  errors: ConfigErrors
  onChange: (config: Config) => void
}

export function BehaviorSection({ config, errors, onChange }: Props): JSX.Element {
  return (
    <Section icon={Keyboard} title="Behavior" description="How the Globe key behaves.">
      <Setting
        title="On-screen HUD"
        description="Show an overlay naming the input source when it changes."
        control={
          <Switch checked={config.hud} onChange={(hud) => onChange({ ...config, hud })} label="Toggle the HUD" />
        }
      />
      <Field
        label="Double-press maximum delay"
        hint="The longest gap between two Globe presses still counted as a double press. A Go duration, e.g. 250ms."
        error={errors.doublePressMaxDelay}
      >
        {({ id, invalid }) => (
          <Input
            id={id}
            invalid={invalid}
            value={config.doublePressMaxDelay}
            placeholder="250ms"
            onChange={(event) => onChange({ ...config, doublePressMaxDelay: event.target.value })}
          />
        )}
      </Field>
    </Section>
  )
}
