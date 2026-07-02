// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import type { Api, Config, InputSource, Version } from '../../shared/types'

const validConfig: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
}

const sources: InputSource[] = [{ id: 'us', name: 'U.S.' }]
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

afterEach(cleanup)

describe('App', () => {
  it('shows the editor once the configuration loads', async () => {
    window.api = mockApi()

    render(<App />)
    expect(screen.queryByText('Behavior')).toBeNull() // still loading

    await waitFor(() => expect(screen.getByText('Behavior')).toBeTruthy())
  })

  it('shows a connection error and recovers on retry', async () => {
    const getConfig = vi
      .fn<Api['getConfig']>()
      .mockImplementationOnce(() => Promise.reject(new Error('no socket')))
      .mockImplementation(() => Promise.resolve(validConfig))
    window.api = mockApi({ getConfig })

    render(<App />)

    await waitFor(() => expect(screen.getByText(/no socket/)).toBeTruthy())

    fireEvent.click(screen.getByRole('button', { name: /Retry/ }))

    await waitFor(() => expect(screen.getByText('Behavior')).toBeTruthy())
  })

  it('saves via the Cmd/Ctrl+S shortcut after an edit', async () => {
    const applyConfig = vi.fn(() => Promise.resolve())
    window.api = mockApi({ applyConfig })

    render(<App />)
    await waitFor(() => expect(screen.getByText('Behavior')).toBeTruthy())

    // Edit something so the config becomes dirty and therefore saveable.
    fireEvent.click(screen.getByRole('switch', { name: 'Toggle the HUD' }))

    fireEvent.keyDown(document.body, { key: 's', metaKey: true })

    await waitFor(() => expect(applyConfig).toHaveBeenCalledTimes(1))
  })
})
