// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Switch } from './Switch'

afterEach(cleanup)

describe('Switch', () => {
  it('reflects the checked state via aria-checked', () => {
    render(<Switch checked label="Toggle" onChange={vi.fn()} />)

    expect(screen.getByRole('switch').getAttribute('aria-checked')).toBe('true')
  })

  it('reports the inverted value when clicked', () => {
    const onChange = vi.fn()
    render(<Switch checked={false} label="Toggle" onChange={onChange} />)

    fireEvent.click(screen.getByRole('switch', { name: 'Toggle' }))
    expect(onChange).toHaveBeenCalledWith(true)
  })
})
