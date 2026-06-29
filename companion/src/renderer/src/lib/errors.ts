/** toMessage extracts a readable message from an unknown thrown value, stripping
 * the gRPC status prefix that the daemon's errors carry. */
export function toMessage(error: unknown): string {
  const raw = error instanceof Error ? error.message : String(error)

  // gRPC errors look like "N UNKNOWN: <message>"; keep only the message.
  const match = /^\d+\s+[A-Z_]+:\s*(.*)$/s.exec(raw)

  return match ? match[1] : raw
}
