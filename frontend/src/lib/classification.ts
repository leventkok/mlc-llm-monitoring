export const CATEGORIES = ["bug", "feature", "praise", "spam", "other"] as const;
export const SENTIMENTS = ["positive", "negative", "neutral"] as const;

export type Category = (typeof CATEGORIES)[number];
export type Sentiment = (typeof SENTIMENTS)[number];

export interface ParsedClassification {
  category: Category | null;
  sentiment: Sentiment | null;
  valid: boolean;
  jsonMatch: boolean;
  consistent: boolean;
}

function isCategory(value: string): value is Category {
  return (CATEGORIES as readonly string[]).includes(value);
}

function isSentiment(value: string): value is Sentiment {
  return (SENTIMENTS as readonly string[]).includes(value);
}

export function parseClassificationJSON(
  raw: string,
): { category: Category; sentiment: Sentiment } | null {
  const idx = raw.indexOf("{");
  if (idx < 0) return null;
  const end = raw.lastIndexOf("}");
  if (end <= idx) return null;

  try {
    const obj = JSON.parse(raw.slice(idx, end + 1)) as {
      category?: string;
      sentiment?: string;
    };
    if (!obj.category || !obj.sentiment) return null;
    if (!isCategory(obj.category) || !isSentiment(obj.sentiment)) return null;
    return { category: obj.category, sentiment: obj.sentiment };
  } catch {
    return null;
  }
}

function clampQuality(value: number): number {
  return Math.min(5, Math.max(1, value));
}

/** Mirrors backend ComputeAutoQuality — output health, not ground-truth accuracy. */
export function computeAutoQuality(
  category: string,
  sentiment: string,
  rawOutput: string,
  latencyMs: number,
): number {
  if (!isCategory(category) || !isSentiment(sentiment)) return 1;

  const raw = rawOutput.trim();
  if (!raw) return 2;

  let quality = 3;
  const parsed = parseClassificationJSON(raw);
  if (parsed) {
    quality = 4;
    if (parsed.category === category && parsed.sentiment === sentiment) {
      quality = 5;
    }
  }

  if (latencyMs <= 0) {
    quality = clampQuality(quality - 1);
  } else if (latencyMs > 10_000) {
    quality = clampQuality(quality - 2);
  } else if (latencyMs > 5_000) {
    quality = clampQuality(quality - 1);
  }

  return quality;
}

export function analyzeClassification(
  category: string,
  sentiment: string,
  rawOutput: string,
): ParsedClassification {
  const parsed = parseClassificationJSON(rawOutput);
  const validCategory = isCategory(category);
  const validSentiment = isSentiment(sentiment);

  return {
    category: validCategory ? category : null,
    sentiment: validSentiment ? sentiment : null,
    valid: validCategory && validSentiment,
    jsonMatch: parsed !== null,
    consistent:
      parsed !== null &&
      parsed.category === category &&
      parsed.sentiment === sentiment,
  };
}

export type LatencyBand = "fast" | "ok" | "slow" | "critical";

export function getLatencyBand(latencyMs: number): LatencyBand {
  if (latencyMs <= 0) return "critical";
  if (latencyMs <= 1_500) return "fast";
  if (latencyMs <= 5_000) return "ok";
  if (latencyMs <= 10_000) return "slow";
  return "critical";
}

export function latencyBandLabel(band: LatencyBand): string {
  const labels: Record<LatencyBand, string> = {
    fast: "fast",
    ok: "normal",
    slow: "slow",
    critical: "very slow",
  };
  return labels[band];
}

export function qualityLabel(quality: number): string {
  if (quality >= 5) return "excellent";
  if (quality >= 4) return "good";
  if (quality >= 3) return "acceptable";
  if (quality >= 2) return "weak";
  return "invalid";
}
