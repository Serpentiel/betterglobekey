// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Select } from './Select'

const options = [
  { value: 'a', label: 'Apple' },
  { value: 'b', label: 'Banana' },
  { value: 'c', label: 'Cherry' },
]

afterEach(cleanup)

describe('Select', () => {
  it('shows the selected option label', () => {
    render(<Select value="b" options={options} onChange={vi.fn()} />)

    expect(screen.getByRole('combobox').textContent).toContain('Banana')
  })

  it('opens and selects an option on click', () => {
    const onChange = vi.fn()
    render(<Select value="a" options={options} onChange={onChange} />)

    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.click(screen.getByRole('option', { name: 'Cherry' }))

    expect(onChange).toHaveBeenCalledWith('c')
  })

  it('navigates with the arrow keys and commits with Enter', () => {
    const onChange = vi.fn()
    render(<Select value="a" options={options} onChange={onChange} />)

    const trigger = screen.getByRole('combobox')
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }) // opens, active = current (a)
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }) // active -> Banana
    fireEvent.keyDown(trigger, { key: 'Enter' })

    expect(onChange).toHaveBeenCalledWith('b')
  })

  it('closes on Escape without selecting', () => {
    const onChange = vi.fn()
    render(<Select value="a" options={options} onChange={onChange} />)

    const trigger = screen.getByRole('combobox')
    fireEvent.keyDown(trigger, { key: 'ArrowDown' })
    expect(screen.queryByRole('listbox')).toBeTruthy()

    fireEvent.keyDown(trigger, { key: 'Escape' })
    expect(screen.queryByRole('listbox')).toBeNull()
    expect(onChange).not.toHaveBeenCalled()
  })

  it('stays closed when disabled', () => {
    render(<Select value="a" options={options} onChange={vi.fn()} disabled />)

    fireEvent.click(screen.getByRole('combobox'))
    expect(screen.queryByRole('listbox')).toBeNull()
  })
})
