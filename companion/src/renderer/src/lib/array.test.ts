import { describe, expect, it } from 'vitest'

import { move } from './array'

describe('move', () => {
  it('moves an element forward', () => {
    expect(move(['a', 'b', 'c'], 0, 1)).toEqual(['b', 'a', 'c'])
  })

  it('moves an element backward', () => {
    expect(move(['a', 'b', 'c'], 2, -1)).toEqual(['a', 'c', 'b'])
  })

  it('returns the list unchanged when the target is out of bounds', () => {
    expect(move(['a', 'b', 'c'], 0, -1)).toEqual(['a', 'b', 'c'])
    expect(move(['a', 'b', 'c'], 2, 1)).toEqual(['a', 'b', 'c'])
  })

  it('does not mutate the original list', () => {
    const items = ['a', 'b', 'c']
    move(items, 0, 1)
    expect(items).toEqual(['a', 'b', 'c'])
  })
})
