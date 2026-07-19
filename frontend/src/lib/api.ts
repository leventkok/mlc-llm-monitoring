import { AuthCredentials, AuthResponse, User } from "../types";

const API_URL = "http://localhost:8080";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || "Bir hata olustu");
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
