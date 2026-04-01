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
}

export const healthApi = {
  getIssues: (): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>('/v1/health/issues'),

  fix: (checks: string[]): Promise<HealthIssuesResponse> =>
    apiClient<HealthIssuesResponse>('/v1/health/fix', {
      method: 'POST',
      body: { checks },
    }),
};
