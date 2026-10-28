import { clsx } from 'clsx'
import { AlertCircle, Keyboard, Plus, X } from 'lucide-react'
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
  /** globe shows the Globe key first, read-only: it is on or off for every keyboard at once. */
  globe?: boolean
  /** onGlobe turns the Globe key on, when the Globe key is recorded while it is off. */
  onGlobe?: () => void
}

/** Slot is the key being recorded: an index into triggers, or a new key. */
type Slot = number | 'new'

/**
 * KeyPicker shows a keyboard's trigger keys as keycaps. Clicking a keycap, or
 * the add button, records a key: the daemon reads the next key pressed straight
 * from the keyboard, before any remapping, so a remapped key records as itself.
 * A keycap's icon cancels recording, or removes its key otherwise. The Globe
 * key is not one of a keyboard's own keys: recording it turns it on for every
 * keyboard. Its keys and status are laid out by the grid it sits in: the
 * status spans the grid's full width, below the keys.
 */
export function KeyPicker({
  triggers,
  onChange,
  keyboardId,
  keyboardName,
  globe = false,
  onGlobe,
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
        } else if (recorded.key === GLOBE) {
          if (onGlobe) {
            onGlobe()
          } else {
            setError('The Globe key already switches input sources, on every keyboard at once.')
          }
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

  const remove = (index: number): void => {
    cancel()
    setError(null)
    onChange(triggers.filter((_, other) => other !== index))
  }

  const cancelIcon = <IconButton icon={X} label="Cancel" className={styles.icon} onClick={cancel} />

  const status = error ?? (recording !== null ? `Press a key on ${keyboardName ?? 'any keyboard'}.` : null)

  return (
    <div className={styles.picker}>
      <div className={styles.keys}>
        {globe ? (
          <Tooltip label="On for every keyboard with a Globe key">
            <span className={clsx(styles.keycap, styles.fixed)}>
              <span className={styles.key}>{triggerLabel(GLOBE)}</span>
            </span>
          </Tooltip>
        ) : null}
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
                cancelIcon
              ) : (
                <IconButton icon={X} label={`Remove ${name}`} className={styles.icon} onClick={() => remove(index)} />
              )}
            </div>
          )
        })}
        {recording === 'new' ? (
          <div className={clsx(styles.keycap, styles.recording)}>
            <span className={styles.key} aria-live="polite">
              Press a key…
            </span>
            {cancelIcon}
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
          {error ? (
            <AlertCircle size={14} strokeWidth={2.25} className={styles.statusIcon} aria-hidden />
          ) : (
            <Keyboard size={14} strokeWidth={2.25} className={styles.statusIcon} aria-hidden />
          )}
          <span>{status}</span>
        </p>
      ) : null}
    </div>
  )
}
