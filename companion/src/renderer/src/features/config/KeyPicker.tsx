import { clsx } from 'clsx'
import { Globe, Keyboard } from 'lucide-react'
import { type ReactElement, useEffect, useState } from 'react'

import { GLOBE, triggerKeyForCode, triggerLabel, triggerName } from '../../lib/trigger-keys'
import { Button } from '../../ui'
import styles from './KeyPicker.module.css'

interface Props {
  value: string
  /** remapped reports whether value is currently remapped, so pressing it no longer reaches the page. */
  remapped: boolean
  onChange: (key: string) => void
  id?: string
}

/**
 * KeyPicker shows the trigger key and records a new one from the next key press.
 * Globe has its own button: web views never receive the Globe (fn) key.
 */
export function KeyPicker({ value, remapped, onChange, id }: Props): ReactElement {
  const [recording, setRecording] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!recording) {
      return undefined
    }

    const onKey = (event: KeyboardEvent): void => {
      // Caps Lock reports its key-down only when locking, so take either edge.
      if (event.type === 'keyup' && event.code !== 'CapsLock') {
        return
      }

      event.preventDefault()
      event.stopPropagation()

      if (event.code === 'Escape') {
        setRecording(false)

        return
      }

      const key = triggerKeyForCode(event.code)
      if (!key) {
        setError(
          `${event.key === ' ' ? 'Space' : event.key || event.code} can't switch input sources. Try another key.`,
        )

        return
      }

      onChange(key.value)
      setError(null)
      setRecording(false)
    }

    const stop = (): void => setRecording(false)

    window.addEventListener('keydown', onKey, true)
    window.addEventListener('keyup', onKey, true)
    window.addEventListener('blur', stop)

    return () => {
      window.removeEventListener('keydown', onKey, true)
      window.removeEventListener('keyup', onKey, true)
      window.removeEventListener('blur', stop)
    }
  }, [recording, onChange])

  const status = error
    ? error
    : recording
      ? remapped
        ? `Press a key, or Esc to cancel. ${triggerName(value)} is remapped right now, so pressing it won't register.`
        : 'Press a key, or Esc to cancel.'
      : null

  return (
    <div className={styles.picker}>
      <kbd className={clsx(styles.keycap, recording && styles.recording)} aria-live="polite">
        {recording ? 'Press a key…' : triggerLabel(value)}
      </kbd>
      <Button
        id={id}
        icon={Keyboard}
        aria-pressed={recording}
        onClick={() => {
          setError(null)
          setRecording(!recording)
        }}
      >
        {recording ? 'Cancel' : 'Record'}
      </Button>
      <Button
        variant="ghost"
        icon={Globe}
        disabled={value === GLOBE}
        onClick={() => {
          setError(null)
          setRecording(false)
          onChange(GLOBE)
        }}
      >
        Use Globe
      </Button>
      {status ? (
        <p className={clsx(styles.status, error && styles.error)} role="status">
          {status}
        </p>
      ) : null}
    </div>
  )
}
