import { useEffect, useState } from 'react'

import type { KeyboardReport } from '../../../shared/types'
import { api } from '../lib/api'
import { toMessage } from '../lib/errors'

/** KEYBOARDS_POLL_MS matches how often the daemon re-applies its mappings. */
export const KEYBOARDS_POLL_MS = 2000

export interface KeyboardsState {
  report: KeyboardReport | null
  error: string | null
}

/** useKeyboards polls the daemon's keyboard report, so keyboards plugged in or
 * out and mappings applied after a save show up on their own. */
export function useKeyboards(): KeyboardsState {
  const [state, setState] = useState<KeyboardsState>({ report: null, error: null })

  useEffect(() => {
    let active = true

    const poll = async (): Promise<void> => {
      try {
        const report = await api.listKeyboards()
        if (active) {
          setState({ report, error: null })
        }
      } catch (error) {
        if (active) {
          setState((previous) => ({ report: previous.report, error: toMessage(error) }))
        }
      }
    }

    void poll()
    const timer = window.setInterval(() => void poll(), KEYBOARDS_POLL_MS)

    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [])

  return state
}
