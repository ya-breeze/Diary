import { apiClient } from './client';
import type {
  DiaryEntry,
  DiaryEntryRequest,
  DiaryListResponse,
  SearchParams,
  SuggestTagsRequest,
  SuggestTagsResponse,
  TagsResponse,
} from '@/types';

export const diaryApi = {
  getItems: (params?: SearchParams) =>
    apiClient<DiaryListResponse>('/v1/items', { params: params as Record<string, string | undefined> }),

  getItemByDate: async (date: string): Promise<DiaryEntry | null> => {
    const response = await apiClient<DiaryListResponse>('/v1/items', {
      params: { date },
    });
    return response.items[0] || null;
  },

  saveItem: (item: DiaryEntryRequest) =>
    apiClient<DiaryEntry>('/v1/items', {
      method: 'PUT',
      body: item,
    }),

  search: (searchText: string, tags?: string) =>
    apiClient<DiaryListResponse>('/v1/items', {
      params: { search: searchText, tags },
    }),

  suggestTags: (req: SuggestTagsRequest) =>
    apiClient<SuggestTagsResponse>('/v1/items/suggest-tags', {
      method: 'POST',
      body: req,
    }),

  getTags: () => apiClient<TagsResponse>('/v1/tags'),
};
