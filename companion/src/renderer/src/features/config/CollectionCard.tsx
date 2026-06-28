import { AlertCircle, ArrowDown, ArrowUp, Plus, Trash2 } from 'lucide-react'
import { type ReactElement, useState } from 'react'

import type { Collection, InputSource } from '../../../../shared/types'
import { move } from '../../lib/array'
import { Card, IconButton, Input, Select, Stack } from '../../ui'
import styles from './CollectionCard.module.css'

interface Props {
  collection: Collection
  sources: InputSource[]
  index: number
  total: number
  nameError?: string
  sourcesError?: string
  nameOf: (id: string) => string
  onChange: (collection: Collection) => void
  onMove: (delta: number) => void
  onRemove: () => void
}

export function CollectionCard({
  collection,
  sources,
  index,
  total,
  nameError,
  sourcesError,
  nameOf,
  onChange,
  onMove,
  onRemove,
}: Props): ReactElement {
  const [toAdd, setToAdd] = useState('')

  const available = sources.filter((source) => !collection.sources.includes(source.id))

  const addSource = (): void => {
    if (toAdd && !collection.sources.includes(toAdd)) {
      onChange({ ...collection, sources: [...collection.sources, toAdd] })
      setToAdd('')
    }
  }

  const setSources = (next: string[]): void => onChange({ ...collection, sources: next })

  return (
    <Card>
      <Stack gap={12}>
        <Stack direction="row" gap={8} align="center">
          <Input
            value={collection.name}
            invalid={Boolean(nameError)}
            placeholder="Collection name"
            aria-label="Collection name"
            onChange={(event) => onChange({ ...collection, name: event.target.value })}
          />
          <IconButton icon={ArrowUp} label="Move collection up" disabled={index === 0} onClick={() => onMove(-1)} />
          <IconButton
            icon={ArrowDown}
            label="Move collection down"
            disabled={index === total - 1}
            onClick={() => onMove(1)}
          />
          <IconButton icon={Trash2} label="Remove collection" tone="danger" onClick={onRemove} />
        </Stack>

        {nameError ? (
          <p className={styles.error}>
            <AlertCircle size={13} strokeWidth={2.25} />
            {nameError}
          </p>
        ) : null}

        {collection.sources.length > 0 ? (
          <ul className={styles.sources}>
            {collection.sources.map((id, sourceIndex) => (
              <li key={id} className={styles.source}>
                <span className={styles.sourceName}>{nameOf(id)}</span>
                <code className={styles.sourceId}>{id}</code>
                <Stack direction="row" gap={2}>
                  <IconButton
                    icon={ArrowUp}
                    label="Move source up"
                    disabled={sourceIndex === 0}
                    onClick={() => setSources(move(collection.sources, sourceIndex, -1))}
                  />
                  <IconButton
                    icon={ArrowDown}
                    label="Move source down"
                    disabled={sourceIndex === collection.sources.length - 1}
                    onClick={() => setSources(move(collection.sources, sourceIndex, 1))}
                  />
                  <IconButton
                    icon={Trash2}
                    label="Remove source"
                    tone="danger"
                    onClick={() => setSources(collection.sources.filter((_, position) => position !== sourceIndex))}
                  />
                </Stack>
              </li>
            ))}
          </ul>
        ) : (
          <p className={styles.empty}>No input sources yet.</p>
        )}

        {sourcesError ? (
          <p className={styles.error}>
            <AlertCircle size={13} strokeWidth={2.25} />
            {sourcesError}
          </p>
        ) : null}

        <Stack direction="row" gap={8}>
          <Select
            value={toAdd}
            placeholder={available.length ? 'Add an input source…' : 'All input sources added'}
            options={available.map((source) => ({ value: source.id, label: `${source.name} — ${source.id}` }))}
            disabled={available.length === 0}
            onChange={setToAdd}
          />
          <IconButton icon={Plus} label="Add input source" disabled={!toAdd} onClick={addSource} />
        </Stack>
      </Stack>
    </Card>
  )
}
