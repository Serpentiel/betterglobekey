import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'

import { Stack } from './Stack'
import styles from './Section.module.css'

interface SectionProps {
  icon: LucideIcon
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}

/** Section is a titled, elevated panel grouping related controls. */
export function Section({ icon: Icon, title, description, actions, children }: SectionProps): JSX.Element {
  return (
    <section className={styles.section}>
      <header className={styles.header}>
        <span className={styles.icon}>
          <Icon size={16} strokeWidth={2} />
        </span>
        <div className={styles.heading}>
          <h2 className={styles.title}>{title}</h2>
          {description ? <p className={styles.description}>{description}</p> : null}
        </div>
        {actions ? <div className={styles.actions}>{actions}</div> : null}
      </header>
      <Stack gap={16}>{children}</Stack>
    </section>
  )
}
