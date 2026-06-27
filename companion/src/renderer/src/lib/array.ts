/** move returns a copy of items with the element at index shifted by delta. */
export function move<T>(items: T[], index: number, delta: number): T[] {
  const target = index + delta

  if (target < 0 || target >= items.length) {
    return items
  }

  const next = [...items]
  ;[next[index], next[target]] = [next[target], next[index]]

  return next
}
