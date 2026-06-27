import { clsx } from 'clsx'
import type { ReactNode } from 'react'

import styles from './Card.module.css'

interface CardProps {
  children: ReactNode
  className?: string
}

/** Card is a low-elevation inset container used to group items within a Section. */
export function Card({ children, className }: CardProps): JSX.Element {
  return <div className={clsx(styles.card, className)}>{children}</div>
}
