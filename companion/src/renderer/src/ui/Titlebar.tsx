import type { ReactElement, ReactNode } from 'react'

import styles from './Titlebar.module.css'

interface TitlebarProps {
  title: string
  actions?: ReactNode
}

/**
 * Titlebar is the draggable window chrome. Only elements marked as a drag
 * region move the window, so the actions stay clickable.
 */
export function Titlebar({ title, actions }: TitlebarProps): ReactElement {
  return (
    <header className={styles.titlebar} data-tauri-drag-region>
      <span className={styles.title} data-tauri-drag-region>
        {title}
      </span>
      {actions ? <div className={styles.actions}>{actions}</div> : null}
    </header>
  )
}
