import type { ReactNode } from 'react'

import styles from './Footer.module.css'

interface FooterProps {
  children: ReactNode
}

/** Footer is the thin status bar pinned to the bottom of the window. */
export function Footer({ children }: FooterProps): ReactNode {
  return <footer className={styles.footer}>{children}</footer>
}
