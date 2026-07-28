import { request } from "./api";
import { Decision } from "../types";

/** MCP payload shape (FINAL BOSS gist — WebMCP → Go backend). */
export type MCPPayload = {
  action: "analyze_review";
  review_id: string;
  query?: string;
  spec?: Record<string, unknown>;
};

/** Rich MCP result from HandleMCPRequest. */
export type MCPRichResult = {
  review_id?: string;
  decision_id?: string;
  category: string;
  sentiment: string;
  raw_output: string;
  latency_ms: number;
};

function toDecision(result: MCPRichResult): Decision {
  return {
    id: result.decision_id ?? "",
    review_id: result.review_id ?? "",
    category: result.category,
    sentiment: result.sentiment,
    raw_output: result.raw_output,
    latency_ms: result.latency_ms,
    created_at: new Date().toISOString(),
  };
}

/** WebMCP client — packages analyze requests as MCP payloads. */
export async function mcpAnalyzeReview(reviewId: string): Promise<Decision> {
  const payload: MCPPayload = {
    action: "analyze_review",
    review_id: reviewId,
  };
  const result = await request<MCPRichResult>("/mcp", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  return toDecision(result);
}
