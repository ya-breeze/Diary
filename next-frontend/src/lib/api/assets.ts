import { API_BASE_URL } from './client';
import type { AssetBatchResponse } from '@/types';

export const assetsApi = {
  uploadAssetsBatch: (files: File[], onProgress?: (percent: number) => void): Promise<AssetBatchResponse> => {
    return new Promise((resolve, reject) => {
      const formData = new FormData();
      files.forEach((file) => formData.append('assets', file));

      const xhr = new XMLHttpRequest();
      xhr.withCredentials = true;

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable && onProgress) {
          onProgress(Math.round((e.loaded / e.total) * 100));
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            resolve(JSON.parse(xhr.responseText));
          } catch {
            reject(new Error('Upload response was not valid JSON'));
          }
        } else {
          reject(new Error(`Batch upload failed: ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => reject(new Error('Upload network error')));
      xhr.addEventListener('abort', () => reject(new Error('Upload aborted')));

      xhr.open('POST', `${API_BASE_URL}/v1/assets/batch`);
      xhr.send(formData);
    });
  },

  getAssetUrl: (path: string): string => {
    // For serving through the web endpoint
    return `${API_BASE_URL}/v1/assets?path=${encodeURIComponent(path)}`;
  },
};
