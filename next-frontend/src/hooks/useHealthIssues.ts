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
    onSuccess: (data) => { queryClient.setQueryData(['health'], data); },
  });
}

export function useDeleteOrphan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (filename: string) => healthApi.deleteOrphan(filename),
    onSuccess: (data) => { queryClient.setQueryData(['health'], data); },
  });
}

export function useAttachOrphan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ filename, date }: { filename: string; date: string }) =>
      healthApi.attachOrphan(filename, date),
    onSuccess: (data) => { queryClient.setQueryData(['health'], data); },
  });
}

export function useIgnoreOrphan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (filename: string) => healthApi.ignoreOrphan(filename),
    onSuccess: (data) => { queryClient.setQueryData(['health'], data); },
  });
}

export function useUnignoreOrphan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (filename: string) => healthApi.unignoreOrphan(filename),
    onSuccess: (data) => { queryClient.setQueryData(['health'], data); },
  });
}
