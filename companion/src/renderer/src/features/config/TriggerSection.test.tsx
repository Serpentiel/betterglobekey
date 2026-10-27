// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { TriggerSection } from './TriggerSection'

const config: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
  trigger: { key: 'caps_lock', mode: 'fn', keyboards: { applyTo: 'without_globe', overrides: {} } },
}

const noErrors: ConfigErrors = { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {}, trigger: {} }

const keyboard = (id: string, hasGlobe: boolean): Keyboard => ({
  id,
  name: id,
  transport: 'USB',
  builtIn: false,
  hasGlobe,
  inScope: false,
  triggerMapping: null,
  otherMappings: [],
})

const twoKeyboards: KeyboardsState = {
  report: { keyboards: [keyboard('05ac:0281', true), keyboard('046d:c33f', false)], systemMappings: [] },
  error: null,
}

function renderSection(trigger: Partial<Config['trigger']> = {}, keyboards: KeyboardsState = twoKeyboards) {
  render(
    <TriggerSection
      config={{ ...config, trigger: { ...config.trigger, ...trigger } }}
      errors={noErrors}
      keyboards={keyboards}
      onChange={vi.fn()}
    />,
  )
}

afterEach(cleanup)

describe('TriggerSection', () => {
  it('summarizes where the key is remapped and opens the sheet', () => {
    renderSection()

    expect(screen.getByText('Remapped on 1 of 2 keyboards.')).toBeTruthy()
    expect(screen.queryByRole('dialog')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Keyboards…' }))
    expect(screen.getByRole('dialog', { name: 'Keyboards' })).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('warns in the summary when a dedicated trigger reaches no keyboard', () => {
    renderSection({ mode: 'dedicated', keyboards: { applyTo: 'without_globe', overrides: { '046d:c33f': false } } })

    expect(screen.getByText('Not remapped on any keyboard (2 connected), so nothing switches.')).toBeTruthy()
  })

  it('counts keyboards when the key is not remapped', () => {
    renderSection({ key: 'globe' })

    expect(screen.getByText('2 keyboards connected.')).toBeTruthy()
  })

  it('reports loading and errors', () => {
    renderSection({}, { report: null, error: null })
    expect(screen.getByText('Checking connected keyboards…')).toBeTruthy()

    cleanup()

    renderSection({}, { report: null, error: 'daemon gone' })
    expect(screen.getByText('Unavailable: daemon gone')).toBeTruthy()
  })
})
