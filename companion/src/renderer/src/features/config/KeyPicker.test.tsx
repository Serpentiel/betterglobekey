// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { KeyPicker } from './KeyPicker'

afterEach(cleanup)

function press(code: string, key = code, type: 'keydown' | 'keyup' = 'keydown'): void {
  act(() => {
    window.dispatchEvent(new KeyboardEvent(type, { code, key, bubbles: true, cancelable: true }))
  })
}

describe('KeyPicker', () => {
  it('shows the current key', () => {
    render(<KeyPicker value="right_command" remapped={false} onChange={vi.fn()} />)

    expect(screen.getByText('Right ⌘ Command')).toBeTruthy()
  })

  it('records the next supported key', () => {
    const onChange = vi.fn()
    render(<KeyPicker value="globe" remapped={false} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record' }))
    expect(screen.getByText('Press a key…')).toBeTruthy()

    press('MetaRight', 'Meta')
    expect(onChange).toHaveBeenCalledWith('right_command')
    expect(screen.queryByText('Press a key…')).toBeNull()
  })

  it('takes Caps Lock from its key-up too', () => {
    const onChange = vi.fn()
    render(<KeyPicker value="globe" remapped={false} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record' }))
    press('KeyA', 'a', 'keyup')
    expect(onChange).not.toHaveBeenCalled()

    press('CapsLock', 'CapsLock', 'keyup')
    expect(onChange).toHaveBeenCalledWith('caps_lock')
  })

  it('rejects unsupported keys and keeps recording', () => {
    const onChange = vi.fn()
    render(<KeyPicker value="globe" remapped={false} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record' }))
    press('KeyA', 'a')

    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByRole('status').textContent).toContain("a can't switch input sources")

    press('F13')
    expect(onChange).toHaveBeenCalledWith('f13')
  })

  it('cancels with Escape', () => {
    const onChange = vi.fn()
    render(<KeyPicker value="f13" remapped={false} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record' }))
    press('Escape')

    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByText('F13')).toBeTruthy()
  })

  it('explains that a remapped key cannot be recorded', () => {
    render(<KeyPicker value="caps_lock" remapped onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Record' }))

    expect(screen.getByRole('status').textContent).toContain('Caps Lock is remapped right now')
  })

  it('resets to Globe, which cannot be recorded', () => {
    const onChange = vi.fn()
    const { rerender } = render(<KeyPicker value="f13" remapped={false} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Use Globe' }))
    expect(onChange).toHaveBeenCalledWith('globe')

    rerender(<KeyPicker value="globe" remapped={false} onChange={onChange} />)
    expect((screen.getByRole('button', { name: 'Use Globe' }) as HTMLButtonElement).disabled).toBe(true)
  })
})
