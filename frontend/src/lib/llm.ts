import { CreateMLCEngine, MLCEngineInterface } from "@mlc-ai/web-llm";

const MODEL_ID = "gemma-2-2b-it-q4f16_1-MLC";

let enginePromise: Promise<MLCEngineInterface> | null = null;

export function getEngine(
  onProgress?: (pct: number) => void,
): Promise<MLCEngineInterface> {
  if (!enginePromise) {
    enginePromise = CreateMLCEngine(MODEL_ID, {
      initProgressCallback: (p) => onProgress?.(p.progress),
    });
  }
  return enginePromise;
}

export interface AnalyzeResult {
  category: string;
  sentiment: string;
  rawOutput: string;
  latencyMs: number;
}

const CATEGORIES = ["bug", "feature", "praise", "spam", "other"];
const SENTIMENTS = ["positive", "negative", "neutral"];

export async function analyzeReview(
  text: string,
  onProgress?: (pct: number) => void,
): Promise<AnalyzeResult> {
  const engine = await getEngine(onProgress);
  const start = performance.now();

  const prompt = `You are a strict classifier for app store reviews.
Classify the review into exactly one category and one sentiment.
Categories: ${CATEGORIES.join(", ")}.
Sentiments: ${SENTIMENTS.join(", ")}.
Respond with ONLY a JSON object like {"category":"bug","sentiment":"negative"} and nothing else.

Review: "${text}"`;

  const res = await engine.chat.completions.create({
    messages: [{ role: "user", content: prompt }],
    temperature: 0,
    max_tokens: 60,
  });

  const raw = res.choices[0]?.message?.content?.trim() ?? "";
  const latencyMs = Math.round(performance.now() - start);

  let category = "other";
  let sentiment = "neutral";
  try {
    const match = raw.match(/\{[\s\S]*\}/);
    if (match) {
      const parsed = JSON.parse(match[0]);
      if (CATEGORIES.includes(parsed.category)) category = parsed.category;
      if (SENTIMENTS.includes(parsed.sentiment)) sentiment = parsed.sentiment;
    }
  } catch {}

  return { category, sentiment, rawOutput: raw, latencyMs };
}
