import { describe, expect, it } from 'vitest'

import {
  allowedModes,
  GLOBE,
  triggerFor,
  withKey,
  TRIGGER_KEYS,
  triggerKey,
  triggerLabel,
  triggerName,
} from './trigger-keys'

describe('trigger keys', () => {
  it('mirror the daemon list, with Globe first', () => {
    expect(TRIGGER_KEYS.map((key) => key.value)).toEqual([
      'globe',
      'caps_lock',
      'right_command',
      'right_option',
      'right_control',
      'right_shift',
      'left_command',
      'left_option',
      'left_control',
      'left_shift',
      'f13',
      'f14',
      'f15',
      'f16',
      'f17',
      'f18',
      'f19',
      'f20',
      'menu',
      'insert',
      'print_screen',
      'scroll_lock',
      'pause',
    ])
  })

  it('only lets keys other than the Globe key and Caps Lock pass through', () => {
    const remapOnly = TRIGGER_KEYS.filter((key) => !key.passThrough).map((key) => key.value)
    expect(remapOnly).toEqual([GLOBE, 'caps_lock'])
  })

  it('lists the modes each key allows', () => {
    expect(allowedModes('right_command')).toEqual(['fn', 'dedicated', 'pass_through'])
    expect(allowedModes('caps_lock')).toEqual(['fn', 'dedicated'])
    expect(allowedModes('f20')).toEqual(['fn', 'dedicated', 'pass_through'])
  })

  it('resolves keyboard triggers like the daemon', () => {
    const insert = { key: 'insert', mode: 'pass_through' }
    const trigger = { keyboards: { '046d:c33f': insert } }

    expect(triggerFor(trigger, '046d:c33f')).toEqual([insert, true])
    expect(triggerFor(trigger, '1234:5678')).toEqual([{ key: 'globe', mode: 'fn' }, false])
  })

  it('keeps a mode the new key allows', () => {
    expect(withKey({ key: 'insert', mode: 'pass_through' }, 'f13')).toEqual({ key: 'f13', mode: 'pass_through' })
    expect(withKey({ key: 'insert', mode: 'pass_through' }, 'caps_lock')).toEqual({ key: 'caps_lock', mode: 'fn' })
  })

  it('names keys, falling back to the config name', () => {
    expect(triggerLabel('right_command')).toBe('Right ⌘ Command')
    expect(triggerName('right_command')).toBe('Right Command')
    expect(triggerName('globe')).toBe('Globe')
    expect(triggerLabel('mystery')).toBe('mystery')
    expect(triggerName('mystery')).toBe('mystery')
    expect(triggerKey('mystery')).toBeUndefined()
  })
})
