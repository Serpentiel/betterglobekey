import type { ReactElement } from 'react'
import { clsx } from 'clsx'

import type { Config, Keyboard, KeyMapping } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import {
  includesKeyboard,
  KEYBOARDS_ALL,
  DEDICATED_USAGE,
  KEYBOARDS_WITHOUT_GLOBE,
  MODE_DEDICATED,
  remappedTrigger,
  triggerName,
} from '../../lib/trigger-keys'
import { Banner, Button, Select, Sheet, Switch } from '../../ui'
import styles from './KeyboardsSheet.module.css'

interface Props {
  open: boolean
  config: Config
  keyboards: KeyboardsState
  onChange: (config: Config) => void
  onClose: () => void
}

function mappingLabel(mapping: KeyMapping): string {
  return `${mapping.srcName} → ${mapping.dstName}`
}

/** describe renders the daemon's own mapping: silencing a dedicated key, or a remap. */
function describe(mapping: KeyMapping): string {
  return mapping.dst === DEDICATED_USAGE ? `${mapping.srcName} silenced` : mappingLabel(mapping)
}

function details(keyboard: Keyboard): string {
  const parts = [keyboard.builtIn ? 'Built-in' : keyboard.transport || 'Keyboard', keyboard.id]

  if (keyboard.hasGlobe) {
    parts.push('has Globe key')
  }

  return parts.join(' · ')
}

/** keyboardName names a keyboard, falling back to its ID. */
function keyboardName(keyboard: Keyboard): string {
  return keyboard.name || keyboard.id
}

/** KeyboardsSheet chooses the keyboards the trigger key is remapped on, the way
 * macOS's Modifier Keys sheet picks a keyboard to configure. */
export function KeyboardsSheet({ open, config, keyboards, onChange, onClose }: Props): ReactElement | null {
  const { key, mode } = config.trigger
  const scope = config.trigger.keyboards
  const remapped = remappedTrigger(config.trigger)
  const name = triggerName(key)
  const connected = keyboards.report?.keyboards ?? []
  const disconnected = Object.keys(scope.overrides).filter((id) => !connected.some((keyboard) => keyboard.id === id))
  const others = connected.filter((keyboard) => keyboard.otherMappings.length > 0)
  const systemMappings = keyboards.report?.systemMappings ?? []

  const setScope = (patch: Partial<Config['trigger']['keyboards']>): void =>
    onChange({ ...config, trigger: { ...config.trigger, keyboards: { ...scope, ...patch } } })

  const forget = (id: string): void =>
    setScope({ overrides: Object.fromEntries(Object.entries(scope.overrides).filter(([other]) => other !== id)) })

  // Store an override only where it differs from the default.
  const include = (keyboard: Keyboard, included: boolean): void => {
    if (includesKeyboard({ ...scope, overrides: {} }, keyboard.id, keyboard.hasGlobe) === included) {
      forget(keyboard.id)
    } else {
      setScope({ overrides: { ...scope.overrides, [keyboard.id]: included } })
    }
  }

  const inScope = (keyboard: Keyboard): boolean => remapped && includesKeyboard(scope, keyboard.id, keyboard.hasGlobe)

  const status = (keyboard: Keyboard): string => {
    const included = inScope(keyboard)

    if (included && keyboard.triggerMapping) {
      return describe(keyboard.triggerMapping)
    }

    if (included) {
      return mode === MODE_DEDICATED ? `${name} silenced once saved` : `${name} → fn once saved`
    }

    if (keyboard.triggerMapping) {
      return `${describe(keyboard.triggerMapping)}, undone once saved`
    }

    return remapped ? `${name} unchanged` : 'Not remapped'
  }

  return (
    <Sheet
      open={open}
      title="Keyboards"
      description={
        remapped
          ? `Choose the keyboards ${name} is remapped on. Changes apply when you save.`
          : 'Record a trigger key that is remapped to choose the keyboards it applies to.'
      }
      onClose={onClose}
    >
      {keyboards.error ? <Banner tone="error">{`Keyboards unavailable: ${keyboards.error}`}</Banner> : null}
      {mode === MODE_DEDICATED && remapped && keyboards.report && !connected.some(inScope) ? (
        <Banner tone="info">
          No connected keyboard remaps the trigger key, and the Globe key does not switch in this mode. Turn on a
          keyboard below.
        </Banner>
      ) : null}

      <div className={styles.group}>
        {keyboards.report && connected.length === 0 ? <p className={styles.empty}>No keyboards connected.</p> : null}
        {connected.map((keyboard) => (
          <div key={keyboard.id} className={styles.row}>
            <div className={styles.text}>
              <span className={styles.name}>{keyboardName(keyboard)}</span>
              <span className={styles.details}>{details(keyboard)}</span>
              <span className={clsx(styles.status, inScope(keyboard) && styles.active)}>{status(keyboard)}</span>
            </div>
            {scope.overrides[keyboard.id] !== undefined ? (
              <Button variant="ghost" onClick={() => forget(keyboard.id)}>
                Reset
              </Button>
            ) : null}
            <Switch
              checked={inScope(keyboard)}
              disabled={!remapped}
              onChange={(next) => include(keyboard, next)}
              label={`Remap ${name} on ${keyboardName(keyboard)}`}
            />
          </div>
        ))}
        <div className={styles.row}>
          <div className={styles.text}>
            <span className={styles.name}>Other keyboards</span>
            <span className={styles.details}>Keyboards connected later, and those not changed above.</span>
          </div>
          <Select
            id="keyboards-default"
            value={scope.applyTo}
            disabled={!remapped}
            options={[
              { value: KEYBOARDS_WITHOUT_GLOBE, label: 'Remap if no Globe key' },
              { value: KEYBOARDS_ALL, label: 'Always remap' },
            ]}
            onChange={(applyTo) => setScope({ applyTo })}
          />
        </div>
      </div>

      {disconnected.length > 0 ? (
        <section>
          <h3 className={styles.heading}>Not connected</h3>
          <div className={styles.group}>
            {disconnected.map((id) => (
              <div key={id} className={clsx(styles.row, styles.dimmed)}>
                <div className={styles.text}>
                  <span className={styles.name}>{id}</span>
                  <span className={styles.details}>{scope.overrides[id] ? 'Always remapped' : 'Never remapped'}</span>
                </div>
                <Button variant="ghost" onClick={() => forget(id)}>
                  Forget
                </Button>
              </div>
            ))}
          </div>
        </section>
      ) : null}

      {others.length > 0 || systemMappings.length > 0 ? (
        <details className={styles.disclosure}>
          <summary className={styles.heading}>Mappings from other tools</summary>
          <div className={styles.group}>
            {others.map((keyboard) => (
              <div key={keyboard.id} className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>{keyboardName(keyboard)}</span>
                  <span className={styles.details}>{keyboard.otherMappings.map(mappingLabel).join(', ')}</span>
                </div>
              </div>
            ))}
            {systemMappings.length > 0 ? (
              <div className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>Every keyboard</span>
                  <span className={styles.details}>{systemMappings.map(mappingLabel).join(', ')}</span>
                </div>
              </div>
            ) : null}
          </div>
        </details>
      ) : null}
    </Sheet>
  )
}
