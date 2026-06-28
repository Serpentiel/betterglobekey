import { AlertCircle } from 'lucide-react'
import { useId } from 'react'
import type { ReactElement } from 'react'

import styles from './Field.module.css'

interface FieldProps {
  label: string
  hint?: string
  error?: string
  /** A render prop receiving the id to wire label/control accessibility. */
  children: (props: { id: string; invalid: boolean }) => ReactElement
}

/** Field is the canonical label + control + hint/error wrapper for all inputs. */
export function Field({ label, hint, error, children }: FieldProps): ReactElement {
  const id = useId()
  const invalid = Boolean(error)

  return (
    <div className={styles.field}>
      <label className={styles.label} htmlFor={id}>
        {label}
      </label>
      {children({ id, invalid })}
      {error ? (
        <p className={styles.error}>
          <AlertCircle size={13} strokeWidth={2.25} />
          {error}
        </p>
      ) : hint ? (
        <p className={styles.hint}>{hint}</p>
      ) : null}
    </div>
  )
}
