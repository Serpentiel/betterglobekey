// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Config, InputSource } from '../../../../shared/types'
import type { ConfigErrors } from '../../lib/config-schema'
import { CollectionsSection } from './CollectionsSection'

const sources: InputSource[] = [{ id: 'us', name: 'U.S.' }]

const nameOf = (id: string): string => sources.find((source) => source.id === id)?.name ?? id

const noErrors: ConfigErrors = { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {} }

const baseConfig: Config = {
  logger: { path: '/tmp/x.log', level: 'info', retentionDays: 7, retentionFiles: 3 },
  doublePress: { enabled: true, maximumDelay: '250ms' },
  reverse: { enabled: true, modifier: 'shift' },
  hud: { enabled: true, duration: '900ms', showCollection: true },
  collections: [{ name: 'primary', sources: ['us'] }],
}

function renderSection(onChange = vi.fn()) {
  render(
    <CollectionsSection config={baseConfig} sources={sources} errors={noErrors} nameOf={nameOf} onChange={onChange} />,
  )

  return onChange
}

afterEach(cleanup)

describe('CollectionsSection', () => {
  it('renders a card per configured collection', () => {
    renderSection()

    expect(screen.getByDisplayValue('primary')).toBeTruthy()
  })

  it('appends a new, empty collection on Add', () => {
    const onChange = renderSection()

    fireEvent.click(screen.getByRole('button', { name: 'Add' }))

    expect(onChange).toHaveBeenCalledWith({
      ...baseConfig,
      collections: [
        { name: 'primary', sources: ['us'] },
        { name: 'collection 2', sources: [] },
      ],
    })
  })

  it('replaces a collection when its card reports a change', () => {
    const onChange = renderSection()

    fireEvent.change(screen.getByLabelText('Collection name'), { target: { value: 'work' } })

    expect(onChange).toHaveBeenCalledWith({ ...baseConfig, collections: [{ name: 'work', sources: ['us'] }] })
  })

  it('removes a collection', () => {
    const onChange = renderSection()

    fireEvent.click(screen.getByRole('button', { name: 'Remove collection' }))

    expect(onChange).toHaveBeenCalledWith({ ...baseConfig, collections: [] })
  })

  it('reorders collections', () => {
    const config: Config = {
      ...baseConfig,
      collections: [
        { name: 'a', sources: ['us'] },
        { name: 'b', sources: [] },
      ],
    }
    const onChange = vi.fn()
    render(
      <CollectionsSection config={config} sources={sources} errors={noErrors} nameOf={nameOf} onChange={onChange} />,
    )

    fireEvent.click(screen.getAllByRole('button', { name: 'Move collection down' })[0])

    expect(onChange).toHaveBeenCalledWith({
      ...config,
      collections: [
        { name: 'b', sources: [] },
        { name: 'a', sources: ['us'] },
      ],
    })
  })
})
