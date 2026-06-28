import type { ReactElement, ReactNode } from 'react'

import styles from './Setting.module.css'

interface SettingProps {
  title: string
  description?: string
  control: ReactNode
}

/** Setting is a row with a title/description on the left and a control on the right. */
export function Setting({ title, description, control }: SettingProps): ReactElement {
  return (
    <div className={styles.setting}>
      <div className={styles.text}>
        <span className={styles.title}>{title}</span>
        {description ? <span className={styles.description}>{description}</span> : null}
      </div>
      <div className={styles.control}>{control}</div>
    </div>
  )
}
