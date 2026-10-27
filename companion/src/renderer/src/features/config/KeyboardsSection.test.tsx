// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard, KeyboardReport } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { KeyboardsSection } from './KeyboardsSection'

const config: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
  trigger: { key: 'caps_lock', mode: 'globe', keyboards: { applyTo: 'without_globe', overrides: {} } },
}

const capsToGlobe = { src: 0x700000039, dst: 0xff00000003, srcName: 'Caps Lock', dstName: 'Globe' }
const graveToEscape = { src: 0x700000035, dst: 0x700000029, srcName: 'Grave Accent', dstName: 'Escape' }

const builtIn: Keyboard = {
  id: '05ac:0281',
  name: 'Apple Internal Keyboard',
  transport: 'SPI',
  builtIn: true,
  hasGlobe: true,
  inScope: false,
  triggerMapping: null,
  otherMappings: [],
}

const logitech: Keyboard = {
  id: '046d:c33f',
  name: 'Logitech K120',
  transport: 'USB',
  builtIn: false,
  hasGlobe: false,
  inScope: true,
  triggerMapping: capsToGlobe,
  otherMappings: [graveToEscape],
}

const report: KeyboardReport = { keyboards: [builtIn, logitech], systemMappings: [graveToEscape] }

function renderSection(overrides: Partial<Config['trigger']> = {}, state: KeyboardsState = { report, error: null }) {
  const onChange = vi.fn()
  const trigger = { ...config.trigger, ...overrides }

  render(<KeyboardsSection config={{ ...config, trigger }} keyboards={state} onChange={onChange} />)

  return { onChange, trigger }
}

const toggle = (name: string): HTMLButtonElement =>
  screen.getByRole('switch', { name: `Remap the trigger key on ${name}` }) as HTMLButtonElement

afterEach(cleanup)

describe('KeyboardsSection', () => {
  it('lists keyboards with their mappings', () => {
    renderSection()

    expect(screen.getByText('05ac:0281 · built-in · has Globe')).toBeTruthy()
    expect(screen.getByText('046d:c33f · USB')).toBeTruthy()
    expect(screen.getByText('Caps Lock → Globe')).toBeTruthy()
    expect(screen.getAllByText('Grave Accent → Escape')).toHaveLength(2)
    expect(screen.getByText('System-wide')).toBeTruthy()
    expect(screen.getByText('No key mappings')).toBeTruthy()

    expect(toggle('Apple Internal Keyboard').getAttribute('aria-checked')).toBe('false')
    expect(toggle('Logitech K120').getAttribute('aria-checked')).toBe('true')
  })

  it('includes an Apple keyboard with an override', () => {
    const { onChange, trigger } = renderSection()

    fireEvent.click(toggle('Apple Internal Keyboard'))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: { '05ac:0281': true } } },
    })
  })

  it('drops an override that matches the default', () => {
    const { onChange, trigger } = renderSection({
      keyboards: { applyTo: 'without_globe', overrides: { '046d:c33f': false } },
    })

    expect(screen.getByText('removed when saved')).toBeTruthy()

    fireEvent.click(toggle('Logitech K120'))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: {} } },
    })
  })

  it('shows a mapping still to apply', () => {
    renderSection({ mode: 'dedicated', keyboards: { applyTo: 'all', overrides: {} } })

    expect(screen.getByText('Caps Lock → F20')).toBeTruthy()
    expect(screen.getByText('applied once saved')).toBeTruthy()
  })

  it('warns when a dedicated trigger reaches no keyboard', () => {
    renderSection(
      { mode: 'dedicated', keyboards: { applyTo: 'without_globe', overrides: {} } },
      { report: { keyboards: [builtIn], systemMappings: [] }, error: null },
    )

    expect(screen.getByText(/No connected keyboard remaps the trigger key/)).toBeTruthy()
  })

  it('lists and forgets overrides of disconnected keyboards', () => {
    const { onChange, trigger } = renderSection({
      keyboards: { applyTo: 'without_globe', overrides: { '1234:5678': true } },
    })

    expect(screen.getByText('1234:5678 (not connected)')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Forget' }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: {} } },
    })
  })

  it('disables the toggles when the key is not remapped', () => {
    renderSection({ key: 'right_command', mode: 'pass_through' })

    expect(toggle('Logitech K120').disabled).toBe(true)
    expect(screen.getByText('The connected keyboards and their key mappings.')).toBeTruthy()
  })

  it('shows a report error', () => {
    renderSection({}, { report: null, error: 'unknown method ListKeyboards' })

    expect(screen.getByText('Keyboards unavailable: unknown method ListKeyboards')).toBeTruthy()
  })
})
