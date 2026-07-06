import { ApiError } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api';

interface FetchOptions extends Omit<RequestInit, 'body'> {
  params?: Record<string, string | undefined>;
  body?: unknown;
}

let isRefreshing = false;
let refreshSubscribers: Array<(success: boolean) => void> = [];

function subscribeTokenRefresh(cb: (success: boolean) => void) {
  refreshSubscribers.push(cb);
}

function onRefreshed(success: boolean) {
  refreshSubscribers.forEach(cb => cb(success));
  refreshSubscribers = [];
}

export async function apiClient<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T> {
  const { params, body, headers: customHeaders, ...fetchOptions } = options;

  // Build URL with query params
  let url = `${API_BASE_URL}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        searchParams.append(key, value);
      }
    });
    const queryString = searchParams.toString();
    if (queryString) {
      url += `?${queryString}`;
    }
  }

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...customHeaders,
  };

  const response = await fetch(url, {
    ...fetchOptions,
    headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    if (response.status === 401 && !endpoint.includes('/authorize') && !endpoint.includes('/auth/refresh')) {
      if (!isRefreshing) {
        isRefreshing = true;
        fetch(`${API_BASE_URL}/auth/refresh`, { method: 'POST', credentials: 'include' })
          .then(r => { isRefreshing = false; onRefreshed(r.ok); })
          .catch(() => { isRefreshing = false; onRefreshed(false); });
      }

      return new Promise<T>((resolve, reject) => {
        subscribeTokenRefresh(async success => {
          if (success) {
            try {
              const retryResponse = await fetch(url, {
                ...fetchOptions,
                headers,
                credentials: 'include',
                body: body ? JSON.stringify(body) : undefined,
              });
              if (retryResponse.ok) {
                const retryContentType = retryResponse.headers.get('content-type');
                if (retryContentType?.includes('application/json')) {
                  resolve(retryResponse.json());
                } else {
                  resolve(retryResponse.text() as unknown as T);
                }
              } else if (retryResponse.status === 401) {
                // Refresh succeeded but the session is still rejected — the
                // session is truly dead. Redirect to login (handled by the
                // redirect, so callers should not surface this as an error).
                if (typeof window !== 'undefined') {
                  window.location.href = '/login';
                }
                reject(new ApiError(401, 'Session expired'));
              } else {
                reject(new ApiError(retryResponse.status, `HTTP ${retryResponse.status}`));
              }
            } catch (err) {
              reject(err);
            }
          } else {
            if (typeof window !== 'undefined') {
              window.location.href = '/login';
            }
            reject(new ApiError(401, 'Session expired'));
          }
        });
      });
    }
    const errorText = await response.text();
    throw new ApiError(response.status, errorText || `HTTP ${response.status}`);
  }

  // Handle empty responses
  const contentType = response.headers.get('content-type');
  if (contentType?.includes('application/json')) {
    return response.json();
  }

  return response.text() as unknown as T;
}

export { API_BASE_URL };
