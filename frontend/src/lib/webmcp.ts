import { request } from "./api";
import { Decision } from "../types";

export type MCPAnalyzePayload = {
  action: "analyze_review";
  review_id: string;
};

export type MCPDeepKwikiPayload = {
  action: "deepkwiki_search";
  query: string;
  spec: {
    repo: string;
    mode?: "ask" | "structure" | "contents";
  };
};

export type MCPAnalyzeResult = {
  action?: "analyze_review";
  review_id?: string;
  decision_id?: string;
  category: string;
  sentiment: string;
  raw_output: string;
  latency_ms: number;
};

export type MCPDeepKwikiResult = {
  action: "deepkwiki_search";
  repo_name: string;
  query?: string;
  mode: string;
  answer: string;
  latency_ms: number;
};

function toDecision(result: MCPAnalyzeResult): Decision {
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

/** WebMCP — review classification via MCP. */
export async function mcpAnalyzeReview(reviewId: string): Promise<Decision> {
  const payload: MCPAnalyzePayload = {
    action: "analyze_review",
    review_id: reviewId,
  };
  const result = await request<MCPAnalyzeResult>("/mcp", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  return toDecision(result);
}

/** WebMCP — DeepKwiki repo research (gist Faz 3). */
export async function mcpDeepKwikiSearch(
  repo: string,
  query: string,
  mode: "ask" | "structure" | "contents" = "ask",
): Promise<MCPDeepKwikiResult> {
  const payload: MCPDeepKwikiPayload = {
    action: "deepkwiki_search",
    query,
    spec: { repo, mode },
  };
  return request<MCPDeepKwikiResult>("/mcp", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
