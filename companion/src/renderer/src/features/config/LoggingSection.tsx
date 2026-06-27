import { ScrollText } from 'lucide-react'

import type { Config } from '../../../../shared/types'
import type { ConfigErrors } from '../../lib/config-schema'
import { Field, Input, Section, Stack } from '../../ui'

interface Props {
  config: Config
  errors: ConfigErrors
  onChange: (config: Config) => void
}

export function LoggingSection({ config, errors, onChange }: Props): JSX.Element {
  const setLogger = (patch: Partial<Config['logger']>): void =>
    onChange({ ...config, logger: { ...config.logger, ...patch } })

  return (
    <Section icon={ScrollText} title="Logging" description="Where logs are written and how long they are kept.">
      <Field label="Log file" error={errors.logger.path}>
        {({ id, invalid }) => (
          <Input
            id={id}
            invalid={invalid}
            value={config.logger.path}
            placeholder="~/Library/Logs/betterglobekey.log"
            onChange={(event) => setLogger({ path: event.target.value })}
          />
        )}
      </Field>
      <Stack direction="row" gap={16}>
        <Field label="Retention (days)" error={errors.logger.retentionDays}>
          {({ id, invalid }) => (
            <Input
              id={id}
              invalid={invalid}
              type="number"
              min={0}
              value={config.logger.retentionDays}
              onChange={(event) => setLogger({ retentionDays: Number(event.target.value) })}
            />
          )}
        </Field>
        <Field label="Retention (files)" error={errors.logger.retentionFiles}>
          {({ id, invalid }) => (
            <Input
              id={id}
              invalid={invalid}
              type="number"
              min={0}
              value={config.logger.retentionFiles}
              onChange={(event) => setLogger({ retentionFiles: Number(event.target.value) })}
            />
          )}
        </Field>
      </Stack>
    </Section>
  )
}
