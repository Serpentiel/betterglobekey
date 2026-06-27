import { Loader2 } from 'lucide-react'

import styles from './Spinner.module.css'

interface SpinnerProps {
  size?: number
}

export function Spinner({ size = 18 }: SpinnerProps): JSX.Element {
  return <Loader2 className={styles.spinner} size={size} strokeWidth={2.25} aria-hidden />
}
