import {
  LoginCredentials,
  RegisterCredentials,
  User,
  Review,
  Decision,
  Score,
  Metrics,
} from "../types";

const PRODUCTION_API_URL = "https://mlc-llm-monitoring.onrender.com";

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ??
  (process.env.VERCEL === "1" ? PRODUCTION_API_URL : "http://localhost:8080");

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...options,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
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

const listQuery = "limit=500&offset=0";

export const authApi = {
  register: (creds: RegisterCredentials) =>
    request<User>("/auth/register", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  login: (creds: LoginCredentials) =>
    request<{ message: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(creds),
    }),

  me: () => request<User>("/auth/me"),

  logout: () => request<{ message: string }>("/auth/logout"),

  changePassword: (oldPassword: string, newPassword: string) =>
    request<{ message: string }>("/auth/change-password", {
      method: "POST",
      body: JSON.stringify({
        old_password: oldPassword,
        new_password: newPassword,
      }),
    }),

  deleteAccount: () =>
    request<{ message: string }>("/auth/me", { method: "DELETE" }),
};

export const reviewApi = {
  list: () => request<Review[]>(`/reviews?${listQuery}`),

  get: (id: string) => request<Review>(`/reviews/${id}`),

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

  decisions: () => request<Decision[]>(`/decisions?${listQuery}`),

  scores: () => request<Score[]>(`/scores?${listQuery}`),

  score: (data: {
    decision_id: string;
    quality: number;
    correct_category?: string;
  }) =>
    request<Score>("/scores", { method: "POST", body: JSON.stringify(data) }),

  metrics: () => request<Metrics>("/stats"),

  analyze: (reviewId: string) =>
    request<Decision>(`/reviews/${reviewId}/analyze`, { method: "POST" }),
};
