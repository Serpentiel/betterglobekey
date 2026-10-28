import { type ReactElement, useState } from 'react'
import { Globe } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import type { KeyboardsState } from '../../hooks/useKeyboards'
import type { ConfigErrors } from '../../lib/config-schema'
import { Button, Section, Setting } from '../../ui'
import { KeyboardsSheet } from './KeyboardsSheet'

interface Props {
  config: Config
  errors: ConfigErrors
  keyboards: KeyboardsState
  onChange: (config: Config) => void
}

/** keyboardsSummary describes the keyboards row: how many there are, and how many set their own trigger. */
function keyboardsSummary(config: Config, { report, error }: KeyboardsState): string {
  if (error) {
    return `Unavailable: ${error}`
  }

  if (!report) {
    return 'Checking connected keyboards…'
  }

  const total = report.keyboards.length
  const listed = Object.keys(config.trigger.keyboards).length
  const connected = total === 1 ? '1 keyboard connected' : `${total} keyboards connected`

  return listed === 0 ? `${connected}.` : `${connected}; ${listed} set their own trigger key.`
}

export function TriggerSection({ config, errors, keyboards, onChange }: Props): ReactElement {
  const [sheetOpen, setSheetOpen] = useState(false)

  return (
    <Section
      icon={Globe}
      title="Trigger Key"
      description="The key that switches input sources. Keyboards use the Globe key unless you set their own."
    >
      <Setting
        title="Keyboards"
        description={errors.trigger.keyboards ?? keyboardsSummary(config, keyboards)}
        control={<Button onClick={() => setSheetOpen(true)}>Keyboards…</Button>}
      />
      <KeyboardsSheet
        open={sheetOpen}
        config={config}
        keyboards={keyboards}
        onChange={onChange}
        onClose={() => setSheetOpen(false)}
      />
    </Section>
  )
}
