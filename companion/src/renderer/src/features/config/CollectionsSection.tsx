import type { ReactElement } from 'react'
import { Layers, Plus } from 'lucide-react'

import type { Collection, Config, InputSource } from '../../../../shared/types'
import { move } from '../../lib/array'
import type { ConfigErrors } from '../../lib/config-schema'
import { Button, Section, Stack } from '../../ui'
import { CollectionCard } from './CollectionCard'

interface Props {
  config: Config
  sources: InputSource[]
  errors: ConfigErrors
  nameOf: (id: string) => string
  onChange: (config: Config) => void
}

export function CollectionsSection({ config, sources, errors, nameOf, onChange }: Props): ReactElement {
  const setCollections = (collections: Collection[]): void => onChange({ ...config, collections })

  const replace = (index: number, collection: Collection): void =>
    setCollections(config.collections.map((existing, position) => (position === index ? collection : existing)))

  const addCollection = (): void =>
    setCollections([...config.collections, { name: `collection ${config.collections.length + 1}`, sources: [] }])

  return (
    <Section
      icon={Layers}
      title="Collections"
      description="Ordered groups of input sources the Globe key cycles through."
      actions={
        <Button icon={Plus} onClick={addCollection}>
          Add
        </Button>
      }
    >
      <Stack gap={12}>
        {config.collections.map((collection, index) => (
          <CollectionCard
            key={index}
            collection={collection}
            sources={sources}
            index={index}
            total={config.collections.length}
            nameError={errors.collections[index]?.name}
            sourcesError={errors.collections[index]?.sources}
            nameOf={nameOf}
            onChange={(updated) => replace(index, updated)}
            onMove={(delta) => setCollections(move(config.collections, index, delta))}
            onRemove={() => setCollections(config.collections.filter((_, position) => position !== index))}
          />
        ))}
      </Stack>
    </Section>
  )
}
