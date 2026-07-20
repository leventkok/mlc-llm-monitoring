import {
  LoginCredentials,
  RegisterCredentials,
  AuthResponse,
  User,
  Review,
  Decision,
  Score,
  Metrics,
} from "../types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null;

  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });
  } catch {
    throw new Error(
      "Could not reach the API. Check NEXT_PUBLIC_API_URL and that the backend is running.",
    );
  }

  let data: { error?: string };
  try {
    data = await res.json();
  } catch {
    throw new Error("Invalid response from server");
  }

  if (!res.ok) {
    throw new Error(data.error || "Something went wrong");
  }
  return data as T;
}

export const authApi = {
  register: (creds: RegisterCredentials) =>
    request<User>("/auth/register", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  login: (creds: LoginCredentials) =>
    request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  me: (token: string) =>
    request<User>("/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    }),
};

export const reviewApi = {
  list: () => request<Review[]>("/reviews"),

  create: (data: {
    app_name: string;
    store: string;
    rating: number;
    text: string;
  }) =>
    request<Review>("/reviews", { method: "POST", body: JSON.stringify(data) }),

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
