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

const release = /^v?(\d+\.\d+\.\d+)$/

/** versionMismatch reports whether the companion and the daemon are different
 * releases. They ship together, so a difference means one of them is stale.
 * Unreleased builds (companion 0.0.0, daemon "dev") never count. */
export function versionMismatch(companion: string | null, daemon: Version | null): boolean {
  const ours = release.exec(companion ?? '')?.[1]
  const theirs = release.exec(daemon?.version ?? '')?.[1]

  return ours !== undefined && ours !== '0.0.0' && theirs !== undefined && ours !== theirs
}
