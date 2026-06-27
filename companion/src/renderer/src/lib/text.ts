/** capitalize upper-cases the first character of s. */
export function capitalize(s: string): string {
  return s.length === 0 ? s : s.charAt(0).toUpperCase() + s.slice(1)
}
