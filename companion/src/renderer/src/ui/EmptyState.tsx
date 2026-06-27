import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'

import { Stack } from './Stack'
import styles from './EmptyState.module.css'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description?: string
  action?: ReactNode
}

/** EmptyState is the centered placeholder for empty, loading, or error screens. */
export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps): JSX.Element {
  return (
    <Stack className={styles.empty} align="center" justify="center" gap={12}>
      <span className={styles.icon}>
        <Icon size={26} strokeWidth={1.75} />
      </span>
      <h2 className={styles.title}>{title}</h2>
      {description ? <p className={styles.description}>{description}</p> : null}
      {action ? <div className={styles.action}>{action}</div> : null}
    </Stack>
  )
}
