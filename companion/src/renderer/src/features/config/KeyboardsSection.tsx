import type { ReactElement } from 'react'
import { Usb } from 'lucide-react'
import { clsx } from 'clsx'

import type { Config, Keyboard, KeyMapping } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { GLOBE, includesKeyboard, MODE_DEDICATED, MODE_PASS_THROUGH, triggerName } from '../../lib/trigger-keys'
import { Banner, Button, Card, Section, Stack, Switch } from '../../ui'
import styles from './KeyboardsSection.module.css'

interface Props {
  config: Config
  keyboards: KeyboardsState
  onChange: (config: Config) => void
}

/** mappingLabel renders a key mapping as "Caps Lock → Globe". */
function mappingLabel(mapping: KeyMapping): string {
  return `${mapping.srcName} → ${mapping.dstName}`
}

function details(keyboard: Keyboard): string {
  const parts = [keyboard.id]

  if (keyboard.builtIn) {
    parts.push('built-in')
  } else if (keyboard.transport) {
    parts.push(keyboard.transport)
  }

  if (keyboard.hasGlobe) {
    parts.push('has Globe')
  }

  return parts.join(' · ')
}

export function KeyboardsSection({ config, keyboards, onChange }: Props): ReactElement {
  const { key, mode } = config.trigger
  const scope = config.trigger.keyboards
  const remapped = key !== GLOBE && mode !== MODE_PASS_THROUGH
  const connected = keyboards.report?.keyboards ?? []
  const systemMappings = keyboards.report?.systemMappings ?? []
  const disconnected = Object.keys(scope.overrides).filter((id) => !connected.some((keyboard) => keyboard.id === id))

  const setOverrides = (overrides: Record<string, boolean>): void =>
    onChange({ ...config, trigger: { ...config.trigger, keyboards: { ...scope, overrides } } })

  const forget = (id: string): void => {
    setOverrides(Object.fromEntries(Object.entries(scope.overrides).filter(([other]) => other !== id)))
  }

  // Store an override only where it differs from the default scope.
  const include = (keyboard: Keyboard, included: boolean): void => {
    if (includesKeyboard({ ...scope, overrides: {} }, keyboard.id, keyboard.hasGlobe) === included) {
      forget(keyboard.id)
    } else {
      setOverrides({ ...scope.overrides, [keyboard.id]: included })
    }
  }

  const inScope = (keyboard: Keyboard): boolean => remapped && includesKeyboard(scope, keyboard.id, keyboard.hasGlobe)
  const description = remapped
    ? `Where ${triggerName(key)} is remapped. Changes apply when you save.`
    : 'The connected keyboards and their key mappings.'

  return (
    <Section icon={Usb} title="Keyboards" description={description}>
      {keyboards.error ? <Banner tone="error">{`Keyboards unavailable: ${keyboards.error}`}</Banner> : null}
      {mode === MODE_DEDICATED && remapped && keyboards.report && !connected.some(inScope) ? (
        <Banner tone="info">
          No connected keyboard remaps the trigger key, and Globe does not switch in this mode. Include a keyboard
          below.
        </Banner>
      ) : null}
      {keyboards.report && connected.length === 0 ? <p className={styles.empty}>No keyboards connected.</p> : null}

      {connected.map((keyboard) => {
        const included = inScope(keyboard)
        const overridden = scope.overrides[keyboard.id] !== undefined
        const pending = included && !keyboard.triggerMapping
        const removing = !included && Boolean(keyboard.triggerMapping)

        return (
          <Card key={keyboard.id} className={styles.keyboard}>
            <Stack gap={8}>
              <div className={styles.header}>
                <div className={styles.identity}>
                  <span className={styles.name}>{keyboard.name || 'Unnamed keyboard'}</span>
                  <span className={styles.details}>{details(keyboard)}</span>
                </div>
                {overridden ? (
                  <Button variant="ghost" onClick={() => forget(keyboard.id)}>
                    Reset
                  </Button>
                ) : null}
                <Switch
                  checked={included}
                  disabled={!remapped}
                  onChange={(next) => include(keyboard, next)}
                  label={`Remap the trigger key on ${keyboard.name || keyboard.id}`}
                />
              </div>
              <ul className={styles.mappings}>
                {keyboard.triggerMapping ? (
                  <li className={clsx(styles.mapping, styles.ours, removing && styles.stale)}>
                    {mappingLabel(keyboard.triggerMapping)}
                    <span className={styles.source}>{removing ? 'removed when saved' : 'betterglobekey'}</span>
                  </li>
                ) : null}
                {pending ? (
                  <li className={clsx(styles.mapping, styles.pending)}>
                    {`${triggerName(key)} → ${mode === MODE_DEDICATED ? 'F20' : 'Globe'}`}
                    <span className={styles.source}>applied once saved</span>
                  </li>
                ) : null}
                {keyboard.otherMappings.map((mapping) => (
                  <li key={`${mapping.src}-${mapping.dst}`} className={styles.mapping}>
                    {mappingLabel(mapping)}
                    <span className={styles.source}>other tool</span>
                  </li>
                ))}
                {!keyboard.triggerMapping && !pending && keyboard.otherMappings.length === 0 ? (
                  <li className={styles.none}>No key mappings</li>
                ) : null}
              </ul>
            </Stack>
          </Card>
        )
      })}

      {disconnected.map((id) => (
        <Card key={id} className={clsx(styles.keyboard, styles.disconnected)}>
          <div className={styles.header}>
            <div className={styles.identity}>
              <span className={styles.name}>{`${id} (not connected)`}</span>
              <span className={styles.details}>{scope.overrides[id] ? 'Always remapped' : 'Never remapped'}</span>
            </div>
            <Button variant="ghost" onClick={() => forget(id)}>
              Forget
            </Button>
          </div>
        </Card>
      ))}

      {systemMappings.length > 0 ? (
        <Card className={styles.keyboard}>
          <Stack gap={8}>
            <div className={styles.identity}>
              <span className={styles.name}>System-wide</span>
              <span className={styles.details}>Set with hidutil for every keyboard</span>
            </div>
            <ul className={styles.mappings}>
              {systemMappings.map((mapping) => (
                <li key={`${mapping.src}-${mapping.dst}`} className={styles.mapping}>
                  {mappingLabel(mapping)}
                  <span className={styles.source}>other tool</span>
                </li>
              ))}
            </ul>
          </Stack>
        </Card>
      ) : null}
    </Section>
  )
}
