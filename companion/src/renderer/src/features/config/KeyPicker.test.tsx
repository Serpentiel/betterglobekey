// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KeyboardTrigger, RecordedKey } from '../../../../shared/types'
import { api } from '../../lib/api'
import { KeyPicker } from './KeyPicker'

vi.mock('../../lib/api', () => ({ api: {} }))

afterEach(cleanup)

const globe: KeyboardTrigger = { key: 'globe', mode: 'dedicated' }
const caps: KeyboardTrigger = { key: 'caps_lock', mode: 'dedicated' }
const insert: KeyboardTrigger = { key: 'insert', mode: 'pass_through' }

function recordKey(result: Promise<RecordedKey>): Api['recordKey'] {
  const fake = vi.fn<Api['recordKey']>(() => result)
  Object.assign(api, { recordKey: fake })

  return fake
}

describe('KeyPicker', () => {
  it('shows every key', () => {
    render(<KeyPicker triggers={[globe, { key: 'right_command', mode: 'dedicated' }]} onChange={vi.fn()} />)

    expect(screen.getByText('🌐 Globe')).toBeTruthy()
    expect(screen.getByText('Right ⌘ Command')).toBeTruthy()
  })

  it('adds a key recorded on the chosen keyboard through the daemon', async () => {
    const fake = recordKey(Promise.resolve({ key: 'insert', keyName: 'Insert', keyboardId: '046d:c33f' }))
    const onChange = vi.fn()
    render(<KeyPicker triggers={[caps]} keyboardId="046d:c33f" keyboardName="Logitech K120" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    expect(screen.getByText('Press a key…')).toBeTruthy()
    expect(screen.getByRole('status').textContent).toBe('Press a key on Logitech K120.')

    await waitFor(() => expect(onChange).toHaveBeenCalledWith([caps, { key: 'insert', mode: 'dedicated' }]))
    expect(fake).toHaveBeenCalledWith('046d:c33f')
    expect(screen.queryByText('Press a key…')).toBeNull()
  })

  it('replaces a key, keeping a mode it allows', async () => {
    recordKey(Promise.resolve({ key: 'f13', keyName: 'F13', keyboardId: '' }))
    const onChange = vi.fn()
    render(<KeyPicker triggers={[globe, insert]} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Change Insert' }))

    await waitFor(() => expect(onChange).toHaveBeenCalledWith([globe, { key: 'f13', mode: 'pass_through' }]))
  })

  it('records from any keyboard without one chosen', async () => {
    const fake = recordKey(Promise.resolve({ key: 'caps_lock', keyName: 'Caps Lock', keyboardId: '046d:c33f' }))
    render(<KeyPicker triggers={[]} onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    expect(screen.getByRole('status').textContent).toBe('Press a key on any keyboard.')

    await waitFor(() => expect(fake).toHaveBeenCalledWith(''))
  })

  it('rejects a key that cannot be a trigger, or is one already', async () => {
    recordKey(Promise.resolve({ key: '', keyName: 'A', keyboardId: '046d:c33f' }))
    const onChange = vi.fn()
    render(<KeyPicker triggers={[caps]} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    await waitFor(() =>
      expect(screen.getByRole('status').textContent).toBe("A can't switch input sources. Try another key."),
    )

    recordKey(Promise.resolve({ key: 'caps_lock', keyName: 'Caps Lock', keyboardId: '046d:c33f' }))
    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    await waitFor(() =>
      expect(screen.getByRole('status').textContent).toBe('Caps Lock already switches input sources here.'),
    )

    expect(onChange).not.toHaveBeenCalled()
  })

  it('shows why recording failed', async () => {
    recordKey(Promise.reject(new Error('enable Input Monitoring')))
    render(<KeyPicker triggers={[]} onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('enable Input Monitoring'))
  })

  it('ignores a key pressed after cancelling', async () => {
    let resolve: (key: RecordedKey) => void = () => undefined
    recordKey(new Promise((done) => (resolve = done)))
    const onChange = vi.fn()
    render(<KeyPicker triggers={[globe]} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Change Globe' }))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    resolve({ key: 'insert', keyName: 'Insert', keyboardId: '' })

    await Promise.resolve()
    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByText('🌐 Globe')).toBeTruthy()
  })

  it('offers the Globe key while recording, unless another keycap has it', () => {
    recordKey(new Promise(() => undefined))
    const onChange = vi.fn()
    render(<KeyPicker triggers={[caps]} onChange={onChange} />)

    expect(screen.queryByRole('button', { name: 'Use Globe key' })).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    fireEvent.click(screen.getByRole('button', { name: 'Use Globe key' }))
    expect(onChange).toHaveBeenCalledWith([caps, globe])

    cleanup()

    render(<KeyPicker triggers={[globe, caps]} onChange={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    expect(screen.queryByRole('button', { name: 'Use Globe key' })).toBeNull()

    cleanup()

    render(<KeyPicker triggers={[]} hasGlobe={false} onChange={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Add a trigger key' }))
    expect(screen.queryByRole('button', { name: 'Use Globe key' })).toBeNull()
  })

  it('removes a key', () => {
    const onChange = vi.fn()
    render(<KeyPicker triggers={[globe, caps]} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove Globe' }))
    expect(onChange).toHaveBeenCalledWith([caps])
  })

  it('names the add button only while there are no keys', () => {
    render(<KeyPicker triggers={[]} onChange={vi.fn()} />)
    expect(screen.getByText('Add key')).toBeTruthy()

    cleanup()

    render(<KeyPicker triggers={[caps]} onChange={vi.fn()} />)
    expect(screen.queryByText('Add key')).toBeNull()
  })
})
