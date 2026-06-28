import { RotateCcw, Save } from 'lucide-react'
import { type ReactElement, useMemo } from 'react'

import type { ConfigController } from '../../hooks/useConfig'
import { formatVersion } from '../../lib/text'
import { Banner, Button, Footer, Stack, Titlebar } from '../../ui'
import { BehaviorSection } from './BehaviorSection'
import { CollectionsSection } from './CollectionsSection'
import { HudSection } from './HudSection'
import { LoggingSection } from './LoggingSection'
import styles from './ConfigEditor.module.css'

interface Props {
  controller: ConfigController
}

export function ConfigEditor({ controller }: Props): ReactElement | null {
  const { config, sources, version, validation, dirty, saving, saved, saveError, update, save, revert } = controller

  const nameOf = useMemo(() => {
    const names = new Map(sources.map((source) => [source.id, source.name]))

    return (id: string): string => names.get(id) ?? id
  }, [sources])

  if (!config) {
    return null
  }

  const canSave = dirty && validation.valid && !saving

  return (
    <>
      <Titlebar
        title="betterglobekey"
        actions={
          <>
            {dirty ? (
              <Button variant="ghost" icon={RotateCcw} onClick={revert}>
                Revert
              </Button>
            ) : null}
            <Button variant="primary" icon={Save} loading={saving} disabled={!canSave} onClick={() => void save()}>
              Save
            </Button>
          </>
        }
      />
      <div className={styles.scroll}>
        <Stack className={styles.content} gap={16}>
          {saved ? <Banner tone="success">Saved. The daemon reloaded the change.</Banner> : null}
          {saveError ? <Banner tone="error">{saveError}</Banner> : null}
          {dirty && !validation.valid ? <Banner tone="info">Fix the highlighted fields to save.</Banner> : null}

          <BehaviorSection config={config} errors={validation.errors} onChange={update} />
          <HudSection config={config} errors={validation.errors} onChange={update} />
          <CollectionsSection
            config={config}
            sources={sources}
            errors={validation.errors}
            nameOf={nameOf}
            onChange={update}
          />
          <LoggingSection config={config} errors={validation.errors} onChange={update} />
        </Stack>
      </div>
      <Footer>{version ? formatVersion(version) : '…'}</Footer>
    </>
  )
}
