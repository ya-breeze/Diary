import { ApiError } from '@/types';

// Friendly, status-aware messages. We deliberately do NOT surface the raw
// message carried by an ApiError (it holds the backend response body) to avoid
// leaking internal details or stack traces into the UI.
const STATUS_MESSAGES: Record<number, string> = {
  400: 'That request could not be processed. Please check your input and try again.',
  401: 'Your session has expired. Please sign in again.',
  403: 'You do not have permission to do that.',
  404: 'The requested item could not be found.',
  409: 'That change conflicts with the current state. Please refresh and try again.',
  413: 'That file is too large to upload.',
  422: 'That request could not be processed. Please check your input and try again.',
  429: 'Too many requests. Please slow down and try again.',
  500: 'Something went wrong on the server. Please try again.',
  502: 'The server is unavailable right now. Please try again shortly.',
  503: 'The server is unavailable right now. Please try again shortly.',
  504: 'The server took too long to respond. Please try again.',
};

const GENERIC = 'Something went wrong. Please try again.';

/**
 * Normalizes any thrown/rejected value into a safe, user-friendly message
 * suitable for display in a toast. Never throws.
 *
 * - `ApiError` -> status-aware friendly message (raw backend body is ignored).
 * - plain `Error` -> its message, which for our code paths (e.g. the batch
 *   upload XHR reject) is a client-generated string, not backend content.
 * - anything else -> generic fallback.
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return STATUS_MESSAGES[error.status] ?? `Request failed (${error.status}). Please try again.`;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return GENERIC;
}
