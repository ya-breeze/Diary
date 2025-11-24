import {
  Component,
  Input,
  Output,
  EventEmitter,
  signal,
  effect,
  ChangeDetectionStrategy,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import { AssetService } from "../../../core/services/asset.service";

@Component({
  selector: "app-asset-preview-modal",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./asset-preview-modal.component.html",
  styleUrl: "./asset-preview-modal.component.css",
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AssetPreviewModalComponent {
  @Input() set assetPath(value: string | null) {
    this._assetPath.set(value);
    if (value) {
      this.assetType.set(this.getAssetType(value));
      this.assetUrl.set(this.assetService.getAssetUrl(value));
    }
  }

  @Input() viewOnly = false; // Hide edit actions in view mode
  @Input() hasMultipleImages = false; // Whether there are multiple images to navigate

  @Output() close = new EventEmitter<void>();
  @Output() delete = new EventEmitter<string>();
  @Output() download = new EventEmitter<string>();
  @Output() navigateNext = new EventEmitter<void>();
  @Output() navigatePrevious = new EventEmitter<void>();

  private _assetPath = signal<string | null>(null);
  assetType = signal<"image" | "video" | "unknown">("unknown");
  assetUrl = signal<string>("");

  constructor(private assetService: AssetService) {
    // Handle keyboard events (Escape, Arrow keys)
    effect(() => {
      const handleKeyboard = (event: KeyboardEvent) => {
        if (!this._assetPath()) return;

        if (event.key === "Escape") {
          this.closeModal();
        } else if (event.key === "ArrowRight" && this.hasMultipleImages) {
          event.preventDefault();
          this.navigateNext.emit();
        } else if (event.key === "ArrowLeft" && this.hasMultipleImages) {
          event.preventDefault();
          this.navigatePrevious.emit();
        }
      };

      if (this._assetPath()) {
        document.addEventListener("keydown", handleKeyboard);
        return () => {
          document.removeEventListener("keydown", handleKeyboard);
        };
      }
      return undefined;
    });
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

  closeModal(): void {
    this.close.emit();
  }

  deleteAsset(): void {
    const path = this._assetPath();
    if (path && confirm(`Are you sure you want to delete ${path}?`)) {
      this.delete.emit(path);
      this.closeModal();
    }
  }

  downloadAsset(): void {
    const path = this._assetPath();
    if (path) {
      this.download.emit(path);
    }
  }

  copyMarkdownLink(): void {
    const path = this._assetPath();
    if (!path) return;

    const type = this.assetType();
    const markdownLink =
      type === "image" ? `![${path}](${path})` : `[${path}](${path})`;

    navigator.clipboard.writeText(markdownLink).then(() => {
      alert("Markdown link copied to clipboard!");
    });
  }

  onBackdropClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      this.closeModal();
    }
  }

  onImageClick(event: MouseEvent): void {
    // In view-only mode, clicking the image closes the modal
    if (this.viewOnly) {
      event.stopPropagation();
      this.closeModal();
    }
  }

  get isOpen(): boolean {
    return this._assetPath() !== null;
  }

  get assetPath(): string | null {
    return this._assetPath();
  }

  get fileName(): string {
    const path = this._assetPath();
    return path ? path.split("/").pop() || path : "";
  }
}
