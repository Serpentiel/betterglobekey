import { clsx } from 'clsx'
import type { ButtonHTMLAttributes } from 'react'
import type { LucideIcon } from 'lucide-react'

import styles from './IconButton.module.css'

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: LucideIcon
  label: string
  tone?: 'default' | 'danger'
}

/** IconButton is an accessible, icon-only button (label is required for a11y). */
export function IconButton({ icon: Icon, label, tone = 'default', className, ...rest }: IconButtonProps): JSX.Element {
  return (
    <button
      type="button"
      className={clsx(styles.iconButton, styles[tone], className)}
      aria-label={label}
      title={label}
      {...rest}
    >
      <Icon size={16} strokeWidth={2} />
    </button>
  )
}
