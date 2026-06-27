import { SquareMousePointer } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import type { ConfigErrors } from '../../lib/config-schema'
import { Field, Input, Section, Setting, Switch } from '../../ui'

interface Props {
  config: Config
  errors: ConfigErrors
  onChange: (config: Config) => void
}

export function HudSection({ config, errors, onChange }: Props): JSX.Element {
  const setHud = (patch: Partial<Config['hud']>): void => onChange({ ...config, hud: { ...config.hud, ...patch } })

  return (
    <Section
      icon={SquareMousePointer}
      title="On-screen HUD"
      description="The overlay shown when the input source changes."
    >
      <Setting
        title="Show the HUD"
        description="Display an overlay naming the new input source."
        control={
          <Switch checked={config.hud.enabled} onChange={(enabled) => setHud({ enabled })} label="Toggle the HUD" />
        }
      />
      <Field
        label="Visible duration"
        hint="How long the overlay stays before fading. A Go duration, e.g. 900ms."
        error={errors.hud.duration}
      >
        {({ id, invalid }) => (
          <Input
            id={id}
            invalid={invalid}
            value={config.hud.duration}
            placeholder="900ms"
            disabled={!config.hud.enabled}
            onChange={(event) => setHud({ duration: event.target.value })}
          />
        )}
      </Field>
      <Setting
        title="Show collection name"
        description="Include the collection as a subtitle in the overlay."
        control={
          <Switch
            checked={config.hud.showCollection}
            onChange={(showCollection) => setHud({ showCollection })}
            label="Toggle collection subtitle"
          />
        }
      />
    </Section>
  )
}
