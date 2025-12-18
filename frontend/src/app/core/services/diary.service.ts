import { Injectable, signal } from "@angular/core";
import { HttpClient, HttpParams } from "@angular/common/http";
import { Observable, BehaviorSubject, tap } from "rxjs";
import {
  DiaryItem,
  DiaryItemRequest,
  DiaryItemsListResponse,
} from "../../shared/models";
import { environment } from "../../../environments/environment";

@Injectable({
  providedIn: "root",
})
export class DiaryService {
  private get apiUrl(): string {
    return environment.apiUrl;
  }

  private currentItemSubject = new BehaviorSubject<DiaryItem | null>(null);
  public currentItem$ = this.currentItemSubject.asObservable();

  public currentDate = signal<string>(this.getTodayDate());

  constructor(private http: HttpClient) {}

  getItems(
    date?: string,
    search?: string,
    tags?: string[]
  ): Observable<DiaryItemsListResponse> {
    let params = new HttpParams();

    if (date) {
      params = params.set("date", date);
    }
    if (search) {
      params = params.set("search", search);
    }
    if (tags && tags.length > 0) {
      params = params.set("tags", tags.join(","));
    }

    return this.http.get<DiaryItemsListResponse>(`${this.apiUrl}/items`, {
      params,
    });
  }

  getItemByDate(date: string): Observable<DiaryItemsListResponse> {
    return this.getItems(date).pipe(
      tap((response) => {
        if (response.items.length > 0) {
          this.currentItemSubject.next(response.items[0]);
          this.currentDate.set(date);
        } else {
          // Create empty item for this date
          const emptyItem: DiaryItem = {
            date,
            title: "",
            body: "",
            tags: [],
          };
          this.currentItemSubject.next(emptyItem);
          this.currentDate.set(date);
        }
      })
    );
  }

  saveItem(item: DiaryItemRequest): Observable<DiaryItem> {
    return this.http.put<DiaryItem>(`${this.apiUrl}/items`, item).pipe(
      tap((savedItem) => {
        this.currentItemSubject.next(savedItem);
      })
    );
  }

  searchItems(
    searchText: string,
    tags?: string[]
  ): Observable<DiaryItemsListResponse> {
    return this.getItems(undefined, searchText, tags);
  }

  getCurrentItem(): DiaryItem | null {
    return this.currentItemSubject.value;
  }

  private getTodayDate(): string {
    const today = new Date();
    return today.toISOString().split("T")[0];
  }
}
