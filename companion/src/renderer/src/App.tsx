import { type ReactElement, useEffect } from 'react'

import { ConfigEditor } from './features/config/ConfigEditor'
import { ConnectionError } from './features/status/ConnectionError'
import { useConfig } from './hooks/useConfig'
import { Spinner, Stack, Titlebar } from './ui'
import styles from './App.module.css'

export function App(): ReactElement {
  const controller = useConfig()
  const { save } = controller

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent): void => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
        event.preventDefault()
        void save()
      }
    }

    window.addEventListener('keydown', onKeyDown)

    return () => window.removeEventListener('keydown', onKeyDown)
  }, [save])

  return (
    <div className={styles.app}>
      {controller.phase.status === 'ready' ? (
        <ConfigEditor controller={controller} />
      ) : (
        <>
          <Titlebar title="betterglobekey" />
          {controller.phase.status === 'loading' ? (
            <Stack flex={1} align="center" justify="center">
              <Spinner size={26} />
            </Stack>
          ) : (
            <ConnectionError message={controller.phase.message} onRetry={() => void controller.reload()} />
          )}
        </>
      )}
    </div>
  )
}
