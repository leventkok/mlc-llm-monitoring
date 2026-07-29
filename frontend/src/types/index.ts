export interface User {
  id: string;
  email: string;
  username: string;
  is_admin?: boolean;
}

export interface RegisterCredentials {
  email: string;
  username: string;
  password: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
}

export interface Review {
  id: string;
  app_name: string;
  store: string;
  rating: number;
  text: string;
  created_at: string;
}

export interface Decision {
  id: string;
  review_id: string;
  category: string;
  sentiment: string;
  raw_output: string;
  latency_ms: number;
  created_at: string;
}

export interface Score {
  id: string;
  decision_id: string;
  quality: number;
  correct_category?: string;
  scored_by: string;
  created_at: string;
}

export interface Metrics {
  total_reviews: number;
  total_decisions: number;
  total_scores: number;
  category_counts: Record<string, number>;
  sentiment_counts: Record<string, number>;
  avg_quality: number;
  avg_latency_ms: number;
  accuracy_pct: number;
}

export interface LLMConfig {
  system_prompt: string;
  temperature: number;
  max_tokens: number;
  top_p: number;
  active_model: string;
  active_adapter: string;
}

export interface AnalyzeLogEntry {
  time: string;
  user_id: string;
  review_id: string;
  category?: string;
  sentiment?: string;
  latency_ms?: number;
  status: string;
  error?: string;
}

export interface ModelProfile {
  id: string;
  label: string;
  request_model: string;
  engine_model: string;
  local_adapter: string;
  description?: string;
}

export interface ModelSwitchRequest {
  id: string;
  profile_id: string;
  request_model: string;
  engine_model: string;
  local_adapter: string;
  status: "pending" | "running" | "completed" | "failed";
  error_message?: string;
  requested_by?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface DatasetReviewRow {
  review_id: string;
  store: string;
  app_name: string;
  app_version: string;
  rating: number;
  text: string;
  language: string;
  category?: string;
  sentiment?: string;
}

export interface DatasetReviewPage {
  dataset_id: string;
  offset: number;
  limit: number;
  rows: DatasetReviewRow[];
}

export interface BatchAnalyzeItem {
  review_id: string;
  text: string;
  expected_category?: string;
  expected_sentiment?: string;
  category: string;
  sentiment: string;
  match: boolean;
  latency_ms: number;
}

export interface BatchAnalyzeResult {
  dataset_id: string;
  processed: number;
  accuracy_pct: number;
  items: BatchAnalyzeItem[];
}
