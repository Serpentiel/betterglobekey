import { clsx } from 'clsx'
import { cloneElement, useEffect, useId, useLayoutEffect, useRef, useState } from 'react'
import type { FocusEvent, ReactElement } from 'react'
import { createPortal } from 'react-dom'

import styles from './Tooltip.module.css'

/** DELAY is how long the pointer rests on an element before its tooltip shows. */
const DELAY = 400

/** WARM is how long after one tooltip hides another shows at once, as moving along a toolbar does on macOS. */
const WARM = 500

/** GAP separates the tooltip from its element, and keeps it off the window's edges. */
const GAP = 6

let warmUntil = 0

interface TooltipProps {
  label: string
  /** className styles the wrapper the tooltip listens on, e.g. to size it in a flex row. */
  className?: string
  /** disabled keeps the tooltip hidden, e.g. while its element is busy. */
  disabled?: boolean
  children: ReactElement<{ 'aria-describedby'?: string }>
}

/**
 * Tooltip labels its child on hover and keyboard focus, in place of the
 * webview's native title tooltip: it shows sooner, and in the app's style.
 */
export function Tooltip({ label, className, disabled = false, children }: TooltipProps): ReactElement {
  const id = useId()
  const anchorRef = useRef<HTMLSpanElement>(null)
  const tipRef = useRef<HTMLDivElement>(null)
  const timer = useRef<number | undefined>(undefined)
  const [open, setOpen] = useState(false)
  const [position, setPosition] = useState<{ top: number; left: number; below: boolean } | null>(null)

  const show = (): void => {
    window.clearTimeout(timer.current)
    if (disabled) {
      return
    }

    if (Date.now() < warmUntil) {
      setOpen(true)
    } else {
      timer.current = window.setTimeout(() => setOpen(true), DELAY)
    }
  }

  const hide = (): void => {
    window.clearTimeout(timer.current)
    if (open) {
      warmUntil = Date.now() + WARM
    }

    setOpen(false)
    setPosition(null)
  }

  // Only keyboard focus shows the tooltip; a click focuses too, and hides it.
  const onFocus = (event: FocusEvent): void => {
    try {
      if ((event.target as Element).matches(':focus-visible')) {
        show()
      }
    } catch {
      // :focus-visible is unsupported: leave the tooltip to the pointer.
    }
  }

  useEffect(() => () => window.clearTimeout(timer.current), [])

  useEffect(() => {
    if (disabled) {
      hide()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [disabled])

  // Centre the tooltip above its element, below it near the window's top, kept inside the window.
  useLayoutEffect(() => {
    if (!open) {
      return undefined
    }

    const anchor = anchorRef.current?.getBoundingClientRect()
    const tip = tipRef.current?.getBoundingClientRect()
    if (!anchor || !tip) {
      return undefined
    }

    const below = anchor.top - tip.height - GAP < GAP
    const centred = anchor.left + anchor.width / 2 - tip.width / 2

    setPosition({
      top: below ? anchor.bottom + GAP : anchor.top - tip.height - GAP,
      left: Math.min(Math.max(centred, GAP), window.innerWidth - tip.width - GAP),
      below,
    })

    const close = (): void => hide()
    window.addEventListener('scroll', close, true)

    return () => window.removeEventListener('scroll', close, true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, label])

  return (
    <>
      <span
        ref={anchorRef}
        className={clsx(styles.anchor, className)}
        onPointerEnter={show}
        onPointerLeave={hide}
        onPointerDown={hide}
        onFocus={onFocus}
        onBlur={hide}
        onKeyDown={(event) => event.key === 'Escape' && hide()}
      >
        {cloneElement(children, { 'aria-describedby': open ? id : undefined })}
      </span>
      {open
        ? createPortal(
            <div
              ref={tipRef}
              id={id}
              role="tooltip"
              className={clsx(styles.tooltip, position && styles.shown, position?.below && styles.below)}
              style={position ? { top: position.top, left: position.left } : undefined}
            >
              {label}
            </div>,
            document.body,
          )
        : null}
    </>
  )
}
