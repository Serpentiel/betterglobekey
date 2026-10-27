// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard, KeyboardReport } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { KeyboardsSheet } from './KeyboardsSheet'

const config: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
  trigger: { key: 'caps_lock', mode: 'fn', keyboards: { applyTo: 'without_globe', overrides: {} } },
}

const capsToFn = { src: 0x700000039, dst: 0xff00000003, srcName: 'Caps Lock', dstName: 'fn' }
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
  triggerMapping: capsToFn,
  otherMappings: [graveToEscape],
}

const report: KeyboardReport = { keyboards: [builtIn, logitech], systemMappings: [graveToEscape] }

function renderSheet(overrides: Partial<Config['trigger']> = {}, state: KeyboardsState = { report, error: null }) {
  const onChange = vi.fn()
  const onClose = vi.fn()
  const trigger = { ...config.trigger, ...overrides }

  render(
    <KeyboardsSheet open config={{ ...config, trigger }} keyboards={state} onChange={onChange} onClose={onClose} />,
  )

  return { onChange, onClose, trigger }
}

const toggle = (name: string): HTMLButtonElement =>
  screen.getByRole('switch', { name: `Remap Caps Lock on ${name}` }) as HTMLButtonElement

afterEach(cleanup)

describe('KeyboardsSheet', () => {
  it('lists keyboards with where the key is remapped', () => {
    renderSheet()

    expect(screen.getByRole('dialog', { name: 'Keyboards' })).toBeTruthy()
    expect(screen.getByText('Built-in · 05ac:0281 · has Globe')).toBeTruthy()
    expect(screen.getByText('USB · 046d:c33f')).toBeTruthy()
    expect(screen.getByText('Caps Lock → fn')).toBeTruthy()
    expect(screen.getByText('Caps Lock unchanged')).toBeTruthy()

    expect(toggle('Apple Internal Keyboard').getAttribute('aria-checked')).toBe('false')
    expect(toggle('Logitech K120').getAttribute('aria-checked')).toBe('true')
  })

  it('shows other tools’ mappings behind a disclosure', () => {
    renderSheet()

    expect(screen.getByText('Mappings from other tools')).toBeTruthy()
    expect(screen.getAllByText('Grave Accent → Escape')).toHaveLength(2)
    expect(screen.getByText('Every keyboard')).toBeTruthy()
  })

  it('includes an Apple keyboard with an override', () => {
    const { onChange, trigger } = renderSheet()

    fireEvent.click(toggle('Apple Internal Keyboard'))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: { '05ac:0281': true } } },
    })
  })

  it('drops an override that matches the default, and resets one', () => {
    const { onChange, trigger } = renderSheet({
      keyboards: { applyTo: 'without_globe', overrides: { '046d:c33f': false } },
    })

    expect(screen.getByText('Caps Lock → fn, removed once saved')).toBeTruthy()

    fireEvent.click(toggle('Logitech K120'))
    fireEvent.click(screen.getByRole('button', { name: 'Reset' }))

    const reset = { ...config, trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: {} } } }
    expect(onChange).toHaveBeenNthCalledWith(1, reset)
    expect(onChange).toHaveBeenNthCalledWith(2, reset)
  })

  it('sets the default for other keyboards', () => {
    const { onChange, trigger } = renderSheet()

    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.click(screen.getByRole('option', { name: 'Always remap' }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'all', overrides: {} } },
    })
  })

  it('shows a mapping still to apply', () => {
    renderSheet({ mode: 'dedicated', keyboards: { applyTo: 'all', overrides: {} } })

    expect(screen.getByText('Caps Lock → F20 once saved')).toBeTruthy()
  })

  it('warns when a dedicated trigger reaches no keyboard', () => {
    renderSheet(
      { mode: 'dedicated', keyboards: { applyTo: 'without_globe', overrides: {} } },
      { report: { keyboards: [builtIn], systemMappings: [] }, error: null },
    )

    expect(screen.getByText(/No connected keyboard remaps the trigger key/)).toBeTruthy()
  })

  it('lists and forgets overrides of disconnected keyboards', () => {
    const { onChange, trigger } = renderSheet({
      keyboards: { applyTo: 'without_globe', overrides: { '1234:5678': true } },
    })

    expect(screen.getByText('Not connected')).toBeTruthy()
    expect(screen.getByText('Always remapped')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Forget' }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { applyTo: 'without_globe', overrides: {} } },
    })
  })

  it('disables the controls when the key is not remapped', () => {
    renderSheet({ key: 'right_command', mode: 'pass_through' })

    expect(
      (screen.getByRole('switch', { name: 'Remap Right Command on Logitech K120' }) as HTMLButtonElement).disabled,
    ).toBe(true)
    expect((screen.getByRole('combobox') as HTMLButtonElement).disabled).toBe(true)
  })

  it('shows a report error and closes with Done', () => {
    const { onClose } = renderSheet({}, { report: null, error: 'unknown method ListKeyboards' })

    expect(screen.getByText('Keyboards unavailable: unknown method ListKeyboards')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(onClose).toHaveBeenCalled()
  })
})
