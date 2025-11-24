import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable } from "rxjs";
import { AssetUploadResponse, AssetsBatchResponse } from "../../shared/models";
import { ConfigService } from "./config.service";

@Injectable({
  providedIn: "root",
})
export class AssetService {
  private get apiUrl(): string {
    return this.configService.getApiUrl();
  }

  private get baseUrl(): string {
    return this.apiUrl.replace("/v1", "");
  }

  constructor(private http: HttpClient, private configService: ConfigService) {}

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
