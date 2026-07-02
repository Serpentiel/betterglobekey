// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Collection, InputSource } from '../../../../shared/types'
import { CollectionCard } from './CollectionCard'

const sources: InputSource[] = [
  { id: 'us', name: 'U.S.' },
  { id: 'de', name: 'German' },
  { id: 'fr', name: 'French' },
]

const nameOf = (id: string): string => sources.find((source) => source.id === id)?.name ?? id

interface Overrides {
  collection?: Collection
  index?: number
  total?: number
  onChange?: (collection: Collection) => void
  onMove?: (delta: number) => void
  onRemove?: () => void
}

function renderCard(overrides: Overrides = {}) {
  const props = {
    collection: overrides.collection ?? { name: 'primary', sources: ['us'] },
    index: overrides.index ?? 0,
    total: overrides.total ?? 2,
    onChange: overrides.onChange ?? vi.fn(),
    onMove: overrides.onMove ?? vi.fn(),
    onRemove: overrides.onRemove ?? vi.fn(),
  }

  render(
    <CollectionCard
      collection={props.collection}
      sources={sources}
      index={props.index}
      total={props.total}
      nameOf={nameOf}
      onChange={props.onChange}
      onMove={props.onMove}
      onRemove={props.onRemove}
    />,
  )

  return props
}

afterEach(cleanup)

describe('CollectionCard', () => {
  it('renames the collection', () => {
    const onChange = vi.fn()
    renderCard({ onChange })

    fireEvent.change(screen.getByLabelText('Collection name'), { target: { value: 'work' } })

    expect(onChange).toHaveBeenCalledWith({ name: 'work', sources: ['us'] })
  })

  it('removes the collection', () => {
    const onRemove = vi.fn()
    renderCard({ onRemove })

    fireEvent.click(screen.getByRole('button', { name: 'Remove collection' }))

    expect(onRemove).toHaveBeenCalledTimes(1)
  })

  it('disables move-up on the first card and moves down', () => {
    const onMove = vi.fn()
    renderCard({ index: 0, total: 2, onMove })

    expect((screen.getByRole('button', { name: 'Move collection up' }) as HTMLButtonElement).disabled).toBe(true)

    fireEvent.click(screen.getByRole('button', { name: 'Move collection down' }))
    expect(onMove).toHaveBeenCalledWith(1)
  })

  it('removes an input source', () => {
    const onChange = vi.fn()
    renderCard({ collection: { name: 'primary', sources: ['us', 'de'] }, onChange })

    fireEvent.click(screen.getAllByRole('button', { name: 'Remove source' })[0])

    expect(onChange).toHaveBeenCalledWith({ name: 'primary', sources: ['de'] })
  })

  it('reorders input sources', () => {
    const onChange = vi.fn()
    renderCard({ collection: { name: 'primary', sources: ['us', 'de'] }, onChange })

    fireEvent.click(screen.getAllByRole('button', { name: 'Move source down' })[0])

    expect(onChange).toHaveBeenCalledWith({ name: 'primary', sources: ['de', 'us'] })
  })

  it('adds an input source chosen from the picker', () => {
    const onChange = vi.fn()
    renderCard({ collection: { name: 'primary', sources: ['us'] }, onChange })

    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.click(screen.getByRole('option', { name: 'German — de' }))
    fireEvent.click(screen.getByRole('button', { name: 'Add input source' }))

    expect(onChange).toHaveBeenCalledWith({ name: 'primary', sources: ['us', 'de'] })
  })

  it('shows an empty hint when the collection has no sources', () => {
    renderCard({ collection: { name: 'primary', sources: [] } })

    expect(screen.getByText('No input sources yet.')).toBeTruthy()
  })
})
