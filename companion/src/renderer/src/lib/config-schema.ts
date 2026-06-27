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
    hud: z.boolean(),
    doublePressMaxDelay: z.string().refine((value) => {
      const milliseconds = parseGoDuration(value)

      return milliseconds !== null && milliseconds > 0
    }, 'Enter a positive Go duration, e.g. 250ms'),
    logger: z.object({
      path: z.string().trim().min(1, 'A log file path is required'),
      retentionDays: z.number().int('Must be a whole number').min(0, 'Cannot be negative'),
      retentionFiles: z.number().int('Must be a whole number').min(0, 'Cannot be negative'),
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
  doublePressMaxDelay?: string
  logger: { path?: string; retentionDays?: string; retentionFiles?: string }
  collections: Record<number, { name?: string; sources?: string }>
}

export interface ValidationResult {
  valid: boolean
  errors: ConfigErrors
}

/** validateConfig runs the schema and maps issues to a per-field error tree. */
export function validateConfig(config: Config): ValidationResult {
  const errors: ConfigErrors = { logger: {}, collections: {} }
  const result = configSchema.safeParse(config)

  if (result.success) {
    return { valid: true, errors }
  }

  for (const issue of result.error.issues) {
    const [head, second, third] = issue.path

    if (head === 'doublePressMaxDelay') {
      errors.doublePressMaxDelay ??= issue.message
    } else if (head === 'logger' && typeof second === 'string') {
      const key = second as keyof ConfigErrors['logger']
      errors.logger[key] ??= issue.message
    } else if (head === 'collections' && typeof second === 'number') {
      const field = third as 'name' | 'sources'
      errors.collections[second] ??= {}
      errors.collections[second][field] ??= issue.message
    }
  }

  return { valid: false, errors }
}
