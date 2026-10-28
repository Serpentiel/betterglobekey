// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard, KeyboardReport, KeyboardTrigger } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { KeyboardsSheet } from './KeyboardsSheet'

vi.mock('../../lib/api', () => ({ api: {} }))

const config: Config = {
  logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
  trigger: { keyboards: {} },
}

const globe: KeyboardTrigger = { key: 'globe', mode: 'fn' }
const capsFn: KeyboardTrigger = { key: 'caps_lock', mode: 'fn' }
const capsToFn = { src: 0x700000039, dst: 0xff00000003, srcName: 'Caps Lock', dstName: 'fn' }
const graveToEscape = { src: 0x700000035, dst: 0x700000029, srcName: 'Grave Accent', dstName: 'Escape' }

const builtIn: Keyboard = {
  id: '05ac:0281',
  name: 'Apple Internal Keyboard',
  transport: 'SPI',
  builtIn: true,
  hasGlobe: true,
  trigger: globe,
  listed: false,
  triggerMapping: null,
  otherMappings: [],
}

const logitech: Keyboard = {
  id: '046d:c33f',
  name: 'Logitech K120',
  transport: 'USB',
  builtIn: false,
  hasGlobe: false,
  trigger: capsFn,
  listed: false,
  triggerMapping: capsToFn,
  otherMappings: [graveToEscape],
}

const report: KeyboardReport = { keyboards: [builtIn, logitech], systemMappings: [graveToEscape] }

function renderSheet(trigger: Partial<Config['trigger']> = {}, state: KeyboardsState = { report, error: null }) {
  const onChange = vi.fn()
  const onClose = vi.fn()
  const merged = { ...config.trigger, ...trigger }

  render(
    <KeyboardsSheet
      open
      config={{ ...config, trigger: merged }}
      keyboards={state}
      onChange={onChange}
      onClose={onClose}
    />,
  )

  return { onChange, onClose, trigger: merged }
}

function chooseKeyboard(name: string): void {
  fireEvent.click(screen.getByRole('combobox', { name: 'Select keyboard' }))
  fireEvent.click(screen.getByRole('option', { name }))
}

afterEach(cleanup)

describe('KeyboardsSheet', () => {
  it('shows the first keyboard using the Globe key by default', () => {
    renderSheet()

    expect(screen.getByRole('dialog', { name: 'Keyboards' })).toBeTruthy()
    expect(screen.getByText('Built-in · 05ac:0281 · has Globe key')).toBeTruthy()
    expect(screen.getByText('🌐 Globe')).toBeTruthy()
    expect((screen.getByRole('combobox', { name: 'Trigger mode' }) as HTMLButtonElement).disabled).toBe(true)
  })

  it('leaves a keyboard without a Globe key unassigned, with Clear instead of Use Globe', () => {
    renderSheet()
    chooseKeyboard('Logitech K120')

    expect(screen.getByText('None')).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Use Globe' })).toBeNull()
    expect((screen.getByRole('button', { name: 'Clear' }) as HTMLButtonElement).disabled).toBe(true)

    cleanup()

    const { onChange, trigger } = renderSheet({ keyboards: { '046d:c33f': capsFn } })
    chooseKeyboard('Logitech K120')
    fireEvent.click(screen.getByRole('button', { name: 'Clear' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('shows another keyboard with its applied mapping', () => {
    renderSheet({ keyboards: { '046d:c33f': capsFn } })

    chooseKeyboard('Logitech K120')

    expect(screen.getByText('USB · 046d:c33f')).toBeTruthy()
    expect(screen.getByText('⇪ Caps Lock')).toBeTruthy()
    expect(screen.getByText('Caps Lock → fn.')).toBeTruthy()
  })

  it('sets a keyboard its own mode', () => {
    const { onChange, trigger } = renderSheet({ keyboards: { '046d:c33f': capsFn } })

    chooseKeyboard('Logitech K120')
    fireEvent.click(screen.getByRole('combobox', { name: 'Trigger mode' }))
    fireEvent.click(screen.getByRole('option', { name: 'Only this key switches' }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { '046d:c33f': { key: 'caps_lock', mode: 'dedicated' } } },
    })
  })

  it('puts a listed keyboard back on the Globe key', () => {
    const insert = { key: 'insert', mode: 'pass_through' }
    const { onChange, trigger } = renderSheet({ keyboards: { '05ac:0281': insert } })

    expect(screen.getByText('Insert')).toBeTruthy()
    expect(screen.getByText('Applies once saved.')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Use Globe' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('notes that the Globe key still switches while another keyboard switches with it', () => {
    renderSheet({ keyboards: { '05ac:0281': { key: 'caps_lock', mode: 'dedicated' }, '046d:c33f': capsFn } })

    expect(screen.getByText(/The Globe key still switches here/)).toBeTruthy()
  })

  it('shows a silenced key', () => {
    const silenced = {
      ...logitech,
      trigger: { key: 'caps_lock', mode: 'dedicated' },
      triggerMapping: { ...capsToFn, dst: 0x700000073, dstName: 'F24' },
    }
    renderSheet(
      { keyboards: { '046d:c33f': { key: 'caps_lock', mode: 'dedicated' } } },
      { report: { keyboards: [silenced], systemMappings: [] }, error: null },
    )

    expect(screen.getByText('Caps Lock silenced.')).toBeTruthy()
  })

  it('lists and forgets keyboards set here that are not connected', () => {
    const { onChange, trigger } = renderSheet({ keyboards: { '1234:5678': { key: 'insert', mode: 'fn' } } })

    chooseKeyboard('1234:5678 (not connected)')
    expect(screen.getByText('Insert, act as fn')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Forget' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('shows other tools’ mappings behind a disclosure', () => {
    renderSheet()

    expect(screen.getByText('Mappings from other tools')).toBeTruthy()
    expect(screen.getAllByText('Grave Accent → Escape')).toHaveLength(2)
    expect(screen.getByText('Every keyboard')).toBeTruthy()
  })

  it('shows a report error and closes with Done', () => {
    const { onClose } = renderSheet({}, { report: null, error: 'unknown method ListKeyboards' })

    expect(screen.getByText('Keyboards unavailable: unknown method ListKeyboards')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(onClose).toHaveBeenCalled()
  })

  it('says when no keyboard is connected', () => {
    renderSheet({}, { report: { keyboards: [], systemMappings: [] }, error: null })

    expect(screen.getByText('No keyboards connected.')).toBeTruthy()
  })
})
