import { type ReactElement, useId } from 'react'
import { Globe } from 'lucide-react'

import type { Config, Keyboard, KeyboardTrigger, KeyMapping } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { allowedModes, DEDICATED_USAGE, GLOBE, triggerFor, withKey } from '../../lib/trigger-keys'
import { Banner, Button, Disclosure, Section, Select } from '../../ui'
import { KeyPicker } from './KeyPicker'
import { describeTrigger, MODE_LABELS, modeHint } from './trigger-text'
import styles from './TriggerSection.module.css'

interface Props {
  config: Config
  errors: ConfigErrors
  keyboards: KeyboardsState
  onChange: (config: Config) => void
}

function mappingLabel(mapping: KeyMapping): string {
  return `${mapping.srcName} → ${mapping.dstName}`
}

/** describe renders the daemon's own mapping: silencing a dedicated key, or a remap. */
function describe(mapping: KeyMapping): string {
  return mapping.dst === DEDICATED_USAGE ? `${mapping.srcName} silenced` : mappingLabel(mapping)
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

/** TriggerSection lists the keyboards, each with its own trigger key; ones set here but not connected are folded away. */
export function TriggerSection({ config, errors, keyboards, onChange }: Props): ReactElement {
  const { trigger } = config

  // One entry per keyboard ID, then keyboards set here that are not connected.
  const connected = (keyboards.report?.keyboards ?? []).filter(
    (keyboard, index, all) => all.findIndex((other) => other.id === keyboard.id) === index,
  )
  const disconnected = Object.keys(trigger.keyboards).filter((id) => !connected.some((keyboard) => keyboard.id === id))

  const others = connected.filter((candidate) => candidate.otherMappings.length > 0)
  const systemMappings = keyboards.report?.systemMappings ?? []

  const setKeyboards = (next: Record<string, KeyboardTrigger>): void =>
    onChange({ ...config, trigger: { ...trigger, keyboards: next } })

  const resetKeyboard = (id: string): void =>
    setKeyboards(Object.fromEntries(Object.entries(trigger.keyboards).filter(([other]) => other !== id)))

  return (
    <Section
      icon={Globe}
      title="Trigger Key"
      description={errors.trigger.keyboards ?? "Keyboards you don't set use their Globe key, if they have one."}
    >
      {keyboards.error ? <Banner tone="error">{`Keyboards unavailable: ${keyboards.error}`}</Banner> : null}
      {!keyboards.report && !keyboards.error ? <p className={styles.empty}>Checking connected keyboards…</p> : null}
      {keyboards.report && connected.length === 0 ? <p className={styles.empty}>No keyboards connected.</p> : null}

      {connected.length > 0 ? (
        <div className={styles.group}>
          {connected.map((keyboard) => (
            <KeyboardRow
              key={keyboard.id}
              keyboard={keyboard}
              config={config}
              globeElsewhere={connected.some(
                (other) =>
                  other.id !== keyboard.id && switchesWithGlobe(triggerFor(trigger, other.id, other.hasGlobe)[0]),
              )}
              onChange={(next) => setKeyboards({ ...trigger.keyboards, [keyboard.id]: next })}
              onReset={() => resetKeyboard(keyboard.id)}
            />
          ))}
        </div>
      ) : null}

      {disconnected.length > 0 ? (
        <Disclosure title={`Other keyboards (${disconnected.length})`}>
          <div className={styles.group}>
            {disconnected.map((id) => (
              <div key={id} className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>{id}</span>
                  <span className={styles.details}>
                    {`${describeTrigger(trigger.keyboards[id])} · applies when connected`}
                  </span>
                </div>
                <Button variant="ghost" onClick={() => resetKeyboard(id)}>
                  Forget
                </Button>
              </div>
            ))}
          </div>
        </Disclosure>
      ) : null}

      {others.length > 0 || systemMappings.length > 0 ? (
        <Disclosure title="Mappings from other tools">
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
        </Disclosure>
      ) : null}
    </Section>
  )
}

interface RowProps {
  keyboard: Keyboard
  config: Config
  /** globeElsewhere reports that another connected keyboard switches with the Globe key. */
  globeElsewhere: boolean
  onChange: (trigger: KeyboardTrigger) => void
  /** onReset puts the keyboard back on its Globe key, or on no key without one. */
  onReset: () => void
}

/** KeyboardRow edits one connected keyboard's trigger key, and its mode once a key other than the Globe key is set. */
function KeyboardRow({ keyboard, config, globeElsewhere, onChange, onReset }: RowProps): ReactElement {
  const [current] = triggerFor(config.trigger, keyboard.id, keyboard.hasGlobe)
  const reset = keyboard.hasGlobe ? GLOBE : ''
  const name = keyboardName(keyboard)
  const modeId = useId()

  const details = [keyboard.builtIn ? 'Built-in' : keyboard.transport || 'Keyboard', keyboard.id]
  if (!sameTrigger(current, keyboard.trigger)) {
    details.push('applies once saved')
  } else if (keyboard.triggerMapping) {
    details.push(describe(keyboard.triggerMapping))
  }

  return (
    <div className={styles.keyboard} role="group" aria-label={name}>
      <div className={styles.row}>
        <div className={styles.text}>
          <span className={styles.name}>{name}</span>
          <span className={styles.details}>{details.join(' · ')}</span>
        </div>
        <div className={styles.control}>
          <KeyPicker
            value={current.key}
            keyboardId={keyboard.id}
            keyboardName={name}
            hasGlobe={keyboard.hasGlobe}
            onChange={(key) => (key === reset ? onReset() : onChange(withKey(current, key)))}
          />
        </div>
      </div>
      {current.key !== '' && current.key !== GLOBE ? (
        <div className={styles.mode}>
          <div className={styles.modeRow}>
            <label className={styles.modeLabel} htmlFor={modeId}>
              Mode
            </label>
            <div className={styles.modeSelect}>
              <Select
                id={modeId}
                value={current.mode}
                disabled={allowedModes(current.key).length < 2}
                options={allowedModes(current.key).map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
                onChange={(mode) => onChange({ ...current, mode })}
              />
            </div>
          </div>
          <p className={styles.hint}>{modeHint(current)}</p>
        </div>
      ) : null}
      {keyboard.hasGlobe && !switchesWithGlobe(current) && globeElsewhere ? (
        <div className={styles.mode}>
          <Banner tone="info">
            The Globe key still switches here: it cannot tell keyboards apart, and another keyboard switches with it.
          </Banner>
        </div>
      ) : null}
    </div>
  )
}
