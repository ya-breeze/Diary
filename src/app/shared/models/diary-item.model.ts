export interface DiaryItem {
  date: string;
  title: string;
  body: string;
  tags?: string[];
  previousDate?: string | null;
  nextDate?: string | null;
}

export interface DiaryItemRequest {
  date: string;
  title: string;
  body: string;
  tags?: string[];
}

export interface DiaryItemsListResponse {
  items: DiaryItem[];
  totalCount: number;
}

