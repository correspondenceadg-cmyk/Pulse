let accessToken: string | null = null;
let refreshPromise: Promise<string> | null = null;
let onAuthLost: (() => void) | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken() {
  return accessToken;
}

export function setAuthLostHandler(fn: () => void) {
  onAuthLost = fn;
}

export class ApiError extends Error {
  status: number;
  detail: string;
  constructor(status: number, detail: string) {
    super(detail);
    this.status = status;
    this.detail = detail;
  }
}

async function refreshAccessToken(): Promise<string> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    try {
      const res = await fetch('/api/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      });
      if (!res.ok) throw new Error('refresh failed');
      const data = (await res.json()) as { accessToken: string };
      accessToken = data.accessToken;
      return data.accessToken;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}

interface RequestOptions {
  method?: string;
  body?: unknown;
  signal?: AbortSignal;
}

async function rawRequest(path: string, opts: RequestOptions): Promise<Response> {
  const headers: Record<string, string> = {};
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json';
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

  return fetch(path, {
    method: opts.method ?? 'GET',
    headers,
    credentials: 'include',
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    signal: opts.signal,
  });
}

export async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  let res = await rawRequest(path, opts);

  const isAuthEndpoint = path.startsWith('/api/auth/');

  if (res.status === 401 && !isAuthEndpoint) {
    try {
      await refreshAccessToken();
      res = await rawRequest(path, opts);
    } catch {
      accessToken = null;
      onAuthLost?.();
      throw new ApiError(401, 'session expired');
    }
  }

  if (!res.ok) {
    let detail = res.statusText;
    try {
      const body = await res.json();
      if (body && typeof body.detail === 'string') detail = body.detail;
      else if (body && typeof body.title === 'string') detail = body.title;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, detail);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}