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

export function useSearchEntries(searchText: string, enabled: boolean = true) {
  return useQuery({
    queryKey: ['search', searchText],
    queryFn: () => diaryApi.search(searchText),
    enabled: enabled && searchText.length > 0,
  });
}
