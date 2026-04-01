'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { healthApi } from '@/lib/api';

export function useHealthIssues() {
  return useQuery({
    queryKey: ['health'],
    queryFn: healthApi.getIssues,
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
}

export function useFixHealthIssues() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (checks: string[]) => healthApi.fix(checks),
    onSuccess: (data) => {
      queryClient.setQueryData(['health'], data);
    },
  });
}
