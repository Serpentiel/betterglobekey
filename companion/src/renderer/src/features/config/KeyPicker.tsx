import { clsx } from 'clsx'
import { Globe, Plus, X } from 'lucide-react'
import { type ReactElement, useEffect, useRef, useState } from 'react'

import type { KeyboardTrigger } from '../../../../shared/types'
import { api } from '../../lib/api'
import { toMessage } from '../../lib/errors'
import { GLOBE, MODE_DEDICATED, triggerLabel, triggerName, withKey } from '../../lib/trigger-keys'
import { IconButton, Tooltip } from '../../ui'
import styles from './KeyPicker.module.css'

interface Props {
  triggers: KeyboardTrigger[]
  onChange: (triggers: KeyboardTrigger[]) => void
  /** keyboardId records only from this keyboard; absent records from any keyboard. */
  keyboardId?: string
  /** keyboardName names the keyboard to press the key on. */
  keyboardName?: string
  /** hasGlobe offers the Globe key while recording, for a keyboard with one. */
  hasGlobe?: boolean
  /** globeShared labels the Globe key's controls as acting on every keyboard with one. */
  globeShared?: boolean
}

/** Slot is the key being recorded: an index into triggers, or a new key. */
type Slot = number | 'new'

/**
 * KeyPicker shows a keyboard's trigger keys as keycaps. Clicking a keycap, or
 * the add button, records a key: the daemon reads the next key pressed straight
 * from the keyboard, before any remapping, so a remapped key records as itself.
 * While recording, icons inside the keycap use the Globe key, on a keyboard with
 * one, or cancel; otherwise a keycap's icon removes its key.
 */
export function KeyPicker({
  triggers,
  onChange,
  keyboardId,
  keyboardName,
  hasGlobe = true,
  globeShared = false,
}: Props): ReactElement {
  const [recording, setRecording] = useState<Slot | null>(null)
  const [error, setError] = useState<string | null>(null)
  // attempt identifies the recording in progress, so a cancelled one is ignored.
  const attempt = useRef(0)

  useEffect(() => () => void attempt.current++, [])

  const takenElsewhere = (slot: Slot, key: string): boolean =>
    triggers.some((trigger, index) => trigger.key === key && index !== slot)

  const put = (slot: Slot, key: string): void => {
    if (slot === 'new') {
      onChange([...triggers, withKey({ key, mode: MODE_DEDICATED }, key)])
    } else {
      onChange(triggers.map((trigger, index) => (index === slot ? withKey(trigger, key) : trigger)))
    }
  }

  const record = (slot: Slot): void => {
    const current = ++attempt.current
    setError(null)
    setRecording(slot)

    api
      .recordKey(keyboardId ?? '')
      .then((recorded) => {
        if (attempt.current !== current) {
          return
        }

        if (!recorded.key) {
          setError(`${recorded.keyName} can't switch input sources. Try another key.`)
        } else if (takenElsewhere(slot, recorded.key)) {
          setError(`${triggerName(recorded.key)} already switches input sources here.`)
        } else {
          put(slot, recorded.key)
        }
      })
      .catch((cause: unknown) => {
        if (attempt.current === current) {
          setError(toMessage(cause))
        }
      })
      .finally(() => {
        if (attempt.current === current) {
          setRecording(null)
        }
      })
  }

  const cancel = (): void => {
    attempt.current++
    setRecording(null)
  }

  const choose = (slot: Slot, key: string): void => {
    cancel()
    put(slot, key)
  }

  const remove = (index: number): void => {
    cancel()
    setError(null)
    onChange(triggers.filter((_, other) => other !== index))
  }

  // While recording, the keycap offers the Globe key, unless another keycap has it, and cancelling.
  const recordingIcons = (slot: Slot): ReactElement => (
    <>
      {hasGlobe && !takenElsewhere(slot, GLOBE) && (slot === 'new' || triggers[slot]?.key !== GLOBE) ? (
        <IconButton
          icon={Globe}
          label={globeShared ? 'Use Globe key on every keyboard' : 'Use Globe key'}
          className={styles.icon}
          onClick={() => choose(slot, GLOBE)}
        />
      ) : null}
      <IconButton icon={X} label="Cancel" className={styles.icon} onClick={cancel} />
    </>
  )

  const status = error ?? (recording !== null ? `Press a key on ${keyboardName ?? 'any keyboard'}.` : null)

  return (
    <div className={styles.picker}>
      <div className={styles.keys}>
        {triggers.map((trigger, index) => {
          const name = triggerName(trigger.key)
          const active = recording === index

          return (
            <div key={trigger.key} className={clsx(styles.keycap, active && styles.recording)}>
              <Tooltip label="Click, then press a key" className={styles.anchor} disabled={recording !== null}>
                <button
                  type="button"
                  className={styles.key}
                  aria-label={`Change ${name}`}
                  aria-pressed={active}
                  onClick={active ? undefined : () => record(index)}
                >
                  <span aria-live="polite">{active ? 'Press a key…' : triggerLabel(trigger.key)}</span>
                </button>
              </Tooltip>
              {active ? (
                recordingIcons(index)
              ) : (
                <IconButton
                  icon={X}
                  label={trigger.key === GLOBE && globeShared ? 'Remove Globe from every keyboard' : `Remove ${name}`}
                  className={styles.icon}
                  onClick={() => remove(index)}
                />
              )}
            </div>
          )
        })}
        {recording === 'new' ? (
          <div className={clsx(styles.keycap, styles.recording)}>
            <span className={styles.key} aria-live="polite">
              Press a key…
            </span>
            {recordingIcons('new')}
          </div>
        ) : (
          <Tooltip label="Add a trigger key">
            <button type="button" className={styles.add} aria-label="Add a trigger key" onClick={() => record('new')}>
              <Plus size={14} strokeWidth={2.25} />
              {triggers.length === 0 ? <span>Add key</span> : null}
            </button>
          </Tooltip>
        )}
      </div>
      {status ? (
        <p className={clsx(styles.status, error && styles.error)} role="status">
          {status}
        </p>
      ) : null}
    </div>
  )
}
