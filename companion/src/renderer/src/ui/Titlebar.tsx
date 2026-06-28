import type { ReactElement, ReactNode } from 'react'

import styles from './Titlebar.module.css'

interface TitlebarProps {
  title: string
  actions?: ReactNode
}

/**
 * Titlebar is the draggable window chrome. The bar itself is a drag region
 * (so the frameless window can be moved); interactive children opt back out.
 */
export function Titlebar({ title, actions }: TitlebarProps): ReactElement {
  return (
    <header className={styles.titlebar}>
      <span className={styles.title}>{title}</span>
      {actions ? <div className={styles.actions}>{actions}</div> : null}
    </header>
  )
}
