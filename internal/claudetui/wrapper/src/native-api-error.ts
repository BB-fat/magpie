/**
 * Claude Code can end a failed native turn with a <synthetic> API-error
 * assistant row and turn_duration, without firing the regular Stop hook.
 * Wait for the terminal row before treating an error as final: Claude Code
 * may recover from a transient API failure within the same turn.
 */
export type NativeApiFailure = { message: string; status?: number };

export class NativeApiErrorTurn {
  private pending: NativeApiFailure | null = null;

  observe(row: Record<string, unknown>): NativeApiFailure | null {
    if (row.type === "assistant") {
      const message = row.message && typeof row.message === "object"
        ? row.message as Record<string, unknown> : {};
      if (row.isApiErrorMessage === true && message.model === "<synthetic>") {
        const blocks = Array.isArray(message.content) ? message.content : [];
        const text = blocks
          .filter((block): block is Record<string, unknown> => Boolean(block && typeof block === "object"))
          .filter(block => block.type === "text" && typeof block.text === "string")
          .map(block => String(block.text).trim())
          .filter(Boolean)
          .join("\n");
        this.pending = text ? {
          message: text,
          ...(row.error === "invalid_request" ? { status: 400 } : {}),
        } : null;
      } else if (message.model !== "<synthetic>") {
        // A subsequent real model reply means the native retry recovered.
        this.pending = null;
      }
    }
    if (row.type === "system" && row.subtype === "turn_duration" && this.pending) {
      const failure = this.pending;
      this.pending = null;
      return failure;
    }
    return null;
  }
}
