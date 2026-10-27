// @vitest-environment jsdom
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KeyboardReport } from '../../../shared/types'
import { api } from '../lib/api'
import { KEYBOARDS_POLL_MS, useKeyboards } from './useKeyboards'

vi.mock('../lib/api', () => ({ api: {} }))

const report: KeyboardReport = { keyboards: [], systemMappings: [] }

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

describe('useKeyboards', () => {
  it('polls the daemon and keeps the last report on an error', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })

    const listKeyboards = vi
      .fn<Api['listKeyboards']>()
      .mockResolvedValueOnce(report)
      .mockRejectedValueOnce(new Error('daemon gone'))
      .mockResolvedValue(report)
    Object.assign(api, { listKeyboards })

    const { result } = renderHook(() => useKeyboards())
    await waitFor(() => expect(result.current.report).toEqual(report))

    await act(() => vi.advanceTimersByTimeAsync(KEYBOARDS_POLL_MS))
    expect(result.current).toEqual({ report, error: 'daemon gone' })

    await act(() => vi.advanceTimersByTimeAsync(KEYBOARDS_POLL_MS))
    expect(result.current).toEqual({ report, error: null })
  })
})
