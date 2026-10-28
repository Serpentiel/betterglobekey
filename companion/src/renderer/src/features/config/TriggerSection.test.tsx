// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, Keyboard, KeyboardReport, KeyboardTrigger } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import { api } from '../../lib/api'
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
  triggers: [globe],
  listed: false,
  triggerMappings: [],
  otherMappings: [],
}

const logitech: Keyboard = {
  id: '046d:c33f',
  name: 'Logitech K120',
  transport: 'USB',
  builtIn: false,
  hasGlobe: false,
  triggers: [capsDedicated],
  listed: true,
  triggerMappings: [capsSilenced],
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
    expect(row('Apple Internal Keyboard').queryByRole('combobox')).toBeNull()
    expect(screen.getByRole('group', { name: 'Logitech K120' })).toBeTruthy()
  })

  it('leaves a keyboard without a Globe key without a trigger key', () => {
    renderSection()

    expect(row('Logitech K120').getByText('Add key')).toBeTruthy()
    expect(row('Logitech K120').queryByText('⇪ Caps Lock')).toBeNull()
  })

  it('unlists a keyboard whose last key is removed when it has no Globe key', () => {
    const { onChange, trigger } = renderSection({ keyboards: { '046d:c33f': [capsDedicated] } })

    fireEvent.click(row('Logitech K120').getByRole('button', { name: 'Remove Caps Lock' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('keeps an Apple keyboard listed without keys once its Globe key is removed', () => {
    const { onChange, trigger } = renderSection()

    fireEvent.click(row('Apple Internal Keyboard').getByRole('button', { name: 'Remove Globe' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: { '05ac:0281': [] } } })
  })

  it('shows a keyboard with its applied mapping', () => {
    renderSection({ keyboards: { '046d:c33f': [capsDedicated] } })

    expect(row('Logitech K120').getByText('USB · 046d:c33f · Caps Lock silenced')).toBeTruthy()
    expect(row('Logitech K120').getByText('⇪ Caps Lock')).toBeTruthy()
  })

  it('lists several keys on one keyboard, the Globe key among them, each other key with its own mode', () => {
    const insert = { key: 'insert', mode: 'dedicated' }
    const { onChange, trigger } = renderSection({ keyboards: { '05ac:0281': [globe, capsDedicated, insert] } })

    const apple = row('Apple Internal Keyboard')
    expect(apple.getByText('🌐 Globe')).toBeTruthy()
    expect(apple.getByText('⇪ Caps Lock')).toBeTruthy()
    expect(apple.getByText('Insert', { selector: 'span[aria-live]' })).toBeTruthy()
    expect((apple.getByRole('combobox', { name: 'Caps Lock' }) as HTMLButtonElement).disabled).toBe(true)

    fireEvent.click(apple.getByRole('combobox', { name: 'Insert' }))
    fireEvent.click(screen.getByRole('option', { name: "Keep the key's function" }))

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      trigger: {
        ...trigger,
        keyboards: { '05ac:0281': [globe, capsDedicated, { key: 'insert', mode: 'pass_through' }] },
      },
    })
  })

  it('unlists a keyboard set back to its Globe key alone', () => {
    const insert = { key: 'insert', mode: 'pass_through' }
    const { onChange, trigger } = renderSection({ keyboards: { '05ac:0281': [globe, insert] } })

    expect(row('Apple Internal Keyboard').getByText('Built-in · 05ac:0281 · applies once saved')).toBeTruthy()

    fireEvent.click(row('Apple Internal Keyboard').getByRole('button', { name: 'Remove Insert' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  describe('with two keyboards that have a Globe key', () => {
    const magic: Keyboard = { ...builtIn, id: '05ac:029c', name: 'Magic Keyboard', builtIn: false, transport: 'USB' }
    const twoApple: KeyboardsState = {
      report: { keyboards: [builtIn, magic, logitech], systemMappings: [] },
      error: null,
    }

    it('removes the Globe key from every keyboard with one', () => {
      const { onChange, trigger } = renderSection({ keyboards: { '05ac:029c': [globe, capsDedicated] } }, twoApple)

      fireEvent.click(row('Apple Internal Keyboard').getByRole('button', { name: 'Remove Globe from every keyboard' }))
      expect(onChange).toHaveBeenCalledWith({
        ...config,
        trigger: { ...trigger, keyboards: { '05ac:0281': [], '05ac:029c': [capsDedicated] } },
      })
    })

    it('adds the Globe key back to every keyboard with one, unlisting those it leaves at their default', async () => {
      Object.assign(api, { recordKey: vi.fn(() => new Promise(() => undefined)) })
      const { onChange, trigger } = renderSection(
        { keyboards: { '05ac:0281': [], '05ac:029c': [capsDedicated], '004c:0267': [] } },
        twoApple,
      )

      fireEvent.click(row('Magic Keyboard').getByRole('button', { name: 'Add a trigger key' }))
      fireEvent.click(row('Magic Keyboard').getByRole('button', { name: 'Use Globe key on every keyboard' }))

      expect(onChange).toHaveBeenCalledWith({
        ...config,
        trigger: { ...trigger, keyboards: { '05ac:029c': [capsDedicated, globe] } },
      })
    })

    it('leaves keyboards without a Globe key alone', () => {
      const { onChange } = renderSection({ keyboards: { '046d:c33f': [capsDedicated] } }, twoApple)

      fireEvent.click(row('Magic Keyboard').getByRole('button', { name: 'Remove Globe from every keyboard' }))
      expect(onChange.mock.calls[0][0].trigger.keyboards['046d:c33f']).toEqual([capsDedicated])
    })
  })

  it('names Globe controls for this keyboard alone when no other keyboard has a Globe key', () => {
    renderSection()

    expect(row('Apple Internal Keyboard').getByRole('button', { name: 'Remove Globe' })).toBeTruthy()
  })

  it('notes that the Globe key still switches while another keyboard switches with it', () => {
    renderSection({ keyboards: { '05ac:0281': [capsDedicated], '046d:c33f': [globe] } })

    expect(row('Apple Internal Keyboard').getByText(/The Globe key still switches here/)).toBeTruthy()
  })

  it('folds keyboards set here that are not connected, and forgets them', () => {
    const { onChange, trigger } = renderSection({
      keyboards: { '1234:5678': [{ key: 'insert', mode: 'dedicated' }, globe] },
    })

    fireEvent.click(screen.getByRole('button', { name: 'Other keyboards (1)' }))
    expect(screen.getByText('Insert (switch input sources only), Globe key · applies when connected')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Forget' }))
    expect(onChange).toHaveBeenCalledWith({ ...config, trigger: { ...trigger, keyboards: {} } })
  })

  it('shows other tools’ mappings behind a disclosure', () => {
    renderSection()

    expect(screen.queryByText('Every keyboard')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Mappings from other tools' }))
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
