import type { ReactNode } from 'react'

import styles from './Titlebar.module.css'

interface TitlebarProps {
  title: string
  subtitle?: string
  actions?: ReactNode
}

/**
 * Titlebar is the draggable window chrome. The bar itself is a drag region
 * (so the frameless window can be moved); interactive children opt back out.
 */
export function Titlebar({ title, subtitle, actions }: TitlebarProps): JSX.Element {
  return (
    <header className={styles.titlebar}>
      <div className={styles.heading}>
        <span className={styles.title}>{title}</span>
        {subtitle ? <span className={styles.subtitle}>{subtitle}</span> : null}
      </div>
      {actions ? <div className={styles.actions}>{actions}</div> : null}
    </header>
  )
}
