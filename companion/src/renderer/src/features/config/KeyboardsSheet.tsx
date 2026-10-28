import { type ReactElement, useState } from 'react'

import type { Config, Keyboard, KeyboardTrigger, KeyMapping } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { allowedModes, DEDICATED_USAGE, GLOBE, triggerFor, withKey } from '../../lib/trigger-keys'
import { Banner, Button, Field, Select, Sheet } from '../../ui'
import { KeyPicker } from './KeyPicker'
import { describeTrigger, MODE_LABELS, modeHint } from './trigger-text'
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

function keyboardName(keyboard: Keyboard): string {
  return keyboard.name || keyboard.id
}

function sameTrigger(a: KeyboardTrigger, b: KeyboardTrigger): boolean {
  return a.key === b.key && (a.key === '' || a.key === GLOBE || a.mode === b.mode)
}

/** switchesWithGlobe reports whether a trigger makes the Globe key switch. */
function switchesWithGlobe(trigger: KeyboardTrigger): boolean {
  return trigger.key === GLOBE
}

/** KeyboardsSheet sets the trigger key of single keyboards, the way macOS's Modifier Keys sheet configures a chosen keyboard. */
export function KeyboardsSheet({ open, config, keyboards, onChange, onClose }: Props): ReactElement | null {
  const [chosen, setChosen] = useState<string | null>(null)
  const { trigger } = config

  // One entry per keyboard ID, then keyboards set here that are not connected.
  const connected = (keyboards.report?.keyboards ?? []).filter(
    (keyboard, index, all) => all.findIndex((other) => other.id === keyboard.id) === index,
  )
  const disconnected = Object.keys(trigger.keyboards).filter((id) => !connected.some((keyboard) => keyboard.id === id))
  const ids = [...connected.map((keyboard) => keyboard.id), ...disconnected]
  const selected = chosen !== null && ids.includes(chosen) ? chosen : ids[0]
  const keyboard = connected.find((candidate) => candidate.id === selected)

  const others = connected.filter((candidate) => candidate.otherMappings.length > 0)
  const systemMappings = keyboards.report?.systemMappings ?? []

  const setKeyboards = (next: Record<string, KeyboardTrigger>): void =>
    onChange({ ...config, trigger: { ...trigger, keyboards: next } })

  const resetKeyboard = (id: string): void =>
    setKeyboards(Object.fromEntries(Object.entries(trigger.keyboards).filter(([other]) => other !== id)))

  const options = [
    ...connected.map((candidate) => ({ value: candidate.id, label: keyboardName(candidate) })),
    ...disconnected.map((id) => ({ value: id, label: `${id} (not connected)` })),
  ]

  return (
    <Sheet
      open={open}
      title="Keyboards"
      description="Choose a keyboard to set its own trigger key. Changes apply when you save."
      onClose={onClose}
    >
      {keyboards.error ? <Banner tone="error">{`Keyboards unavailable: ${keyboards.error}`}</Banner> : null}
      {keyboards.report && ids.length === 0 ? <p className={styles.empty}>No keyboards connected.</p> : null}

      {selected ? (
        <>
          <Field label="Select keyboard">
            {({ id }) => <Select id={id} value={selected} options={options} onChange={setChosen} />}
          </Field>
          {keyboard ? (
            <KeyboardSettings
              keyboard={keyboard}
              config={config}
              globeElsewhere={connected.some(
                (other) =>
                  other.id !== keyboard.id && switchesWithGlobe(triggerFor(trigger, other.id, other.hasGlobe)[0]),
              )}
              onChange={(next) => setKeyboards({ ...trigger.keyboards, [keyboard.id]: next })}
              onReset={() => resetKeyboard(keyboard.id)}
            />
          ) : (
            <div className={styles.group}>
              <div className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>{describeTrigger(trigger.keyboards[selected])}</span>
                  <span className={styles.details}>Applies when the keyboard is connected.</span>
                </div>
                <Button variant="ghost" onClick={() => resetKeyboard(selected)}>
                  Forget
                </Button>
              </div>
            </div>
          )}
        </>
      ) : null}

      {others.length > 0 || systemMappings.length > 0 ? (
        <details className={styles.disclosure}>
          <summary className={styles.heading}>Mappings from other tools</summary>
          <div className={styles.group}>
            {others.map((candidate) => (
              <div key={candidate.id} className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>{keyboardName(candidate)}</span>
                  <span className={styles.details}>{candidate.otherMappings.map(mappingLabel).join(', ')}</span>
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

interface SettingsProps {
  keyboard: Keyboard
  config: Config
  /** globeElsewhere reports that another connected keyboard switches with the Globe key. */
  globeElsewhere: boolean
  onChange: (trigger: KeyboardTrigger) => void
  /** onReset puts the keyboard back on its Globe key, or on no key without one. */
  onReset: () => void
}

/** KeyboardSettings edits one connected keyboard's trigger key and mode. */
function KeyboardSettings({ keyboard, config, globeElsewhere, onChange, onReset }: SettingsProps): ReactElement {
  const [current] = triggerFor(config.trigger, keyboard.id, keyboard.hasGlobe)
  const reset = keyboard.hasGlobe ? GLOBE : ''
  const saved = sameTrigger(current, keyboard.trigger)

  let status = 'No key is remapped.'
  if (!saved) {
    status = 'Applies once saved.'
  } else if (keyboard.triggerMapping) {
    status = `${describe(keyboard.triggerMapping)}.`
  }

  return (
    <>
      <p className={styles.details}>{details(keyboard)}</p>
      <Field label="Trigger key">
        {({ id }) => (
          <KeyPicker
            id={id}
            value={current.key}
            keyboardId={keyboard.id}
            keyboardName={keyboardName(keyboard)}
            hasGlobe={keyboard.hasGlobe}
            onChange={(key) => (key === reset ? onReset() : onChange(withKey(current, key)))}
          />
        )}
      </Field>
      <Field label="Trigger mode" hint={modeHint(current)}>
        {({ id }) => (
          <Select
            id={id}
            value={current.mode}
            disabled={allowedModes(current.key).length < 2 || current.key === '' || current.key === GLOBE}
            options={allowedModes(current.key).map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
            onChange={(mode) => onChange({ ...current, mode })}
          />
        )}
      </Field>
      <div className={styles.footer}>
        <span className={styles.status}>{status}</span>
      </div>
      {keyboard.hasGlobe && !switchesWithGlobe(current) && globeElsewhere ? (
        <Banner tone="info">
          The Globe key still switches here: it cannot tell keyboards apart, and another keyboard switches with it.
        </Banner>
      ) : null}
    </>
  )
}
