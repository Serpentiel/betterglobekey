// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sheet } from './Sheet'

afterEach(cleanup)

describe('Sheet', () => {
  it('renders nothing while closed', () => {
    render(
      <Sheet open={false} title="Keyboards" onClose={vi.fn()}>
        body
      </Sheet>,
    )

    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('closes with Done, Esc, and the backdrop', () => {
    const onClose = vi.fn()
    render(
      <Sheet open title="Keyboards" description="Where the key is remapped." onClose={onClose}>
        body
      </Sheet>,
    )

    const dialog = screen.getByRole('dialog', { name: 'Keyboards' })
    expect(screen.getByText('Where the key is remapped.')).toBeTruthy()
    expect(document.activeElement).toBe(dialog)

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    fireEvent.keyDown(document, { key: 'Escape' })
    fireEvent.pointerDown(dialog.parentElement as HTMLElement)
    fireEvent.pointerDown(dialog)

    expect(onClose).toHaveBeenCalledTimes(3)
  })
})
