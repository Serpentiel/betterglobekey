import { ChevronDown, ChevronUp } from 'lucide-react'
import { clsx } from 'clsx'
import type { ReactElement } from 'react'

import styles from './NumberField.module.css'

interface NumberFieldProps {
  value: number
  onChange: (value: number) => void
  min?: number
  max?: number
  step?: number
  id?: string
  invalid?: boolean
}

/** NumberField is a numeric input with a custom up/down stepper (no native spinner). */
export function NumberField({
  value,
  onChange,
  min,
  max,
  step = 1,
  id,
  invalid = false,
}: NumberFieldProps): ReactElement {
  const clamp = (next: number): number => {
    if (Number.isNaN(next)) {
      return min ?? 0
    }

    if (min !== undefined && next < min) {
      return min
    }

    if (max !== undefined && next > max) {
      return max
    }

    return next
  }

  const set = (next: number): void => onChange(clamp(next))

  const canDecrease = min === undefined || value > min
  const canIncrease = max === undefined || value < max

  return (
    <div className={clsx(styles.wrapper, invalid && styles.invalid)}>
      <input
        id={id}
        type="number"
        inputMode="numeric"
        className={styles.input}
        value={Number.isNaN(value) ? '' : value}
        min={min}
        max={max}
        step={step}
        aria-invalid={invalid}
        onChange={(event) => set(Number(event.target.value))}
      />
      <div className={styles.stepper}>
        <button
          type="button"
          className={styles.step}
          aria-label="Increase"
          disabled={!canIncrease}
          onClick={() => set(value + step)}
        >
          <ChevronUp size={12} strokeWidth={2.75} />
        </button>
        <button
          type="button"
          className={styles.step}
          aria-label="Decrease"
          disabled={!canDecrease}
          onClick={() => set(value - step)}
        >
          <ChevronDown size={12} strokeWidth={2.75} />
        </button>
      </div>
    </div>
  )
}
