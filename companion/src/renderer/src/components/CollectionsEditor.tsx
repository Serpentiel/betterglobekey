import { useState } from 'react'

import type { Collection, InputSource } from '../../../shared/types'

interface Props {
  collections: Collection[]
  sources: InputSource[]
  nameOf: (id: string) => string
  onChange: (collections: Collection[]) => void
}

// move returns a copy of items with the element at index shifted by delta.
function move<T>(items: T[], index: number, delta: number): T[] {
  const target = index + delta

  if (target < 0 || target >= items.length) {
    return items
  }

  const next = [...items]
  ;[next[index], next[target]] = [next[target], next[index]]

  return next
}

export function CollectionsEditor({ collections, sources, nameOf, onChange }: Props): JSX.Element {
  const replace = (index: number, collection: Collection): void => {
    onChange(collections.map((existing, position) => (position === index ? collection : existing)))
  }

  const addCollection = (): void => {
    onChange([...collections, { name: `collection-${collections.length + 1}`, sources: [] }])
  }

  return (
    <div className="collections">
      {collections.map((collection, index) => (
        <CollectionRow
          key={index}
          collection={collection}
          sources={sources}
          nameOf={nameOf}
          canMoveUp={index > 0}
          canMoveDown={index < collections.length - 1}
          onMove={(delta) => onChange(move(collections, index, delta))}
          onRemove={() => onChange(collections.filter((_, position) => position !== index))}
          onChange={(updated) => replace(index, updated)}
        />
      ))}

      <button type="button" className="secondary" onClick={addCollection}>
        Add collection
      </button>
    </div>
  )
}

interface RowProps {
  collection: Collection
  sources: InputSource[]
  nameOf: (id: string) => string
  canMoveUp: boolean
  canMoveDown: boolean
  onMove: (delta: number) => void
  onRemove: () => void
  onChange: (collection: Collection) => void
}

function CollectionRow({
  collection,
  sources,
  nameOf,
  canMoveUp,
  canMoveDown,
  onMove,
  onRemove,
  onChange,
}: RowProps): JSX.Element {
  const [toAdd, setToAdd] = useState('')

  const available = sources.filter((source) => !collection.sources.includes(source.id))

  const addSource = (): void => {
    if (toAdd && !collection.sources.includes(toAdd)) {
      onChange({ ...collection, sources: [...collection.sources, toAdd] })
      setToAdd('')
    }
  }

  return (
    <div className="collection">
      <div className="collection__head">
        <input
          type="text"
          className="collection__name"
          value={collection.name}
          onChange={(event) => onChange({ ...collection, name: event.target.value })}
          placeholder="Collection name"
        />
        <div className="collection__controls">
          <button type="button" className="icon" disabled={!canMoveUp} onClick={() => onMove(-1)} title="Move up">
            ↑
          </button>
          <button type="button" className="icon" disabled={!canMoveDown} onClick={() => onMove(1)} title="Move down">
            ↓
          </button>
          <button type="button" className="icon icon--danger" onClick={onRemove} title="Remove collection">
            ✕
          </button>
        </div>
      </div>

      <ul className="sources">
        {collection.sources.map((id, index) => (
          <li key={id} className="sources__item">
            <span className="sources__name">{nameOf(id)}</span>
            <code className="sources__id">{id}</code>
            <div className="sources__controls">
              <button
                type="button"
                className="icon"
                disabled={index === 0}
                onClick={() => onChange({ ...collection, sources: move(collection.sources, index, -1) })}
                title="Move up"
              >
                ↑
              </button>
              <button
                type="button"
                className="icon"
                disabled={index === collection.sources.length - 1}
                onClick={() => onChange({ ...collection, sources: move(collection.sources, index, 1) })}
                title="Move down"
              >
                ↓
              </button>
              <button
                type="button"
                className="icon icon--danger"
                onClick={() =>
                  onChange({ ...collection, sources: collection.sources.filter((_, position) => position !== index) })
                }
                title="Remove source"
              >
                ✕
              </button>
            </div>
          </li>
        ))}
        {collection.sources.length === 0 ? <li className="muted">No sources yet.</li> : null}
      </ul>

      <div className="sources__add">
        <select value={toAdd} onChange={(event) => setToAdd(event.target.value)}>
          <option value="">Add an input source…</option>
          {available.map((source) => (
            <option key={source.id} value={source.id}>
              {source.name} ({source.id})
            </option>
          ))}
        </select>
        <button type="button" className="secondary" disabled={!toAdd} onClick={addSource}>
          Add
        </button>
      </div>
    </div>
  )
}
