import type { ReactElement } from 'react'
import { PlugZap, RotateCw } from 'lucide-react'

import { Button, EmptyState } from '../../ui'

interface Props {
  message: string
  onRetry: () => void
}

export function ConnectionError({ message, onRetry }: Props): ReactElement {
  return (
    <EmptyState
      icon={PlugZap}
      title="Can’t reach betterglobekey"
      description={`${message}. Make sure the betterglobekey service is running, then try again.`}
      action={
        <Button variant="primary" icon={RotateCw} onClick={onRetry}>
          Retry
        </Button>
      }
    />
  )
}
