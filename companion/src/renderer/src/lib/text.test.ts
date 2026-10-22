import { describe, expect, it } from 'vitest'

import { capitalize, formatVersion, versionMismatch } from './text'

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

describe('versionMismatch', () => {
  it('flags different releases', () => {
    expect(versionMismatch('4.1.0', { version: '4.0.1', commit: 'abc123' })).toBe(true)
  })

  it('accepts the same release, with or without a v prefix', () => {
    expect(versionMismatch('4.1.0', { version: 'v4.1.0', commit: 'abc123' })).toBe(false)
  })

  it('ignores unreleased builds on either side', () => {
    expect(versionMismatch('0.0.0', { version: '4.1.0', commit: 'abc123' })).toBe(false)
    expect(versionMismatch('4.1.0', { version: 'dev', commit: 'abc123' })).toBe(false)
  })

  it('ignores versions that are not loaded yet', () => {
    expect(versionMismatch(null, { version: '4.1.0', commit: 'abc123' })).toBe(false)
    expect(versionMismatch('4.1.0', null)).toBe(false)
  })
})
