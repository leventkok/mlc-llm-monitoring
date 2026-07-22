import {
  CreateMLCEngine,
  MLCEngineInterface,
  prebuiltAppConfig,
  type AppConfig,
} from "@mlc-ai/web-llm";

const MODEL_ID = "gemma-2-2b-it-q4f16_1-MLC";
const ENGINE_PROMISE_KEY = "__mlc_llm_engine_promise__";
const ENGINE_READY_KEY = "webllm_engine_ready";

const appConfig: AppConfig = {
  ...prebuiltAppConfig,
  // IndexedDB survives reloads better than Cache API alone in some browsers.
  cacheBackend: "indexeddb",
};

type ProgressListener = (pct: number) => void;
const progressListeners = new Set<ProgressListener>();

function getStoredEnginePromise(): Promise<MLCEngineInterface> | null {
  const g = globalThis as Record<string, unknown>;
  const value = g[ENGINE_PROMISE_KEY];
  if (value instanceof Promise) {
    return value as Promise<MLCEngineInterface>;
  }
  return null;
}

function setStoredEnginePromise(
  promise: Promise<MLCEngineInterface> | null,
): void {
  const g = globalThis as Record<string, unknown>;
  if (promise) {
    g[ENGINE_PROMISE_KEY] = promise;
  } else {
    delete g[ENGINE_PROMISE_KEY];
  }
}

function notifyProgress(pct: number): void {
  for (const listener of progressListeners) {
    listener(pct);
  }
  if (pct >= 1 && typeof sessionStorage !== "undefined") {
    sessionStorage.setItem(ENGINE_READY_KEY, MODEL_ID);
  }
}

export function isEngineCachedInSession(): boolean {
  if (typeof sessionStorage === "undefined") return false;
  return sessionStorage.getItem(ENGINE_READY_KEY) === MODEL_ID;
}

export function subscribeEngineProgress(
  listener: ProgressListener,
): () => void {
  progressListeners.add(listener);
  return () => progressListeners.delete(listener);
}

/**
 * Returns the shared WebLLM engine (single-flight init).
 * Model weights are cached in the browser after the first successful load.
 */
export function getEngine(
  onProgress?: ProgressListener,
): Promise<MLCEngineInterface> {
  if (onProgress) {
    progressListeners.add(onProgress);
  }

  let promise = getStoredEnginePromise();
  if (!promise) {
    promise = CreateMLCEngine(MODEL_ID, {
      appConfig,
      initProgressCallback: (report) => {
        notifyProgress(report.progress);
      },
    })
      .then((engine) => {
        notifyProgress(1);
        return engine;
      })
      .catch((err) => {
        setStoredEnginePromise(null);
        if (typeof sessionStorage !== "undefined") {
          sessionStorage.removeItem(ENGINE_READY_KEY);
        }
        throw err;
      });

    setStoredEnginePromise(promise);
  } else if (isEngineCachedInSession()) {
    queueMicrotask(() => onProgress?.(1));
  }

  return promise;
}

/** Pre-load the model once when the app opens so Analyze does not re-trigger download. */
export function warmupEngine(
  onProgress?: ProgressListener,
): Promise<MLCEngineInterface> {
  return getEngine(onProgress);
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
  onProgress?: ProgressListener,
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
