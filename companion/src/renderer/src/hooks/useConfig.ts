import { useCallback, useEffect, useMemo, useState } from 'react'

import type { Config, InputSource, Version } from '../../../shared/types'
import { toMessage } from '../lib/errors'
import { validateConfig, type ValidationResult } from '../lib/config-schema'

type Phase = { status: 'loading' } | { status: 'ready' } | { status: 'error'; message: string }

export interface ConfigController {
  phase: Phase
  config: Config | null
  sources: InputSource[]
  current: InputSource | null
  version: Version | null
  validation: ValidationResult
  dirty: boolean
  saving: boolean
  saved: boolean
  saveError: string | null
  update: (config: Config) => void
  save: () => Promise<void>
  revert: () => void
  reload: () => Promise<void>
}

const emptyValidation: ValidationResult = {
  valid: false,
  errors: { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {} },
}

/** useConfig owns loading, editing, validating, and saving the configuration. */
export function useConfig(): ConfigController {
  const [phase, setPhase] = useState<Phase>({ status: 'loading' })
  const [config, setConfig] = useState<Config | null>(null)
  const [original, setOriginal] = useState<Config | null>(null)
  const [sources, setSources] = useState<InputSource[]>([])
  const [current, setCurrent] = useState<InputSource | null>(null)
  const [version, setVersion] = useState<Version | null>(null)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  const reload = useCallback(async () => {
    setPhase({ status: 'loading' })

    try {
      const [loadedConfig, loadedSources, loadedCurrent, loadedVersion] = await Promise.all([
        window.api.getConfig(),
        window.api.listInputSources(),
        window.api.getCurrentSource(),
        window.api.getVersion(),
      ])

      setConfig(loadedConfig)
      setOriginal(loadedConfig)
      setSources(loadedSources)
      setCurrent(loadedCurrent)
      setVersion(loadedVersion)
      setSaved(false)
      setSaveError(null)
      setPhase({ status: 'ready' })
    } catch (error) {
      setPhase({ status: 'error', message: toMessage(error) })
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const update = useCallback((next: Config) => {
    setConfig(next)
    setSaved(false)
    setSaveError(null)
  }, [])

  const revert = useCallback(() => {
    setConfig(original)
    setSaved(false)
    setSaveError(null)
  }, [original])

  const validation = useMemo(() => (config ? validateConfig(config) : emptyValidation), [config])
  const dirty = useMemo(() => JSON.stringify(config) !== JSON.stringify(original), [config, original])

  const save = useCallback(async () => {
    if (!config || !validation.valid) {
      return
    }

    setSaving(true)
    setSaveError(null)

    try {
      await window.api.applyConfig(config)
      setOriginal(config)
      setSaved(true)
    } catch (error) {
      setSaveError(toMessage(error))
    } finally {
      setSaving(false)
    }
  }, [config, validation.valid])

  return {
    phase,
    config,
    sources,
    current,
    version,
    validation,
    dirty,
    saving,
    saved,
    saveError,
    update,
    save,
    revert,
    reload,
  }
}
