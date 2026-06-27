import { clsx } from 'clsx'
import type { ButtonHTMLAttributes, ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'

import { Spinner } from './Spinner'
import styles from './Button.module.css'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost'
  icon?: LucideIcon
  loading?: boolean
  children: ReactNode
}

export function Button({
  variant = 'secondary',
  icon: Icon,
  loading = false,
  className,
  children,
  disabled,
  ...rest
}: ButtonProps): JSX.Element {
  return (
    <button
      type="button"
      className={clsx(styles.button, styles[variant], className)}
      disabled={disabled ?? loading}
      {...rest}
    >
      {loading ? <Spinner size={14} /> : Icon ? <Icon size={15} strokeWidth={2.25} /> : null}
      <span>{children}</span>
    </button>
  )
}
