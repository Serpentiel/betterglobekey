// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard, KeyboardReport, KeyboardTrigger } from '../../../../shared/types'
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

const globe: KeyboardTrigger = { key: 'globe', mode: 'dedicated' }
const capsDedicated: KeyboardTrigger = { key: 'caps_lock', mode: 'dedicated' }
const capsSilenced = { src: 0x700000039, dst: 0x700000073, srcName: 'Caps Lock', dstName: 'F24' }
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
  trigger: capsDedicated,
  listed: false,
  triggerMapping: capsSilenced,
  otherMappings: [graveToEscape],
}

const report: KeyboardReport = { keyboards: [builtIn, logitech], systemMappings: [graveToEscape] }

function renderSection(trigger: Partial<Config['trigger']> = {}, state: KeyboardsState = { report, error: null }) {
  const onChange = vi.fn()
  const merged = { ...config.trigger, ...trigger }

  render(
    <TriggerSection config={{ ...config, trigger: merged }} errors={noErrors} keyboards={state} onChange={onChange} />,
  )

  return { onChange, trigger: merged }
}

const row = (name: string) => within(screen.getByRole('group', { name }))

afterEach(cleanup)

describe('TriggerSection', () => {
  it('lists every keyboard inline, a Globe keyboard on its Globe key without a mode', () => {
    renderSection()

    expect(row('Apple Internal Keyboard').getByText('Built-in · 05ac:0281')).toBeTruthy()
    expect(row('Apple Internal Keyboard').getByText('🌐 Globe')).toBeTruthy()
    expect(row('Apple Internal Keyboard').queryByRole('button', { name: 'Use Globe' })).toBeNull()
    expect(row('Apple Internal Keyboard').queryByRole('combobox', { name: 'Mode' })).toBeNull()
    expect(screen.getByRole('group', { name: 'Logitech K120' })).toBeTruthy()
  })

  it('leaves a keyboard without a Globe key unassigned, with Clear instead of Use Globe', () => {
    renderSection()

    expect(row('Logitech K120').getByText('None')).toBeTruthy()
    expect(row('Logitech K120').queryByRole('button', { name: 'Use Globe' })).toBeNull()
    expect(row('Logitech K120').queryByRole('button', { name: 'Clear' })).toBeNull()

    cleanup()

    const { onChange, trigger } = renderSection({ keyboards: { '046d:c33f': capsDedicated } })
    fireEvent.click(row('Logitech K120').getByRole('button', { name: 'Clear' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('shows a keyboard with its applied mapping', () => {
    renderSection({ keyboards: { '046d:c33f': capsDedicated } })

    expect(row('Logitech K120').getByText('USB · 046d:c33f · Caps Lock silenced')).toBeTruthy()
    expect(row('Logitech K120').getByText('⇪ Caps Lock')).toBeTruthy()
  })

  it('sets a keyboard its own mode', () => {
    const { onChange, trigger } = renderSection({ keyboards: { '046d:c33f': { key: 'insert', mode: 'dedicated' } } })

    fireEvent.click(row('Logitech K120').getByRole('combobox', { name: 'Mode' }))
    fireEvent.click(screen.getByRole('option', { name: "Keep the key's function" }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: { ...trigger, keyboards: { '046d:c33f': { key: 'insert', mode: 'pass_through' } } },
    })
  })

  it('puts a listed keyboard back on the Globe key', () => {
    const insert = { key: 'insert', mode: 'pass_through' }
    const { onChange, trigger } = renderSection({ keyboards: { '05ac:0281': insert } })

    expect(row('Apple Internal Keyboard').getByText('Insert')).toBeTruthy()
    expect(row('Apple Internal Keyboard').getByText('Built-in · 05ac:0281 · applies once saved')).toBeTruthy()

    fireEvent.click(row('Apple Internal Keyboard').getByRole('button', { name: 'Use Globe' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('notes that the Globe key still switches while another keyboard switches with it', () => {
    renderSection({ keyboards: { '05ac:0281': capsDedicated, '046d:c33f': globe } })

    expect(row('Apple Internal Keyboard').getByText(/The Globe key still switches here/)).toBeTruthy()
  })

  it('folds keyboards set here that are not connected, and forgets them', () => {
    const { onChange, trigger } = renderSection({ keyboards: { '1234:5678': { key: 'insert', mode: 'dedicated' } } })

    expect(screen.getByText('Other keyboards (1)')).toBeTruthy()
    expect(screen.getByText('Insert, switch input sources only · applies when connected')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Forget' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('shows other tools’ mappings behind a disclosure', () => {
    renderSection()

    expect(screen.getByText('Mappings from other tools')).toBeTruthy()
    expect(screen.getAllByText('Grave Accent → Escape')).toHaveLength(2)
    expect(screen.getByText('Every keyboard')).toBeTruthy()
  })

  it('reports loading, errors and no keyboards', () => {
    renderSection({}, { report: null, error: null })
    expect(screen.getByText('Checking connected keyboards…')).toBeTruthy()

    cleanup()

    renderSection({}, { report: null, error: 'daemon gone' })
    expect(screen.getByText('Keyboards unavailable: daemon gone')).toBeTruthy()

    cleanup()

    renderSection({}, { report: { keyboards: [], systemMappings: [] }, error: null })
    expect(screen.getByText('No keyboards connected.')).toBeTruthy()
  })
})
