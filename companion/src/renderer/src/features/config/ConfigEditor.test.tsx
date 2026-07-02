// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, InputSource } from '../../../../shared/types'
import type { ConfigController } from '../../hooks/useConfig'
import type { ConfigErrors } from '../../lib/config-schema'
import { ConfigEditor } from './ConfigEditor'

const validConfig: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
}

const sources: InputSource[] = [
  { id: 'us', name: 'U.S.' },
  { id: 'de', name: 'German' },
]

const noErrors: ConfigErrors = { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {} }

function makeController(overrides: Partial<ConfigController> = {}): ConfigController {
  return {
    phase: { status: 'ready' },
    config: validConfig,
    sources,
    version: { version: '4.0.0', commit: 'abc1234' },
    validation: { valid: true, errors: noErrors },
    dirty: false,
    saving: false,
    saved: false,
    saveError: null,
    update: vi.fn(),
    save: vi.fn(() => Promise.resolve()),
    revert: vi.fn(),
    reload: vi.fn(() => Promise.resolve()),
    ...overrides,
  }
}

afterEach(cleanup)

describe('ConfigEditor', () => {
  it('renders every section and the version footer', () => {
    render(<ConfigEditor controller={makeController()} />)

    for (const heading of ['Behavior', 'On-screen HUD', 'Collections', 'Logging']) {
      expect(screen.getByText(heading)).toBeTruthy()
    }

    expect(screen.getByText('v4.0.0')).toBeTruthy()
  })

  it('renders nothing without a config', () => {
    const { container } = render(<ConfigEditor controller={makeController({ config: null })} />)

    expect(container.firstChild).toBeNull()
  })

  it('disables Save and hides Revert when the config is clean', () => {
    render(<ConfigEditor controller={makeController({ dirty: false })} />)

    expect((screen.getByRole('button', { name: 'Save' }) as HTMLButtonElement).disabled).toBe(true)
    expect(screen.queryByRole('button', { name: 'Revert' })).toBeNull()
  })

  it('saves and reverts when dirty and valid', () => {
    const save = vi.fn(() => Promise.resolve())
    const revert = vi.fn()
    render(<ConfigEditor controller={makeController({ dirty: true, save, revert })} />)

    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(save).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'Revert' }))
    expect(revert).toHaveBeenCalledTimes(1)
  })

  it('toggling a behavior switch reports the patched config', () => {
    const update = vi.fn()
    render(<ConfigEditor controller={makeController({ update })} />)

    fireEvent.click(screen.getByRole('switch', { name: 'Toggle double press' }))

    expect(update).toHaveBeenCalledWith({
      ...validConfig,
      doublePress: { ...validConfig.doublePress, enabled: false },
    })
  })

  it('editing a logging number field reports the patched config', () => {
    const update = vi.fn()
    render(<ConfigEditor controller={makeController({ update })} />)

    fireEvent.change(screen.getByLabelText('Retention (days)'), { target: { value: '14' } })

    expect(update).toHaveBeenCalledWith({
      ...validConfig,
      logger: { ...validConfig.logger, retentionDays: 14 },
    })
  })

  it('drives the HUD and logging section controls', () => {
    const update = vi.fn()
    render(<ConfigEditor controller={makeController({ update })} />)

    fireEvent.change(screen.getByLabelText('Visible duration'), { target: { value: '1s' } })
    expect(update).toHaveBeenCalledWith({ ...validConfig, hud: { ...validConfig.hud, duration: '1s' } })

    fireEvent.click(screen.getByRole('switch', { name: 'Toggle collection subtitle' }))
    expect(update).toHaveBeenCalledWith({ ...validConfig, hud: { ...validConfig.hud, showCollection: false } })

    fireEvent.change(screen.getByLabelText('Log file'), { target: { value: '/var/log/bgk.log' } })
    expect(update).toHaveBeenCalledWith({ ...validConfig, logger: { ...validConfig.logger, path: '/var/log/bgk.log' } })

    fireEvent.change(screen.getByLabelText('Retention (files)'), { target: { value: '5' } })
    expect(update).toHaveBeenCalledWith({ ...validConfig, logger: { ...validConfig.logger, retentionFiles: 5 } })
  })

  it('toggles the reverse and HUD switches', () => {
    const update = vi.fn()
    render(<ConfigEditor controller={makeController({ update })} />)

    fireEvent.change(screen.getByLabelText('Double-press maximum delay'), { target: { value: '300ms' } })
    expect(update).toHaveBeenCalledWith({
      ...validConfig,
      doublePress: { ...validConfig.doublePress, maximumDelay: '300ms' },
    })

    fireEvent.click(screen.getByRole('switch', { name: 'Toggle reverse modifier' }))
    expect(update).toHaveBeenCalledWith({ ...validConfig, reverse: { ...validConfig.reverse, enabled: false } })

    fireEvent.click(screen.getByRole('switch', { name: 'Toggle the HUD' }))
    expect(update).toHaveBeenCalledWith({ ...validConfig, hud: { ...validConfig.hud, enabled: false } })
  })

  it('shows the saved confirmation banner', () => {
    render(<ConfigEditor controller={makeController({ saved: true })} />)

    expect(screen.getByText(/Saved\. The daemon reloaded the change\./)).toBeTruthy()
  })

  it('shows the save-error banner', () => {
    render(<ConfigEditor controller={makeController({ saveError: 'permission denied' })} />)

    expect(screen.getByText('permission denied')).toBeTruthy()
  })

  it('prompts to fix fields when dirty but invalid', () => {
    render(
      <ConfigEditor controller={makeController({ dirty: true, validation: { valid: false, errors: noErrors } })} />,
    )

    expect(screen.getByText(/Fix the highlighted fields to save\./)).toBeTruthy()
  })
})
