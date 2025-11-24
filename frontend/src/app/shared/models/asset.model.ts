export interface AssetUploadResponse {
  filename: string;
}

export interface AssetsBatchFile {
  originalName: string;
  savedName: string;
  size: number;
  contentType?: string;
}

export interface AssetsBatchResponse {
  files: AssetsBatchFile[];
  count: number;
}

export interface Asset {
  path: string;
  url: string;
}
