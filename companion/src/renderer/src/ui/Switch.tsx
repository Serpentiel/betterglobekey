import type { ReactElement } from 'react'
import { clsx } from 'clsx'

import styles from './Switch.module.css'

interface SwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label: string
  id?: string
}

/** Switch is an Apple-style toggle. */
export function Switch({ checked, onChange, label, id }: SwitchProps): ReactElement {
  return (
    <button
      type="button"
      role="switch"
      id={id}
      aria-checked={checked}
      aria-label={label}
      className={clsx(styles.track, checked && styles.on)}
      onClick={() => onChange(!checked)}
    >
      <span className={styles.thumb} />
    </button>
  )
}
