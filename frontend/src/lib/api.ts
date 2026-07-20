import { AuthCredentials, AuthResponse, User } from "../types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || "Something went wrong");
  }
  return data as T;
}

export const authApi = {
  register: (creds: AuthCredentials) =>
    request<User>("/auth/register", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  login: (creds: AuthCredentials) =>
    request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  me: (token: string) =>
    request<User>("/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    }),
};
import { Review, Decision, Score, Metrics } from "@/types";

export const reviewApi = {
  list: () => request<Review[]>("/reviews"),

  create: (data: {
    app_name: string;
    store: string;
    rating: number;
    text: string;
  }) =>
    request<Review>("/reviews", { method: "POST", body: JSON.stringify(data) }),

  analyze: (reviewId: string) =>
    request<Decision>("/analyze", {
      method: "POST",
      body: JSON.stringify({ review_id: reviewId }),
    }),

  saveDecision: (data: {
    review_id: string;
    category: string;
    sentiment: string;
    raw_output: string;
    latency_ms: number;
  }) =>
    request<Decision>("/decisions", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  decisions: () => request<Decision[]>("/decisions"),

  scores: () => request<Score[]>("/scores"),

  score: (data: {
    decision_id: string;
    quality: number;
    correct_category?: string;
  }) =>
    request<Score>("/scores", { method: "POST", body: JSON.stringify(data) }),

  metrics: () => request<Metrics>("/metrics"),
};
