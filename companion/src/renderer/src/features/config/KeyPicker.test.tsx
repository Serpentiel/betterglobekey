// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, RecordedKey } from '../../../../shared/types'
import { api } from '../../lib/api'
import { KeyPicker } from './KeyPicker'

vi.mock('../../lib/api', () => ({ api: {} }))

afterEach(cleanup)

function recordKey(result: Promise<RecordedKey>): Api['recordKey'] {
  const fake = vi.fn<Api['recordKey']>(() => result)
  Object.assign(api, { recordKey: fake })

  return fake
}

describe('KeyPicker', () => {
  it('shows the current key', () => {
    render(<KeyPicker value="right_command" onChange={vi.fn()} />)

    expect(screen.getByText('Right ⌘ Command')).toBeTruthy()
  })

  it('records the next key on the chosen keyboard through the daemon', async () => {
    const fake = recordKey(Promise.resolve({ key: 'insert', keyName: 'Insert', keyboardId: '046d:c33f' }))
    const onChange = vi.fn()
    render(<KeyPicker value="globe" keyboardId="046d:c33f" keyboardName="Logitech K120" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record a trigger key' }))
    expect(screen.getByText('Press a key…')).toBeTruthy()
    expect(screen.getByRole('status').textContent).toBe('Press a key on Logitech K120.')

    await waitFor(() => expect(onChange).toHaveBeenCalledWith('insert'))
    expect(fake).toHaveBeenCalledWith('046d:c33f')
    expect(screen.queryByText('Press a key…')).toBeNull()
  })

  it('records from any keyboard without one chosen', async () => {
    const fake = recordKey(Promise.resolve({ key: 'caps_lock', keyName: 'Caps Lock', keyboardId: '046d:c33f' }))
    render(<KeyPicker value="globe" onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record a trigger key' }))
    expect(screen.getByRole('status').textContent).toBe('Press a key on any keyboard.')

    await waitFor(() => expect(fake).toHaveBeenCalledWith(''))
  })

  it('rejects a key that cannot be a trigger', async () => {
    recordKey(Promise.resolve({ key: '', keyName: 'A', keyboardId: '046d:c33f' }))
    const onChange = vi.fn()
    render(<KeyPicker value="globe" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record a trigger key' }))

    await waitFor(() =>
      expect(screen.getByRole('status').textContent).toBe("A can't switch input sources. Try another key."),
    )
    expect(onChange).not.toHaveBeenCalled()
  })

  it('shows why recording failed', async () => {
    recordKey(Promise.reject(new Error('enable Input Monitoring')))
    render(<KeyPicker value="globe" onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record a trigger key' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('enable Input Monitoring'))
  })

  it('ignores a key pressed after cancelling', async () => {
    let resolve: (key: RecordedKey) => void = () => undefined
    recordKey(new Promise((done) => (resolve = done)))
    const onChange = vi.fn()
    render(<KeyPicker value="globe" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record a trigger key' }))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    resolve({ key: 'insert', keyName: 'Insert', keyboardId: '' })

    await Promise.resolve()
    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByText('🌐 Globe')).toBeTruthy()
  })

  it('switches back to the Globe key', () => {
    const onChange = vi.fn()
    render(<KeyPicker value="caps_lock" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Use Globe key' }))
    expect(onChange).toHaveBeenCalledWith('globe')
  })

  it('hides Use Globe on the Globe key, and offers Clear only without a Globe key', () => {
    render(<KeyPicker value="globe" onChange={vi.fn()} />)
    expect(screen.queryByRole('button', { name: 'Use Globe key' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Clear' })).toBeNull()

    cleanup()

    const onChange = vi.fn()
    render(<KeyPicker value="caps_lock" hasGlobe={false} onChange={onChange} />)
    expect(screen.queryByRole('button', { name: 'Use Globe key' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Clear' }))
    expect(onChange).toHaveBeenCalledWith('')
  })

  it('disables recording', () => {
    render(<KeyPicker value="caps_lock" disabled onChange={vi.fn()} />)

    expect((screen.getByRole('button', { name: 'Record a trigger key' }) as HTMLButtonElement).disabled).toBe(true)
  })
})
