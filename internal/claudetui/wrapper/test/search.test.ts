import { expect, test } from "bun:test";
import { transcriptRowToClaudeStreamEvent } from "../src/claude-compat.ts";

test("native WebSearch results reach the gateway with their sources", () => {
  const result = {
    query: "latest go",
    results: [{ content: [{ title: "Go downloads", url: "https://go.dev/dl/" }] }],
  };
  const message = {
    role: "user",
    content: [{ type: "tool_result", tool_use_id: "native-search", content: "Go downloads" }],
  };
  const event = transcriptRowToClaudeStreamEvent({
    type: "user", uuid: "search-result", sessionId: "s1", message, toolUseResult: result,
  });
  expect(event?.tool_use_result).toEqual(result);
  expect(event?.message).toEqual(message);
  expect(event?.session_id).toBe("s1");
});
