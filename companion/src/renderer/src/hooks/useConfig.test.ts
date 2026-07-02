// @vitest-environment jsdom
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, Config, InputSource, Version } from '../../../shared/types'
import { useConfig } from './useConfig'

const validConfig: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['com.apple.keylayout.US'] }],
}

const sources: InputSource[] = [
  { id: 'com.apple.keylayout.US', name: 'U.S.' },
  { id: 'com.apple.keylayout.German', name: 'German' },
]

const version: Version = { version: '4.0.0', commit: 'abc1234' }

function mockApi(overrides: Partial<Api> = {}): Api {
  return {
    getConfig: vi.fn(() => Promise.resolve(validConfig)),
    listInputSources: vi.fn(() => Promise.resolve(sources)),
    getVersion: vi.fn(() => Promise.resolve(version)),
    applyConfig: vi.fn(() => Promise.resolve()),
    ...overrides,
  }
}

async function renderReady(api: Api) {
  window.api = api

  const view = renderHook(() => useConfig())
  await waitFor(() => expect(view.result.current.phase.status).toBe('ready'))

  return view
}

afterEach(cleanup)

describe('useConfig', () => {
  it('loads config, sources, and version on mount', async () => {
    window.api = mockApi()

    const { result } = renderHook(() => useConfig())
    expect(result.current.phase.status).toBe('loading')

    await waitFor(() => expect(result.current.phase.status).toBe('ready'))

    expect(result.current.config).toEqual(validConfig)
    expect(result.current.sources).toEqual(sources)
    expect(result.current.version).toEqual(version)
    expect(result.current.dirty).toBe(false)
    expect(result.current.validation.valid).toBe(true)
  })

  it('enters the error phase when loading fails', async () => {
    window.api = mockApi({ getConfig: vi.fn(() => Promise.reject(new Error('boom'))) })

    const { result } = renderHook(() => useConfig())

    await waitFor(() => expect(result.current.phase.status).toBe('error'))
    expect(result.current.phase).toMatchObject({ status: 'error', message: 'boom' })
  })

  it('marks the config dirty after an update and clears the saved flag', async () => {
    const { result } = await renderReady(mockApi())

    act(() => {
      result.current.update({ ...validConfig, hud: { ...validConfig.hud, enabled: false } })
    })

    expect(result.current.dirty).toBe(true)
    expect(result.current.saved).toBe(false)
  })

  it('reverts to the originally loaded config', async () => {
    const { result } = await renderReady(mockApi())

    act(() => {
      result.current.update({ ...validConfig, logger: { ...validConfig.logger, level: 'debug' } })
    })
    expect(result.current.dirty).toBe(true)

    act(() => {
      result.current.revert()
    })

    expect(result.current.dirty).toBe(false)
    expect(result.current.config).toEqual(validConfig)
  })

  it('saves a valid, edited config and clears dirty', async () => {
    const applyConfig = vi.fn(() => Promise.resolve())
    const { result } = await renderReady(mockApi({ applyConfig }))

    const edited: Config = { ...validConfig, hud: { ...validConfig.hud, showCollection: false } }

    act(() => {
      result.current.update(edited)
    })

    await act(async () => {
      await result.current.save()
    })

    expect(applyConfig).toHaveBeenCalledWith(edited)
    expect(result.current.saved).toBe(true)
    expect(result.current.dirty).toBe(false)
  })

  it('refuses to save an invalid config', async () => {
    const applyConfig = vi.fn(() => Promise.resolve())
    const { result } = await renderReady(mockApi({ applyConfig }))

    act(() => {
      result.current.update({ ...validConfig, logger: { ...validConfig.logger, path: '' } })
    })

    expect(result.current.validation.valid).toBe(false)

    await act(async () => {
      await result.current.save()
    })

    expect(applyConfig).not.toHaveBeenCalled()
  })

  it('surfaces a save failure without marking saved', async () => {
    const applyConfig = vi.fn(() => Promise.reject(new Error('write failed')))
    const { result } = await renderReady(mockApi({ applyConfig }))

    act(() => {
      result.current.update({ ...validConfig, reverse: { ...validConfig.reverse, modifier: 'option' } })
    })

    await act(async () => {
      await result.current.save()
    })

    expect(result.current.saveError).toBe('write failed')
    expect(result.current.saved).toBe(false)
    expect(result.current.dirty).toBe(true)
  })

  it('reloads configuration on demand', async () => {
    const getConfig = vi.fn(() => Promise.resolve(validConfig))
    const { result } = await renderReady(mockApi({ getConfig }))

    expect(getConfig).toHaveBeenCalledTimes(1)

    await act(async () => {
      await result.current.reload()
    })

    expect(getConfig).toHaveBeenCalledTimes(2)
  })
})
