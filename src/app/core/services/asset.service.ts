import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable } from "rxjs";
import { environment } from "../../../environments/environment";
import { AssetUploadResponse, AssetsBatchResponse } from "../../shared/models";

@Injectable({
  providedIn: "root",
})
export class AssetService {
  private readonly apiUrl = environment.apiUrl;
  private readonly baseUrl = environment.apiUrl.replace("/v1", "");

  constructor(private http: HttpClient) {}

  uploadAsset(file: File): Observable<string> {
    const formData = new FormData();
    formData.append("asset", file);

    return this.http.post(`${this.apiUrl}/assets`, formData, {
      responseType: "text",
    });
  }

  uploadAssetsBatch(files: File[]): Observable<AssetsBatchResponse> {
    const formData = new FormData();
    files.forEach((file) => {
      formData.append("assets", file);
    });

    return this.http.post<AssetsBatchResponse>(
      `${this.apiUrl}/assets/batch`,
      formData
    );
  }

  getAssetUrl(path: string): string {
    // For rendering in markdown/HTML, use the web assets path
    return `${this.baseUrl}/web/assets/${path}`;
  }

  getAssetApiUrl(path: string): string {
    // For API calls (download, etc.), use the API endpoint
    return `${this.apiUrl}/assets?path=${encodeURIComponent(path)}`;
  }

  downloadAsset(path: string): Observable<Blob> {
    return this.http.get(this.getAssetApiUrl(path), {
      responseType: "blob",
    });
  }
}
