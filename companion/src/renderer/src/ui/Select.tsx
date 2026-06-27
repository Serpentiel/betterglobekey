import { ChevronDown } from 'lucide-react'
import { clsx } from 'clsx'
import type { SelectHTMLAttributes } from 'react'

import styles from './Select.module.css'

interface SelectOption {
  value: string
  label: string
}

interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'children'> {
  options: SelectOption[]
  placeholder?: string
}

export function Select({ options, placeholder, className, ...rest }: SelectProps): JSX.Element {
  return (
    <div className={clsx(styles.wrapper, className)}>
      <select className={styles.select} {...rest}>
        {placeholder ? <option value="">{placeholder}</option> : null}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <ChevronDown className={styles.chevron} size={15} strokeWidth={2} aria-hidden />
    </div>
  )
}
