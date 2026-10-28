import { clsx } from 'clsx'
import { Globe, X } from 'lucide-react'
import { type ReactElement, useEffect, useRef, useState } from 'react'

import { api } from '../../lib/api'
import { toMessage } from '../../lib/errors'
import { GLOBE, triggerLabel } from '../../lib/trigger-keys'
import { IconButton, Tooltip } from '../../ui'
import styles from './KeyPicker.module.css'

interface Props {
  value: string
  onChange: (key: string) => void
  /** keyboardId records only from this keyboard; absent records from any keyboard. */
  keyboardId?: string
  /** keyboardName names the keyboard to press the key on. */
  keyboardName?: string
  /** disabled turns recording off, e.g. for a keyboard that is not connected. */
  disabled?: boolean
  /** hasGlobe offers Use Globe, for a keyboard with a Globe key; without one, Clear sets no key. */
  hasGlobe?: boolean
  id?: string
}

/**
 * KeyPicker shows a trigger key on a keycap; clicking it records a new one: the
 * daemon reads the next key pressed straight from the keyboard, before any
 * remapping, so a remapped key records as itself. Icons inside the keycap use
 * the Globe key, on a keyboard with one, cancel recording, or clear the key on a
 * keyboard without a Globe key.
 */
export function KeyPicker({
  value,
  onChange,
  keyboardId,
  keyboardName,
  disabled = false,
  hasGlobe = true,
  id,
}: Props): ReactElement {
  const [recording, setRecording] = useState(false)
  const [error, setError] = useState<string | null>(null)
  // attempt identifies the recording in progress, so a cancelled one is ignored.
  const attempt = useRef(0)

  useEffect(() => () => void attempt.current++, [])

  const record = (): void => {
    const current = ++attempt.current
    setError(null)
    setRecording(true)

    api
      .recordKey(keyboardId ?? '')
      .then((recorded) => {
        if (attempt.current !== current) {
          return
        }

        if (recorded.key) {
          onChange(recorded.key)
        } else {
          setError(`${recorded.keyName} can't switch input sources. Try another key.`)
        }
      })
      .catch((cause: unknown) => {
        if (attempt.current === current) {
          setError(toMessage(cause))
        }
      })
      .finally(() => {
        if (attempt.current === current) {
          setRecording(false)
        }
      })
  }

  const cancel = (): void => {
    attempt.current++
    setRecording(false)
  }

  const status = error ?? (recording ? `Press a key on ${keyboardName ?? 'any keyboard'}.` : null)

  const choose = (key: string): void => {
    cancel()
    setError(null)
    onChange(key)
  }

  return (
    <div className={styles.picker}>
      <div className={clsx(styles.keycap, recording && styles.recording, disabled && styles.disabled)}>
        <Tooltip label="Click, then press a key" className={styles.anchor} disabled={recording || disabled}>
          <button
            id={id}
            type="button"
            className={styles.key}
            aria-label="Record a trigger key"
            aria-pressed={recording}
            disabled={disabled}
            onClick={recording ? undefined : record}
          >
            <span aria-live="polite">{recording ? 'Press a key…' : triggerLabel(value)}</span>
          </button>
        </Tooltip>
        {hasGlobe && value !== GLOBE ? (
          <IconButton icon={Globe} label="Use Globe key" className={styles.icon} onClick={() => choose(GLOBE)} />
        ) : null}
        {recording ? <IconButton icon={X} label="Cancel" className={styles.icon} onClick={cancel} /> : null}
        {!recording && !hasGlobe && value !== '' ? (
          <IconButton icon={X} label="Clear" className={styles.icon} onClick={() => choose('')} />
        ) : null}
      </div>
      {status ? (
        <p className={clsx(styles.status, error && styles.error)} role="status">
          {status}
        </p>
      ) : null}
    </div>
  )
}
