/**
 * The control a wrapper run with --desplega-serve reads on its stdin, one
 * JSON object a line, from magpie's gateway:
 *
 *   {"op":"submit","segment":"<id>","input":"<prompt>"}  one turn's prompt
 *   {"op":"shutdown"}                                    end the run
 *
 * The run answers on stdout, besides each turn's stream-json events and
 * terminal result:
 *
 *   {"type":"bridge_ready","turns":<n>}        Claude Code takes a prompt;
 *                                              n turns have completed
 *   {"type":"bridge_accepted","segment":"<id>"} the turn's prompt was typed
 *
 * A turn is submitted only after a bridge_ready: one turn runs at a time,
 * and a turn that fails ends the run.
 */
export type ServeControl =
  | { op: "submit"; segment: string; input: string }
  | { op: "shutdown" };

// A segment names a file the Stop hook reads (SERVE_TURN_FILE): letters,
// digits and dashes only, bounded.
const SEGMENT_RE = /^[A-Za-z0-9-]{1,64}$/;

export function parseServeControl(line: string): ServeControl | { error: string } {
  let value: unknown;
  try {
    value = JSON.parse(line);
  } catch {
    return { error: "control line is not JSON" };
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return { error: "control line is not an object" };
  }
  const control = value as Record<string, unknown>;
  switch (control.op) {
    case "shutdown":
      return { op: "shutdown" };
    case "submit":
      if (typeof control.segment !== "string" || !SEGMENT_RE.test(control.segment)) {
        return { error: "submit needs a segment of letters, digits and dashes" };
      }
      if (typeof control.input !== "string" || control.input === "") {
        return { error: "submit needs a non-empty input" };
      }
      return { op: "submit", segment: control.segment, input: control.input };
    default:
      return { error: `unknown control op ${JSON.stringify(control.op)}` };
  }
}

/**
 * A stop event belongs to the turn in flight when it names that turn: an
 * event left from an earlier turn, or one with no name, never completes it.
 */
export function stopEventIsTurns(event: { bridge_turn?: unknown }, segment: string): boolean {
  return typeof event.bridge_turn === "string" && event.bridge_turn === segment;
}
