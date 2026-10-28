import type { ReactElement, ReactNode } from 'react'
import { AlertTriangle, CheckCircle2, Info } from 'lucide-react'
import { clsx } from 'clsx'

import styles from './Banner.module.css'

type Tone = 'success' | 'error' | 'warning' | 'info'

interface BannerProps {
  tone: Tone
  children: ReactNode
  /** action is a control at the banner's end, such as a button. */
  action?: ReactNode
}

const icons = {
  success: CheckCircle2,
  error: AlertTriangle,
  warning: AlertTriangle,
  info: Info,
} as const

export function Banner({ tone, children, action }: BannerProps): ReactElement {
  const Icon = icons[tone]

  return (
    <div className={clsx(styles.banner, styles[tone])} role="status">
      <Icon size={16} strokeWidth={2.25} className={styles.icon} />
      <div className={styles.body}>{children}</div>
      {action ? <div className={styles.action}>{action}</div> : null}
    </div>
  )
}
