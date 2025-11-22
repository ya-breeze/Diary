import { Component, EventEmitter, Output, signal } from "@angular/core";
import { CommonModule } from "@angular/common";
import { AssetService } from "../../../core/services/asset.service";

export interface UploadedAsset {
  path: string;
  file: File;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  error?: string;
}

@Component({
  selector: "app-asset-upload",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./asset-upload.component.html",
  styleUrl: "./asset-upload.component.css",
})
export class AssetUploadComponent {
  @Output() uploadComplete = new EventEmitter<string[]>();

  files = signal<UploadedAsset[]>([]);
  isDragging = signal<boolean>(false);
  isUploading = signal<boolean>(false);

  constructor(private assetService: AssetService) {}

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
    if (droppedFiles) {
      this.handleFiles(Array.from(droppedFiles));
    }
  }

  onFileSelect(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files) {
      this.handleFiles(Array.from(input.files));
    }
  }

  private handleFiles(selectedFiles: File[]): void {
    const newFiles: UploadedAsset[] = selectedFiles.map((file) => ({
      path: "",
      file,
      progress: 0,
      status: "pending",
    }));

    this.files.update((current) => [...current, ...newFiles]);
  }

  uploadFiles(): void {
    const pendingFiles = this.files().filter((f) => f.status === "pending");
    if (pendingFiles.length === 0) return;

    this.isUploading.set(true);

    // Upload files one by one
    this.uploadNextFile(0);
  }

  private uploadNextFile(index: number): void {
    const currentFiles = this.files();
    const pendingFiles = currentFiles.filter((f) => f.status === "pending");

    if (index >= pendingFiles.length) {
      this.isUploading.set(false);
      const successfulPaths = currentFiles
        .filter((f) => f.status === "success")
        .map((f) => f.path);

      if (successfulPaths.length > 0) {
        this.uploadComplete.emit(successfulPaths);
      }
      return;
    }

    const fileToUpload = pendingFiles[index];
    const fileIndex = currentFiles.indexOf(fileToUpload);

    // Update status to uploading
    this.files.update((files) => {
      const updated = [...files];
      updated[fileIndex] = {
        ...updated[fileIndex],
        status: "uploading",
        progress: 0,
      };
      return updated;
    });

    // Upload the file
    this.assetService.uploadAsset(fileToUpload.file).subscribe({
      next: (path) => {
        this.files.update((files) => {
          const updated = [...files];
          updated[fileIndex] = {
            ...updated[fileIndex],
            status: "success",
            progress: 100,
            path,
          };
          return updated;
        });

        // Upload next file
        this.uploadNextFile(index + 1);
      },
      error: (error) => {
        this.files.update((files) => {
          const updated = [...files];
          updated[fileIndex] = {
            ...updated[fileIndex],
            status: "error",
            error: error.message || "Upload failed",
          };
          return updated;
        });

        // Continue with next file even if this one failed
        this.uploadNextFile(index + 1);
      },
    });
  }

  removeFile(index: number): void {
    this.files.update((files) => files.filter((_, i) => i !== index));
  }

  clearCompleted(): void {
    this.files.update((files) => files.filter((f) => f.status !== "success"));
  }

  clearAll(): void {
    this.files.set([]);
  }

  get hasFiles(): boolean {
    return this.files().length > 0;
  }

  get hasPendingFiles(): boolean {
    return this.files().some((f) => f.status === "pending");
  }

  get hasSuccessfulFiles(): boolean {
    return this.files().some((f) => f.status === "success");
  }

  formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  }
}
