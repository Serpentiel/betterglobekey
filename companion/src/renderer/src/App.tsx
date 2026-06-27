import { useCallback, useEffect, useMemo, useState } from 'react'

import type { Config, InputSource } from '../../shared/types'
import { CollectionsEditor } from './components/CollectionsEditor'

type Status = { kind: 'idle' } | { kind: 'saving' } | { kind: 'saved' } | { kind: 'error'; message: string }

export function App(): JSX.Element {
  const [config, setConfig] = useState<Config | null>(null)
  const [sources, setSources] = useState<InputSource[]>([])
  const [current, setCurrent] = useState<InputSource | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [status, setStatus] = useState<Status>({ kind: 'idle' })

  const load = useCallback(async () => {
    setLoadError(null)

    try {
      const [loadedConfig, loadedSources, loadedCurrent] = await Promise.all([
        window.api.getConfig(),
        window.api.listInputSources(),
        window.api.getCurrentSource(),
      ])

      setConfig(loadedConfig)
      setSources(loadedSources)
      setCurrent(loadedCurrent)
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error))
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const nameOf = useMemo(() => {
    const byId = new Map(sources.map((source) => [source.id, source.name]))

    return (id: string): string => byId.get(id) ?? id
  }, [sources])

  const save = useCallback(async () => {
    if (!config) {
      return
    }

    setStatus({ kind: 'saving' })

    try {
      await window.api.applyConfig(config)
      setStatus({ kind: 'saved' })
    } catch (error) {
      setStatus({ kind: 'error', message: error instanceof Error ? error.message : String(error) })
    }
  }, [config])

  if (loadError) {
    return (
      <div className="screen screen--center">
        <h1>Cannot reach betterglobekey</h1>
        <p className="muted">{loadError}</p>
        <p className="muted">Make sure the betterglobekey service is running, then try again.</p>
        <button type="button" onClick={() => void load()}>
          Retry
        </button>
      </div>
    )
  }

  if (!config) {
    return (
      <div className="screen screen--center">
        <p className="muted">Loading…</p>
      </div>
    )
  }

  const update = (partial: Partial<Config>): void => {
    setConfig({ ...config, ...partial })
    setStatus({ kind: 'idle' })
  }

  return (
    <div className="screen">
      <header className="header">
        <div>
          <h1>betterglobekey</h1>
          {current ? <p className="muted">Active: {current.name}</p> : null}
        </div>
        <div className="header__actions">
          <button type="button" className="secondary" onClick={() => void load()}>
            Reload
          </button>
          <button type="button" onClick={() => void save()} disabled={status.kind === 'saving'}>
            {status.kind === 'saving' ? 'Saving…' : 'Save'}
          </button>
        </div>
      </header>

      {status.kind === 'saved' ? <div className="banner banner--ok">Saved. The daemon reloaded the change.</div> : null}
      {status.kind === 'error' ? <div className="banner banner--error">{status.message}</div> : null}

      <section className="card">
        <h2>Behavior</h2>
        <label className="field field--checkbox">
          <input type="checkbox" checked={config.hud} onChange={(event) => update({ hud: event.target.checked })} />
          <span>Show the on-screen HUD when the input source changes</span>
        </label>
        <label className="field">
          <span>Double-press maximum delay</span>
          <input
            type="text"
            value={config.doublePressMaxDelay}
            onChange={(event) => update({ doublePressMaxDelay: event.target.value })}
            placeholder="250ms"
          />
          <small className="muted">A Go duration, e.g. 250ms.</small>
        </label>
      </section>

      <section className="card">
        <h2>Collections</h2>
        <CollectionsEditor
          collections={config.collections}
          sources={sources}
          nameOf={nameOf}
          onChange={(collections) => update({ collections })}
        />
      </section>

      <section className="card">
        <h2>Logging</h2>
        <label className="field">
          <span>Log file</span>
          <input
            type="text"
            value={config.logger.path}
            onChange={(event) => update({ logger: { ...config.logger, path: event.target.value } })}
          />
        </label>
        <div className="row">
          <label className="field">
            <span>Retention (days)</span>
            <input
              type="number"
              min={0}
              value={config.logger.retentionDays}
              onChange={(event) => update({ logger: { ...config.logger, retentionDays: Number(event.target.value) } })}
            />
          </label>
          <label className="field">
            <span>Retention (files)</span>
            <input
              type="number"
              min={0}
              value={config.logger.retentionFiles}
              onChange={(event) => update({ logger: { ...config.logger, retentionFiles: Number(event.target.value) } })}
            />
          </label>
        </div>
      </section>
    </div>
  )
}
