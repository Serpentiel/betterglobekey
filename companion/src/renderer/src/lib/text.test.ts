import { describe, expect, it } from 'vitest'

import { capitalize, formatVersion } from './text'

describe('capitalize', () => {
  it('upper-cases the first character', () => {
    expect(capitalize('shift')).toBe('Shift')
  })

  it('leaves an empty string untouched', () => {
    expect(capitalize('')).toBe('')
  })

  it('leaves an already-capitalized string untouched', () => {
    expect(capitalize('Option')).toBe('Option')
  })
})

describe('formatVersion', () => {
  it('adds a leading v to a bare release version', () => {
    expect(formatVersion({ version: '4.0.0', commit: 'abc123' })).toBe('v4.0.0')
  })

  it('keeps an existing v prefix', () => {
    expect(formatVersion({ version: 'v4.0.0', commit: 'abc123' })).toBe('v4.0.0')
  })

  it('falls back to the commit for a dev build', () => {
    expect(formatVersion({ version: 'dev', commit: 'abc123' })).toBe('abc123')
  })

  it('falls back to dev when neither version nor commit is set', () => {
    expect(formatVersion({ version: '', commit: '' })).toBe('dev')
  })
})
