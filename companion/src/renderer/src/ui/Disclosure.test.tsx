// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Disclosure } from './Disclosure'

afterEach(cleanup)

describe('Disclosure', () => {
  it('folds its content away until opened', () => {
    render(<Disclosure title="More">Hidden content</Disclosure>)

    const toggle = screen.getByRole('button', { name: 'More' })
    expect(toggle.getAttribute('aria-expanded')).toBe('false')
    expect(screen.queryByText('Hidden content')).toBeNull()

    fireEvent.click(toggle)
    expect(toggle.getAttribute('aria-expanded')).toBe('true')
    expect(screen.getByText('Hidden content')).toBeTruthy()

    fireEvent.click(toggle)
    expect(screen.queryByText('Hidden content')).toBeNull()
  })
})
