import { clsx } from 'clsx'
import { ChevronRight } from 'lucide-react'
import { useId, useState } from 'react'
import type { ReactElement, ReactNode } from 'react'

import styles from './Disclosure.module.css'

interface DisclosureProps {
  title: string
  children: ReactNode
  defaultOpen?: boolean
}

/** Disclosure is a titled group that folds its content away, in place of a native <details>. */
export function Disclosure({ title, children, defaultOpen = false }: DisclosureProps): ReactElement {
  const id = useId()
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div>
      <button
        type="button"
        className={styles.summary}
        aria-expanded={open}
        aria-controls={id}
        onClick={() => setOpen(!open)}
      >
        <ChevronRight size={14} strokeWidth={2.25} className={clsx(styles.chevron, open && styles.open)} />
        {title}
      </button>
      {open ? (
        <div id={id} className={styles.content}>
          {children}
        </div>
      ) : null}
    </div>
  )
}
