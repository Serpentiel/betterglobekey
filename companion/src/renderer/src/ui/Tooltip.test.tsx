// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Tooltip } from './Tooltip'

beforeEach(() => vi.useFakeTimers())

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

function hover(name: string): void {
  fireEvent.pointerEnter(screen.getByRole('button', { name }).parentElement as HTMLElement)
}

function leave(name: string): void {
  fireEvent.pointerLeave(screen.getByRole('button', { name }).parentElement as HTMLElement)
}

describe('Tooltip', () => {
  it('shows after a short rest, describes its element, and hides on leave', () => {
    render(
      <Tooltip label="Use the Globe key">
        <button type="button">Globe</button>
      </Tooltip>,
    )

    hover('Globe')
    expect(screen.queryByRole('tooltip')).toBeNull()

    act(() => vi.advanceTimersByTime(400))
    const tooltip = screen.getByRole('tooltip')
    expect(tooltip.textContent).toBe('Use the Globe key')
    expect(screen.getByRole('button', { name: 'Globe' }).getAttribute('aria-describedby')).toBe(tooltip.id)

    leave('Globe')
    expect(screen.queryByRole('tooltip')).toBeNull()
  })

  it('shows the next tooltip at once while warm', () => {
    render(
      <>
        <Tooltip label="First">
          <button type="button">One</button>
        </Tooltip>
        <Tooltip label="Second">
          <button type="button">Two</button>
        </Tooltip>
      </>,
    )

    hover('One')
    act(() => vi.advanceTimersByTime(400))
    leave('One')

    hover('Two')
    expect(screen.getByRole('tooltip').textContent).toBe('Second')
  })

  it('stays hidden when disabled, and hides on press', () => {
    const { rerender } = render(
      <Tooltip label="Hint" disabled>
        <button type="button">Key</button>
      </Tooltip>,
    )

    hover('Key')
    act(() => vi.advanceTimersByTime(1000))
    expect(screen.queryByRole('tooltip')).toBeNull()

    rerender(
      <Tooltip label="Hint">
        <button type="button">Key</button>
      </Tooltip>,
    )
    act(() => vi.advanceTimersByTime(1000))
    hover('Key')
    act(() => vi.advanceTimersByTime(400))
    expect(screen.getByRole('tooltip')).toBeTruthy()

    fireEvent.pointerDown(screen.getByRole('button', { name: 'Key' }))
    expect(screen.queryByRole('tooltip')).toBeNull()
  })
})
