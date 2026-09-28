/** toMessage extracts a readable message from an unknown thrown value. Tauri
 * commands reject with the daemon's message as a plain string. */
export function toMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}
