import { clsx } from 'clsx'
import type { InputHTMLAttributes } from 'react'

import styles from './Input.module.css'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean
}

export function Input({ invalid = false, className, ...rest }: InputProps): JSX.Element {
  return <input className={clsx(styles.input, invalid && styles.invalid, className)} aria-invalid={invalid} {...rest} />
}
