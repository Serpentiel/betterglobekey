import { describe, expect, it } from 'vitest'

import type { Config } from '../../../shared/types'
import { parseGoDuration, validateConfig } from './config-schema'

function validConfig(): Config {
  return {
    logger: { path: '/tmp/betterglobekey.log', level: 'info', retentionDays: 7, retentionFiles: 5 },
    doublePress: { enabled: true, maximumDelay: '250ms' },
    reverse: { enabled: true, modifier: 'shift' },
    hud: { enabled: true, duration: '900ms', showCollection: true },
    collections: [{ name: 'Latin', sources: ['com.apple.keylayout.US'] }],
  }
}

describe('parseGoDuration', () => {
  it('parses simple units to milliseconds', () => {
    expect(parseGoDuration('250ms')).toBe(250)
    expect(parseGoDuration('1s')).toBe(1_000)
    expect(parseGoDuration('2m')).toBe(120_000)
    expect(parseGoDuration('1h')).toBe(3_600_000)
  })

  it('parses compound and fractional durations', () => {
    expect(parseGoDuration('1m30s')).toBe(90_000)
    expect(parseGoDuration('1.5s')).toBe(1_500)
  })

  it('parses microseconds spelled either way', () => {
    expect(parseGoDuration('1us')).toBe(0.001)
    expect(parseGoDuration('1µs')).toBe(0.001)
  })

  it('trims surrounding whitespace', () => {
    expect(parseGoDuration('  250ms  ')).toBe(250)
  })

  it('returns null for malformed input', () => {
    expect(parseGoDuration('')).toBeNull()
    expect(parseGoDuration('abc')).toBeNull()
    expect(parseGoDuration('250')).toBeNull()
    expect(parseGoDuration('250 ms')).toBeNull()
    expect(parseGoDuration('ms')).toBeNull()
  })
})

describe('validateConfig', () => {
  it('accepts a valid config', () => {
    expect(validateConfig(validConfig()).valid).toBe(true)
  })

  it('requires a non-blank log path', () => {
    const config = validConfig()
    config.logger.path = '   '
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.logger.path).toBeDefined()
  })

  it('rejects an unknown log level', () => {
    const config = validConfig()
    config.logger.level = 'trace'
    expect(validateConfig(config).valid).toBe(false)
  })

  it('rejects negative retention', () => {
    const config = validConfig()
    config.logger.retentionDays = -1
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.logger.retentionDays).toBeDefined()
  })

  it('rejects non-integer retention', () => {
    const config = validConfig()
    config.logger.retentionFiles = 1.5
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.logger.retentionFiles).toBeDefined()
  })

  it('rejects a non-positive double-press delay', () => {
    const config = validConfig()
    config.doublePress.maximumDelay = '0ms'
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.doublePress.maximumDelay).toBeDefined()
  })

  it('rejects a malformed HUD duration', () => {
    const config = validConfig()
    config.hud.duration = 'soon'
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.hud.duration).toBeDefined()
  })

  it('rejects an unknown reverse modifier', () => {
    const config = validConfig()
    config.reverse.modifier = 'fn'
    expect(validateConfig(config).valid).toBe(false)
  })

  it('rejects an empty collection name and empty sources', () => {
    const config = validConfig()
    config.collections = [{ name: '', sources: [] }]
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.collections[0]?.name).toBeDefined()
    expect(result.errors.collections[0]?.sources).toBeDefined()
  })

  it('rejects duplicate sources within a collection', () => {
    const config = validConfig()
    config.collections = [{ name: 'Dupes', sources: ['a', 'a'] }]
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.collections[0]?.sources).toBeDefined()
  })

  it('rejects duplicate collection names', () => {
    const config = validConfig()
    config.collections = [
      { name: 'Same', sources: ['a'] },
      { name: 'Same', sources: ['b'] },
    ]
    const result = validateConfig(config)
    expect(result.valid).toBe(false)
    expect(result.errors.collections[1]?.name).toBeDefined()
  })

  it('allows an empty collections list', () => {
    const config = validConfig()
    config.collections = []
    expect(validateConfig(config).valid).toBe(true)
  })
})
