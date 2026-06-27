import { z } from 'zod'

import type { Config } from '../../../shared/types'

// GO_DURATION matches a Go time.Duration string (e.g. "250ms", "1m30s").
const GO_DURATION = /^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$/

const UNIT_MS: Record<string, number> = {
  ns: 1e-6,
  us: 1e-3,
  µs: 1e-3,
  ms: 1,
  s: 1_000,
  m: 60_000,
  h: 3_600_000,
}

export const LOG_LEVELS = ['debug', 'info', 'warn', 'error'] as const
export const REVERSE_MODIFIERS = ['shift', 'option', 'control', 'command'] as const

/** parseGoDuration returns the duration in milliseconds, or null if malformed. */
export function parseGoDuration(input: string): number | null {
  const value = input.trim()
  if (!GO_DURATION.test(value)) {
    return null
  }

  let total = 0
  const matcher = /([0-9]+(?:\.[0-9]+)?)(ns|us|µs|ms|s|m|h)/g

  for (let match = matcher.exec(value); match !== null; match = matcher.exec(value)) {
    total += Number.parseFloat(match[1]) * UNIT_MS[match[2]]
  }

  return total
}

const positiveDuration = (message: string) =>
  z.string().refine((value) => {
    const milliseconds = parseGoDuration(value)

    return milliseconds !== null && milliseconds > 0
  }, message)

const collectionSchema = z
  .object({
    name: z.string().trim().min(1, 'Name is required'),
    sources: z.array(z.string()).min(1, 'Add at least one input source'),
  })
  .superRefine((collection, ctx) => {
    const seen = new Set<string>()

    for (const source of collection.sources) {
      if (seen.has(source)) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['sources'], message: 'Each input source can appear once' })

        return
      }

      seen.add(source)
    }
  })

// configSchema mirrors the daemon's domain validation so the UI rejects exactly
// what the service would reject.
export const configSchema = z
  .object({
    logger: z.object({
      path: z.string().trim().min(1, 'A log file path is required'),
      level: z.enum(LOG_LEVELS),
      retentionDays: z.number().int('Must be a whole number').min(0, 'Cannot be negative'),
      retentionFiles: z.number().int('Must be a whole number').min(0, 'Cannot be negative'),
    }),
    doublePress: z.object({
      enabled: z.boolean(),
      maximumDelay: positiveDuration('Enter a positive Go duration, e.g. 250ms'),
    }),
    reverse: z.object({
      enabled: z.boolean(),
      modifier: z.enum(REVERSE_MODIFIERS),
    }),
    hud: z.object({
      enabled: z.boolean(),
      duration: positiveDuration('Enter a positive Go duration, e.g. 900ms'),
      showCollection: z.boolean(),
    }),
    collections: z.array(collectionSchema),
  })
  .superRefine((config, ctx) => {
    const seen = new Set<string>()

    config.collections.forEach((collection, index) => {
      const name = collection.name.trim()

      if (name && seen.has(name)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['collections', index, 'name'],
          message: 'Collection names must be unique',
        })
      }

      seen.add(name)
    })
  })

export interface ConfigErrors {
  logger: { path?: string; level?: string; retentionDays?: string; retentionFiles?: string }
  doublePress: { maximumDelay?: string }
  reverse: { modifier?: string }
  hud: { duration?: string }
  collections: Record<number, { name?: string; sources?: string }>
}

export interface ValidationResult {
  valid: boolean
  errors: ConfigErrors
}

function emptyErrors(): ConfigErrors {
  return { logger: {}, doublePress: {}, reverse: {}, hud: {}, collections: {} }
}

/** validateConfig runs the schema and maps issues to a per-field error tree. */
export function validateConfig(config: Config): ValidationResult {
  const errors = emptyErrors()
  const result = configSchema.safeParse(config)

  if (result.success) {
    return { valid: true, errors }
  }

  for (const issue of result.error.issues) {
    const [head, second, third] = issue.path

    if (head === 'logger' && typeof second === 'string') {
      const key = second as keyof ConfigErrors['logger']
      errors.logger[key] ??= issue.message
    } else if (head === 'doublePress' && second === 'maximumDelay') {
      errors.doublePress.maximumDelay ??= issue.message
    } else if (head === 'reverse' && second === 'modifier') {
      errors.reverse.modifier ??= issue.message
    } else if (head === 'hud' && second === 'duration') {
      errors.hud.duration ??= issue.message
    } else if (head === 'collections' && typeof second === 'number') {
      const field = third as 'name' | 'sources'
      errors.collections[second] ??= {}
      errors.collections[second][field] ??= issue.message
    }
  }

  return { valid: false, errors }
}
