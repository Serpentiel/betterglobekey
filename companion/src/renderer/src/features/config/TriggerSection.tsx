import { type ReactElement, useId } from 'react'
import { Globe } from 'lucide-react'

import type { Config, Keyboard, KeyboardTrigger, KeyMapping } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { allowedModes, DEDICATED_USAGE, triggerName, triggersFor } from '../../lib/trigger-keys'
import { Banner, Button, Disclosure, Section, Select, Setting, Switch } from '../../ui'
import { KeyPicker } from './KeyPicker'
import { describeTriggers, MODE_LABELS, MODES_HINT } from './trigger-text'
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

/** sameTriggers reports whether two lists hold the same keys in the same modes. */
function sameTriggers(a: KeyboardTrigger[], b: KeyboardTrigger[]): boolean {
  return (
    a.length === b.length && a.every((trigger, index) => trigger.key === b[index].key && trigger.mode === b[index].mode)
  )
}

/**
 * TriggerSection turns the Globe key on or off for every keyboard at once, and lists the keyboards, each with its own
 * trigger keys; ones set here but not connected are folded away.
 */
export function TriggerSection({ config, errors, keyboards, onChange }: Props): ReactElement {
  const { trigger } = config

  // One entry per keyboard ID, then keyboards set here that are not connected.
  const connected = (keyboards.report?.keyboards ?? []).filter(
    (keyboard, index, all) => all.findIndex((other) => other.id === keyboard.id) === index,
  )
  const disconnected = Object.keys(trigger.keyboards).filter((id) => !connected.some((keyboard) => keyboard.id === id))

  const others = connected.filter((candidate) => candidate.otherMappings.length > 0)
  const systemMappings = keyboards.report?.systemMappings ?? []

  const setKeyboards = (next: Record<string, KeyboardTrigger[]>): void =>
    onChange({ ...config, trigger: { ...trigger, keyboards: next } })

  const resetKeyboard = (id: string): void =>
    setKeyboards(Object.fromEntries(Object.entries(trigger.keyboards).filter(([other]) => other !== id)))

  const setGlobe = (globe: boolean): void => onChange({ ...config, trigger: { ...trigger, globe } })

  // A keyboard without keys of its own is unlisted.
  const setTriggers = (id: string, next: KeyboardTrigger[]): void =>
    next.length === 0 ? resetKeyboard(id) : setKeyboards({ ...trigger.keyboards, [id]: next })

  return (
    <Section
      icon={Globe}
      title="Trigger Keys"
      description={errors.trigger.keyboards ?? 'The keys that switch input sources, on each keyboard.'}
    >
      <Setting
        title="Globe key"
        description="Switches input sources on every keyboard that has one. macOS can't tell which keyboard it came from."
        control={<Switch checked={trigger.globe} onChange={setGlobe} label="Toggle the Globe key" />}
      />
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
              onChange={(next) => setTriggers(keyboard.id, next)}
              onGlobe={trigger.globe ? undefined : () => setGlobe(true)}
            />
          ))}
        </div>
      ) : null}
      {connected.some((keyboard) => triggersFor(trigger, keyboard.id)[0].length > 0) ? (
        <p className={styles.hint}>{MODES_HINT}</p>
      ) : null}

      {disconnected.length > 0 ? (
        <Disclosure title={`Other keyboards (${disconnected.length})`}>
          <div className={styles.group}>
            {disconnected.map((id) => (
              <div key={id} className={styles.row}>
                <div className={styles.text}>
                  <span className={styles.name}>{id}</span>
                  <span className={styles.details}>
                    {`${describeTriggers(trigger.keyboards[id])} · applies when connected`}
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
  onChange: (triggers: KeyboardTrigger[]) => void
  /** onGlobe turns the Globe key on, when it is recorded while off. */
  onGlobe?: () => void
}

/** KeyboardRow edits one connected keyboard's own trigger keys and the mode of each; it shows the Globe key read-only. */
function KeyboardRow({ keyboard, config, onChange, onGlobe }: RowProps): ReactElement {
  const [current] = triggersFor(config.trigger, keyboard.id)
  const name = keyboardName(keyboard)

  const details = [keyboard.builtIn ? 'Built-in' : keyboard.transport || 'Keyboard', keyboard.id]
  if (!sameTriggers(current, keyboard.triggers)) {
    details.push('applies once saved')
  } else {
    details.push(...keyboard.triggerMappings.map(describe))
  }

  return (
    <div className={styles.keyboard} role="group" aria-label={name}>
      <div className={styles.head}>
        <div className={styles.text}>
          <span className={styles.name}>{name}</span>
          <span className={styles.details}>{details.join(' · ')}</span>
        </div>
        <KeyPicker
          triggers={current}
          keyboardId={keyboard.id}
          keyboardName={name}
          globe={keyboard.hasGlobe && config.trigger.globe}
          onGlobe={onGlobe}
          onChange={onChange}
        />
      </div>
      {current.length > 0 ? (
        <div className={styles.mode}>
          {current.map((trigger) => (
            <ModeRow
              key={trigger.key}
              trigger={trigger}
              onChange={(mode) =>
                onChange(current.map((other) => (other.key === trigger.key ? { ...other, mode } : other)))
              }
            />
          ))}
        </div>
      ) : null}
    </div>
  )
}

interface ModeRowProps {
  trigger: KeyboardTrigger
  onChange: (mode: string) => void
}

/** ModeRow chooses how one of a keyboard's own trigger keys works. */
function ModeRow({ trigger, onChange }: ModeRowProps): ReactElement {
  const id = useId()
  const modes = allowedModes(trigger.key)

  return (
    <div className={styles.modeRow}>
      <label className={styles.modeLabel} htmlFor={id}>
        {triggerName(trigger.key)}
      </label>
      <div className={styles.modeSelect}>
        <Select
          id={id}
          value={trigger.mode}
          disabled={modes.length < 2}
          options={modes.map((value) => ({ value, label: MODE_LABELS[value] ?? value }))}
          onChange={onChange}
        />
      </div>
    </div>
  )
}
