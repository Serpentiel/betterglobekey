// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { TriggerSection } from './TriggerSection'

vi.mock('../../lib/api', () => ({ api: {} }))

const config: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
  trigger: { keyboards: {} },
}

const noErrors: ConfigErrors = { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {}, trigger: {} }

const keyboard = (id: string, hasGlobe: boolean): Keyboard => ({
  id,
  name: id,
  transport: 'USB',
  builtIn: false,
  hasGlobe,
  trigger: { key: 'globe', mode: 'dedicated' },
  listed: false,
  triggerMapping: null,
  otherMappings: [],
})

const twoKeyboards: KeyboardsState = {
  report: { keyboards: [keyboard('05ac:0281', true), keyboard('046d:c33f', false)], systemMappings: [] },
  error: null,
}

function renderSection(trigger: Partial<Config['trigger']> = {}, keyboards: KeyboardsState = twoKeyboards) {
  const onChange = vi.fn()
  const merged = { ...config.trigger, ...trigger }

  render(
    <TriggerSection
      config={{ ...config, trigger: merged }}
      errors={noErrors}
      keyboards={keyboards}
      onChange={onChange}
    />,
  )

  return { onChange, trigger: merged }
}

afterEach(cleanup)

describe('TriggerSection', () => {
  it('summarizes the keyboards and opens the sheet', () => {
    renderSection({ keyboards: { '046d:c33f': { key: 'insert', mode: 'pass_through' } } })

    expect(screen.getByText('2 keyboards connected; 1 set their own trigger key.')).toBeTruthy()
    expect(screen.queryByRole('dialog')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Keyboards…' }))
    expect(screen.getByRole('dialog', { name: 'Keyboards' })).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('reports loading and errors', () => {
    renderSection({}, { report: null, error: null })
    expect(screen.getByText('Checking connected keyboards…')).toBeTruthy()

    cleanup()

    renderSection({}, { report: null, error: 'daemon gone' })
    expect(screen.getByText('Unavailable: daemon gone')).toBeTruthy()
  })

  it('counts a single keyboard', () => {
    renderSection({}, { report: { keyboards: [keyboard('046d:c33f', false)], systemMappings: [] }, error: null })

    expect(screen.getByText('1 keyboard connected.')).toBeTruthy()
  })
})
