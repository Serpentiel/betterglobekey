import { Check, ChevronsUpDown } from 'lucide-react'
import { clsx } from 'clsx'
import { useEffect, useId, useLayoutEffect, useRef, useState } from 'react'
import type { KeyboardEvent, ReactElement } from 'react'
import { createPortal } from 'react-dom'

import styles from './Select.module.css'

export interface SelectOption {
  value: string
  label: string
}

interface SelectProps {
  value: string
  options: SelectOption[]
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
  id?: string
}

/**
 * Select is a custom popup button (not a native <select>): a styled trigger that
 * opens a portalled, keyboard-navigable menu with a checkmark on the active item.
 */
export function Select({ value, options, onChange, placeholder, disabled = false, id }: SelectProps): ReactElement {
  const listId = `${useId()}-list`
  const triggerRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(0)
  const [position, setPosition] = useState<{ top: number; left: number; width: number } | null>(null)

  const selected = options.find((option) => option.value === value)

  useLayoutEffect(() => {
    if (!open) {
      return undefined
    }

    const place = (): void => {
      const rect = triggerRef.current?.getBoundingClientRect()
      if (rect) {
        setPosition({ top: rect.bottom + 4, left: rect.left, width: rect.width })
      }
    }

    place()
    window.addEventListener('scroll', place, true)
    window.addEventListener('resize', place)

    return () => {
      window.removeEventListener('scroll', place, true)
      window.removeEventListener('resize', place)
    }
  }, [open])

  useEffect(() => {
    if (!open) {
      return undefined
    }

    const onPointerDown = (event: PointerEvent): void => {
      const target = event.target as Node
      if (!triggerRef.current?.contains(target) && !menuRef.current?.contains(target)) {
        setOpen(false)
      }
    }

    document.addEventListener('pointerdown', onPointerDown)

    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  const openMenu = (): void => {
    if (disabled) {
      return
    }

    const index = options.findIndex((option) => option.value === value)
    setActiveIndex(index >= 0 ? index : 0)
    setOpen(true)
  }

  const choose = (next: string): void => {
    onChange(next)
    setOpen(false)
    triggerRef.current?.focus()
  }

  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>): void => {
    if (disabled) {
      return
    }

    if (!open) {
      if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
        event.preventDefault()
        openMenu()
      }

      return
    }

    switch (event.key) {
      case 'Escape':
      case 'Tab':
        setOpen(false)
        break
      case 'ArrowDown':
        event.preventDefault()
        setActiveIndex((index) => Math.min(index + 1, options.length - 1))
        break
      case 'ArrowUp':
        event.preventDefault()
        setActiveIndex((index) => Math.max(index - 1, 0))
        break
      case 'Enter':
      case ' ': {
        event.preventDefault()
        const option = options[activeIndex]
        if (option) {
          choose(option.value)
        }

        break
      }
      default:
        break
    }
  }

  return (
    <div className={styles.wrapper}>
      <button
        ref={triggerRef}
        type="button"
        id={id}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        aria-activedescendant={open ? `${listId}-${activeIndex}` : undefined}
        disabled={disabled}
        className={clsx(styles.trigger, !selected && styles.placeholder)}
        onClick={() => (open ? setOpen(false) : openMenu())}
        onKeyDown={onKeyDown}
      >
        <span className={styles.value}>{selected ? selected.label : placeholder}</span>
        <ChevronsUpDown className={styles.chevron} size={14} strokeWidth={2} aria-hidden />
      </button>

      {open && position
        ? createPortal(
            <div
              ref={menuRef}
              id={listId}
              role="listbox"
              className={styles.menu}
              style={{ top: position.top, left: position.left, minWidth: position.width }}
            >
              {options.map((option, index) => (
                <div
                  key={option.value}
                  id={`${listId}-${index}`}
                  role="option"
                  aria-selected={option.value === value}
                  className={clsx(styles.option, index === activeIndex && styles.active)}
                  onPointerEnter={() => setActiveIndex(index)}
                  onClick={() => choose(option.value)}
                >
                  <span className={styles.check}>
                    {option.value === value ? <Check size={14} strokeWidth={2.75} /> : null}
                  </span>
                  <span className={styles.optionLabel}>{option.label}</span>
                </div>
              ))}
            </div>,
            document.body,
          )
        : null}
    </div>
  )
}
