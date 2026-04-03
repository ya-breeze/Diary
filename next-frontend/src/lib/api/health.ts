import { apiClient } from './client';

export interface HealthIssue {
  check: string;
  path: string;
  message: string;
  fixable: boolean;
}

export interface HealthIssuesResponse {
  issues: HealthIssue[];
  lastChecked?: string;
  ignoredOrphans?: string[];
}

export const healthApi = {
  getIssues: (): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>('/v1/health/issues'),

  fix: (checks: string[]): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>('/v1/health/fix', {
      method: 'POST',
      body: { checks },
    }),

  deleteOrphan: (filename: string): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>(`/v1/health/orphans/${encodeURIComponent(filename)}`, {
      method: 'DELETE',
    }),

  attachOrphan: (filename: string, date: string): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>(`/v1/health/orphans/${encodeURIComponent(filename)}/attach`, {
      method: 'POST',
      body: { date },
    }),

  ignoreOrphan: (filename: string): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>(`/v1/health/orphans/${encodeURIComponent(filename)}/ignore`, {
      method: 'POST',
    }),

  unignoreOrphan: (filename: string): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>(`/v1/health/orphans/${encodeURIComponent(filename)}/ignore`, {
      method: 'DELETE',
    }),
};
