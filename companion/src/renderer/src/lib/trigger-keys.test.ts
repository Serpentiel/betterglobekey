import { describe, expect, it } from 'vitest'

import {
  allowedModes,
  GLOBE,
  includesKeyboard,
  TRIGGER_KEYS,
  triggerKey,
  triggerKeyForCode,
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
    ])
  })

  it('maps key codes to keys', () => {
    expect(triggerKeyForCode('MetaRight')?.value).toBe('right_command')
    expect(triggerKeyForCode('CapsLock')?.value).toBe('caps_lock')
    expect(triggerKeyForCode('F13')?.value).toBe('f13')
    expect(triggerKeyForCode('ContextMenu')?.value).toBe('menu')
    expect(triggerKeyForCode('KeyA')).toBeUndefined()
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

  it('scopes keyboards like the daemon', () => {
    const withoutGlobe = { applyTo: 'without_globe', overrides: { '004c:029c': true, '046d:c33f': false } }
    expect(includesKeyboard(withoutGlobe, '1234:5678', false)).toBe(true)
    expect(includesKeyboard(withoutGlobe, '05ac:0281', true)).toBe(false)
    expect(includesKeyboard(withoutGlobe, '004c:029c', true)).toBe(true)
    expect(includesKeyboard(withoutGlobe, '046d:c33f', false)).toBe(false)
    expect(includesKeyboard({ applyTo: 'all', overrides: {} }, '05ac:0281', true)).toBe(true)
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
