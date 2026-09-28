import { describe, expect, it } from 'vitest'

import { toMessage } from './errors'

describe('toMessage', () => {
  it('passes a command rejection through', () => {
    expect(toMessage('maximum_delay must be positive')).toBe('maximum_delay must be positive')
  })

  it('keeps an Error message', () => {
    expect(toMessage(new Error('socket not found'))).toBe('socket not found')
  })

  it('stringifies other values', () => {
    expect(toMessage(42)).toBe('42')
  })
})
