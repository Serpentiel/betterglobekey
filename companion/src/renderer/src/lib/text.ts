import type { Version } from '../../../shared/types'

/** capitalize upper-cases the first character of s. */
export function capitalize(s: string): string {
  return s.length === 0 ? s : s.charAt(0).toUpperCase() + s.slice(1)
}

/** formatVersion shows the release version with a leading `v` (matching the git
 * tags), falling back to the bare build commit when the daemon reports an
 * unversioned (dev) build. */
export function formatVersion({ version, commit }: Version): string {
  if (version && version !== 'dev') {
    return version.startsWith('v') ? version : `v${version}`
  }

  return commit || version || 'dev'
}
