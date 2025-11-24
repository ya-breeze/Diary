import { Component, Input, Output, EventEmitter, signal } from "@angular/core";
import { CommonModule } from "@angular/common";
import { AssetService } from "../../../core/services/asset.service";

@Component({
  selector: "app-assets-panel",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./assets-panel.component.html",
  styleUrl: "./assets-panel.component.css",
})
export class AssetsPanelComponent {
  @Input() assets: string[] = [];
  @Output() assetClick = new EventEmitter<string>();
  @Output() uploadAssets = new EventEmitter<File[]>();

  isDragging = signal<boolean>(false);

  constructor(private assetService: AssetService) {}

  getAssetUrl(path: string): string {
    return this.assetService.getAssetUrl(path);
  }

  isImage(path: string): boolean {
    const ext = path.split(".").pop()?.toLowerCase();
    return ["jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"].includes(
      ext || ""
    );
  }

  isVideo(path: string): boolean {
    const ext = path.split(".").pop()?.toLowerCase();
    return ["mp4", "webm", "ogg", "mov", "avi"].includes(ext || "");
  }

  onAssetClick(path: string): void {
    this.assetClick.emit(path);
  }

  onDragOver(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(true);
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(false);
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(false);

    const droppedFiles = event.dataTransfer?.files;
    if (droppedFiles && droppedFiles.length > 0) {
      this.uploadAssets.emit(Array.from(droppedFiles));
    }
  }

  onFileSelect(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.uploadAssets.emit(Array.from(input.files));
      // Reset input so the same file can be selected again
      input.value = "";
    }
  }
}

