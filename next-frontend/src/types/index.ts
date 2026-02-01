// API Types based on OpenAPI spec

export interface AuthData {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
}

export interface User {
  id: string;
  email: string;
  startDate: string;
}

export interface DiaryEntry {
  date: string;
  title: string;
  tags: string[];
  body: string;
  previousDate?: string | null;
  nextDate?: string | null;
}

export interface DiaryEntryRequest {
  date: string;
  title: string;
  tags: string[];
  body: string;
}

export interface DiaryListResponse {
  items: DiaryEntry[];
  totalCount: number;
}

export interface AssetBatchFile {
  originalName: string;
  savedName: string;
  size: number;
  contentType?: string;
}

export interface AssetBatchResponse {
  files: AssetBatchFile[];
  count: number;
}

export interface SearchParams {
  date?: string;
  search?: string;
  tags?: string;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}
