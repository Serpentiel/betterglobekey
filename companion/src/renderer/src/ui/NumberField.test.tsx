// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { NumberField } from './NumberField'

afterEach(cleanup)

describe('NumberField', () => {
  it('increments and decrements by the step', () => {
    const onChange = vi.fn()
    render(<NumberField value={5} min={0} max={10} onChange={onChange} />)

    fireEvent.click(screen.getByLabelText('Increase'))
    expect(onChange).toHaveBeenCalledWith(6)

    fireEvent.click(screen.getByLabelText('Decrease'))
    expect(onChange).toHaveBeenCalledWith(4)
  })

  it('clamps typed values to the min/max range', () => {
    const onChange = vi.fn()
    render(<NumberField value={5} min={0} max={10} onChange={onChange} />)

    const input = screen.getByRole('spinbutton')

    fireEvent.change(input, { target: { value: '-3' } })
    expect(onChange).toHaveBeenCalledWith(0)

    fireEvent.change(input, { target: { value: '42' } })
    expect(onChange).toHaveBeenCalledWith(10)
  })

  it('disables the steppers at the bounds', () => {
    const { rerender } = render(<NumberField value={0} min={0} max={10} onChange={vi.fn()} />)
    expect((screen.getByLabelText('Decrease') as HTMLButtonElement).disabled).toBe(true)

    rerender(<NumberField value={10} min={0} max={10} onChange={vi.fn()} />)
    expect((screen.getByLabelText('Increase') as HTMLButtonElement).disabled).toBe(true)
  })
})
