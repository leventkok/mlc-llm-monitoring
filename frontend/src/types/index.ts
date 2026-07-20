export interface User {
  id: string;
  email: string;
  username: string;
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
