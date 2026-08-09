import {
  LoginCredentials,
  RegisterCredentials,
  User,
  Review,
  Decision,
  Score,
  Metrics,
  LLMConfig,
  AnalyzeLogEntry,
  ModelProfile,
  ModelSwitchRequest,
  DatasetReviewPage,
  BatchAnalyzeResult,
  Organization,
  OrgInvite,
  OrgMember,
  InvitePreview,
  OrganizationMembership,
  StoreApp,
  Audit,
  AuditReport,
} from "../types";

const PRODUCTION_API_URL = "https://mlc-llm-monitoring.onrender.com";

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL ??
  (process.env.VERCEL === "1" ? PRODUCTION_API_URL : "http://localhost:8080");

const AUTH_TOKEN_KEY = "inferreview_auth_token";

export function getAuthToken(): string | null {
  if (typeof sessionStorage === "undefined") return null;
  return sessionStorage.getItem(AUTH_TOKEN_KEY);
}

export function setAuthToken(token: string | null) {
  if (typeof sessionStorage === "undefined") return;
  if (token) sessionStorage.setItem(AUTH_TOKEN_KEY, token);
  else sessionStorage.removeItem(AUTH_TOKEN_KEY);
}

function authHeaders(extra: HeadersInit = {}): HeadersInit {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(extra as Record<string, string>),
  };
  const token = getAuthToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  return headers;
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...options,
      credentials: "include",
      headers: authHeaders(options.headers),
    });
  } catch {
    throw new Error(
      "Could not reach the API. Check NEXT_PUBLIC_API_URL and that the backend is running.",
    );
  }

  if (res.status === 204) {
    return undefined as T;
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

  login: async (creds: LoginCredentials) => {
    const res = await request<{ message: string; token?: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(creds),
    });
    if (res.token) setAuthToken(res.token);
    return res;
  },

  me: () => request<User>("/auth/me"),

  logout: async () => {
    try {
      return await request<{ message: string }>("/auth/logout");
    } finally {
      setAuthToken(null);
    }
  },

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

export const adminApi = {
  getLLMConfig: () => request<LLMConfig>("/admin/llm-config"),

  updateLLMConfig: (cfg: Partial<LLMConfig>) =>
    request<LLMConfig>("/admin/llm-config", {
      method: "PUT",
      body: JSON.stringify(cfg),
    }),

  analyzeLogs: (limit = 50) =>
    request<AnalyzeLogEntry[]>(`/admin/analyze-logs?limit=${limit}`),

  modelProfiles: () => request<ModelProfile[]>("/admin/model-profiles"),

  switchModel: (profileId: string) =>
    request<ModelSwitchRequest>("/admin/switch-model", {
      method: "POST",
      body: JSON.stringify({ profile_id: profileId }),
    }),

  modelSwitchStatus: () =>
    request<ModelSwitchRequest | null>("/admin/model-switch/status"),
};

async function requestText(path: string): Promise<string> {
  let res: Response;
  try {
    const headers: Record<string, string> = {};
    const token = getAuthToken();
    if (token) headers.Authorization = `Bearer ${token}`;
    res = await fetch(`${API_URL}${path}`, { credentials: "include", headers });
  } catch {
    throw new Error(
      "Could not reach the API. Check NEXT_PUBLIC_API_URL and that the backend is running.",
    );
  }
  if (!res.ok) {
    let message = "Something went wrong";
    try {
      const data = (await res.json()) as { error?: string };
      message = data.error || message;
    } catch {
      /* non-JSON error body */
    }
    throw new Error(message);
  }
  return res.text();
}

export const datasetApi = {
  list: (offset = 0, limit = 20) =>
    request<DatasetReviewPage>(
      `/dataset/reviews?offset=${offset}&limit=${limit}`,
    ),

  exportCSV: (offset = 0, limit = 100) =>
    requestText(`/dataset/reviews.csv?offset=${offset}&limit=${limit}`),

  batchAnalyze: (offset = 0, limit = 10) =>
    request<BatchAnalyzeResult>(
      `/dataset/batch-analyze?offset=${offset}&limit=${limit}`,
      { method: "POST" },
    ),
};

export const inviteApi = {
  preview: (token: string) => request<InvitePreview>(`/invites/${token}`),

  accept: (token: string) =>
    request<{ message: string; organization: OrganizationMembership }>(
      `/invites/${token}/accept`,
      { method: "POST" },
    ),
};

export const orgAdminApi = {
  listOrganizations: () => request<Organization[]>("/admin/organizations"),

  createOrganization: (name: string) =>
    request<Organization>("/admin/organizations", {
      method: "POST",
      body: JSON.stringify({ name }),
    }),

  deleteOrganization: (orgId: string) =>
    request<{ message: string }>(`/admin/organizations/${orgId}`, {
      method: "DELETE",
    }),

  listInvites: (orgId: string) =>
    request<OrgInvite[]>(`/admin/organizations/${orgId}/invites`),

  createInvite: (
    orgId: string,
    data: { role: string; email?: string; days?: number },
  ) =>
    request<OrgInvite>(`/admin/organizations/${orgId}/invites`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
};

export const orgApi = {
  listMembers: () => request<OrgMember[]>("/organization/members"),

  removeMember: (userId: string) =>
    request<{ message: string }>(`/organization/members/${userId}`, {
      method: "DELETE",
    }),

  listInvites: () => request<OrgInvite[]>("/organization/invites"),

  createInvite: (data: { role: string; email?: string; days?: number }) =>
    request<OrgInvite>("/organization/invites", {
      method: "POST",
      body: JSON.stringify(data),
    }),
};

export const auditApi = {
  search: (query: string, country?: string) =>
    request<{ play: StoreApp[]; appstore: StoreApp[] }>("/audits/search", {
      method: "POST",
      body: JSON.stringify({ query, country: country || "tr" }),
    }),

  list: () => request<Audit[]>("/audits"),

  create: (data: {
    client_name: string;
    app_display_name: string;
    play_app_id?: string;
    appstore_app_id?: string;
    country?: string;
    play_review_limit?: number;
    appstore_review_limit?: number;
    mode: "quick" | "full";
  }) =>
    request<Audit>("/audits", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  get: (id: string) => request<Audit>(`/audits/${id}`),

  report: (id: string) => request<AuditReport>(`/audits/${id}/report`),

  remove: (id: string) =>
    request<void>(`/audits/${id}`, {
      method: "DELETE",
    }),
};
