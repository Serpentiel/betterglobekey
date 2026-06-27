import { AlertTriangle, CheckCircle2, Info } from 'lucide-react'
import { clsx } from 'clsx'

import styles from './Banner.module.css'

type Tone = 'success' | 'error' | 'info'

interface BannerProps {
  tone: Tone
  children: string
}

const icons = {
  success: CheckCircle2,
  error: AlertTriangle,
  info: Info,
} as const

export function Banner({ tone, children }: BannerProps): JSX.Element {
  const Icon = icons[tone]

  return (
    <div className={clsx(styles.banner, styles[tone])} role="status">
      <Icon size={16} strokeWidth={2.25} />
      <span>{children}</span>
    </div>
  )
}
