import { Component, Input, Output, EventEmitter, signal } from "@angular/core";
import { CommonModule } from "@angular/common";
import { AssetService } from "../../../core/services/asset.service";

export interface GalleryAsset {
  path: string;
  type: "image" | "video" | "unknown";
  thumbnail?: string;
}

@Component({
  selector: "app-asset-gallery",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./asset-gallery.component.html",
  styleUrl: "./asset-gallery.component.css",
})
export class AssetGalleryComponent {
  @Input() set assets(value: string[]) {
    this.galleryAssets.set(this.processAssets(value));
  }

  @Output() assetSelected = new EventEmitter<string>();
  @Output() assetDeleted = new EventEmitter<string>();

  galleryAssets = signal<GalleryAsset[]>([]);
  selectedAsset = signal<GalleryAsset | null>(null);

  constructor(private assetService: AssetService) {}

  private processAssets(paths: string[]): GalleryAsset[] {
    return paths.map((path) => ({
      path,
      type: this.getAssetType(path),
      thumbnail: this.assetService.getAssetUrl(path),
    }));
  }

  private getAssetType(path: string): "image" | "video" | "unknown" {
    const ext = path.split(".").pop()?.toLowerCase();

    const imageExts = ["jpg", "jpeg", "png", "gif", "webp", "svg", "bmp"];
    const videoExts = ["mp4", "webm", "ogg", "mov", "avi"];

    if (ext && imageExts.includes(ext)) {
      return "image";
    } else if (ext && videoExts.includes(ext)) {
      return "video";
    }

    return "unknown";
  }

  selectAsset(asset: GalleryAsset): void {
    this.selectedAsset.set(asset);
    this.assetSelected.emit(asset.path);
  }

  deleteAsset(asset: GalleryAsset, event: Event): void {
    event.stopPropagation();

    if (confirm(`Are you sure you want to delete ${asset.path}?`)) {
      this.assetDeleted.emit(asset.path);
    }
  }

  getAssetUrl(path: string): string {
    return this.assetService.getAssetUrl(path);
  }

  copyAssetPath(asset: GalleryAsset, event: Event): void {
    event.stopPropagation();

    const markdownLink =
      asset.type === "image"
        ? `![${asset.path}](${asset.path})`
        : `[${asset.path}](${asset.path})`;

    navigator.clipboard.writeText(markdownLink).then(() => {
      alert("Markdown link copied to clipboard!");
    });
  }

  get hasAssets(): boolean {
    return this.galleryAssets().length > 0;
  }

  getFileName(path: string): string {
    return path.split("/").pop() || path;
  }
}
