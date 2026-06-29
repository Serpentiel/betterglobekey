import { describe, expect, it } from 'vitest'

import { toMessage } from './errors'

describe('toMessage', () => {
  it('strips the gRPC status prefix', () => {
    expect(toMessage(new Error('3 INVALID_ARGUMENT: maximum_delay must be positive'))).toBe(
      'maximum_delay must be positive',
    )
  })

  it('keeps a plain error message', () => {
    expect(toMessage(new Error('socket not found'))).toBe('socket not found')
  })

  it('handles a multi-line gRPC message', () => {
    expect(toMessage(new Error('2 UNKNOWN: line one\nline two'))).toBe('line one\nline two')
  })

  it('stringifies non-Error values', () => {
    expect(toMessage('boom')).toBe('boom')
    expect(toMessage(42)).toBe('42')
  })
})
