'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { diaryApi } from '@/lib/api';
import type { DiaryEntryRequest, SearchParams } from '@/types';

export function useDiaryEntries(params?: SearchParams) {
  return useQuery({
    queryKey: ['entries', params],
    queryFn: () => diaryApi.getItems(params),
  });
}

export function useDiaryEntry(date: string | null) {
  return useQuery({
    queryKey: ['entry', date],
    queryFn: () => (date ? diaryApi.getItemByDate(date) : null),
    enabled: !!date,
  });
}

export function useSaveEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (entry: DiaryEntryRequest) => diaryApi.saveItem(entry),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      queryClient.setQueryData(['entry', data.date], data);
    },
  });
}

export function useSearchEntries(searchText: string, tags: string = '', enabled: boolean = true) {
  return useQuery({
    queryKey: ['search', searchText, tags],
    queryFn: () => diaryApi.search(searchText, tags || undefined),
    enabled: enabled && (searchText.length > 0 || tags.length > 0),
  });
}

export function useTagStats() {
  return useQuery({
    queryKey: ['tag-stats'],
    queryFn: () => diaryApi.getTagStats(),
  });
}

export function useRenameTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ name, newName }: { name: string; newName: string }) =>
      diaryApi.renameTag(name, newName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tag-stats'] });
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      queryClient.invalidateQueries({ queryKey: ['search'] });
    },
  });
}

export function useDeleteTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => diaryApi.deleteTag(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tag-stats'] });
      queryClient.invalidateQueries({ queryKey: ['entries'] });
      queryClient.invalidateQueries({ queryKey: ['search'] });
    },
  });
}
