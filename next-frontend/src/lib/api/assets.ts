import { API_BASE_URL } from './client';
import type { AssetBatchResponse } from '@/types';

export const assetsApi = {
  uploadAsset: async (file: File): Promise<string> => {
    const formData = new FormData();
    formData.append('asset', file);

    const response = await fetch(`${API_BASE_URL}/v1/assets`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.status}`);
    }

    return response.text();
  },

  uploadAssetsBatch: async (files: File[]): Promise<AssetBatchResponse> => {
    const formData = new FormData();
    files.forEach((file) => formData.append('assets', file));

    const response = await fetch(`${API_BASE_URL}/v1/assets/batch`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Batch upload failed: ${response.status}`);
    }

    return response.json();
  },

  getAssetUrl: (path: string): string => {
    // For serving through the web endpoint
    return `${API_BASE_URL}/v1/assets?path=${encodeURIComponent(path)}`;
  },
};
