import { Component, OnInit, signal, effect, computed } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
} from "@angular/forms";
import { Router } from "@angular/router";
import { DiaryService } from "../../core/services/diary.service";
import { AuthService } from "../../core/services/auth.service";
import { AssetService } from "../../core/services/asset.service";
import { DiaryItem } from "../../shared/models";
import { AssetUploadComponent } from "../../shared/components/asset-upload/asset-upload.component";
import { AssetGalleryComponent } from "../../shared/components/asset-gallery/asset-gallery.component";
import { AssetPreviewModalComponent } from "../../shared/components/asset-preview-modal/asset-preview-modal.component";
import { marked } from "marked";
import { DomSanitizer, SafeHtml } from "@angular/platform-browser";

@Component({
  selector: "app-diary-editor",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    AssetUploadComponent,
    AssetGalleryComponent,
    AssetPreviewModalComponent,
  ],
  templateUrl: "./diary-editor.component.html",
  styleUrl: "./diary-editor.component.css",
})
export class DiaryEditorComponent implements OnInit {
  diaryForm: FormGroup;
  currentDate = signal<string>("");
  currentItem = signal<DiaryItem | null>(null);
  isLoading = signal<boolean>(false);
  isSaving = signal<boolean>(false);
  userEmail = signal<string>("");
  saveMessage = signal<string>("");
  tagInput = signal<string>("");
  showAssetUpload = signal<boolean>(false);
  uploadedAssets = signal<string[]>([]);
  previewAssetPath = signal<string | null>(null);
  showMarkdownPreview = signal<boolean>(false);
  markdownHtml = computed<SafeHtml>(() => {
    const body = this.diaryForm.get("body")?.value || "";
    // Replace asset references with full URLs
    const processedBody = this.processAssetLinks(body);
    const html = marked.parse(processedBody, { async: false }) as string;
    return this.sanitizer.sanitize(1, html) || "";
  });

  constructor(
    private fb: FormBuilder,
    private diaryService: DiaryService,
    private authService: AuthService,
    private assetService: AssetService,
    private router: Router,
    private sanitizer: DomSanitizer
  ) {
    this.diaryForm = this.fb.group({
      title: ["", [Validators.required, Validators.maxLength(200)]],
      body: ["", [Validators.required]],
      tags: [[]],
    });

    // Auto-save effect when form changes
    effect(() => {
      if (this.diaryForm.dirty && !this.isSaving()) {
        // Debounce auto-save would go here
      }
    });
  }

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    if (user) {
      this.userEmail.set(user.email);
    }

    // Get today's date in YYYY-MM-DD format
    const today = new Date().toISOString().split("T")[0];
    this.currentDate.set(today);
    this.loadDiaryEntry(today);

    // Subscribe to current item changes
    this.diaryService.currentItem$.subscribe((item) => {
      this.currentItem.set(item);
      if (item) {
        this.diaryForm.patchValue({
          title: item.title || "",
          body: item.body || "",
          tags: item.tags || [],
        });
        this.diaryForm.markAsPristine();
      }
    });
  }

  loadDiaryEntry(date: string): void {
    this.isLoading.set(true);
    this.diaryService.getItemByDate(date).subscribe({
      next: () => {
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error("Error loading diary entry:", error);
        this.isLoading.set(false);
        // If entry doesn't exist, create a blank one
        this.currentItem.set({
          date: date,
          title: "",
          body: "",
          tags: [],
        });
      },
    });
  }

  onDateChange(event: Event): void {
    const input = event.target as HTMLInputElement;
    const newDate = input.value;

    if (this.diaryForm.dirty) {
      if (
        confirm(
          "You have unsaved changes. Do you want to save before changing the date?"
        )
      ) {
        this.saveEntry().subscribe(() => {
          this.currentDate.set(newDate);
          this.loadDiaryEntry(newDate);
        });
      } else {
        this.currentDate.set(newDate);
        this.loadDiaryEntry(newDate);
      }
    } else {
      this.currentDate.set(newDate);
      this.loadDiaryEntry(newDate);
    }
  }

  goToPreviousDate(): void {
    const item = this.currentItem();
    if (item?.previousDate) {
      this.navigateToDate(item.previousDate);
    }
  }

  goToNextDate(): void {
    const item = this.currentItem();
    if (item?.nextDate) {
      this.navigateToDate(item.nextDate);
    }
  }

  private navigateToDate(date: string): void {
    if (this.diaryForm.dirty) {
      if (
        confirm(
          "You have unsaved changes. Do you want to save before navigating?"
        )
      ) {
        this.saveEntry().subscribe(() => {
          this.currentDate.set(date);
          this.loadDiaryEntry(date);
        });
      } else {
        this.currentDate.set(date);
        this.loadDiaryEntry(date);
      }
    } else {
      this.currentDate.set(date);
      this.loadDiaryEntry(date);
    }
  }

  saveEntry() {
    if (this.diaryForm.valid) {
      this.isSaving.set(true);
      this.saveMessage.set("");

      const formValue = this.diaryForm.value;
      return this.diaryService.saveItem({
        date: this.currentDate(),
        title: formValue.title,
        body: formValue.body,
        tags: formValue.tags,
      });
    }
    throw new Error("Form is invalid");
  }

  onSave(): void {
    this.saveEntry().subscribe({
      next: () => {
        this.isSaving.set(false);
        this.saveMessage.set("Saved successfully!");
        this.diaryForm.markAsPristine();
        setTimeout(() => this.saveMessage.set(""), 3000);
      },
      error: (error) => {
        this.isSaving.set(false);
        this.saveMessage.set("Error saving entry. Please try again.");
        console.error("Error saving diary entry:", error);
      },
    });
  }

  addTag(): void {
    const tag = this.tagInput().trim();
    if (tag) {
      const currentTags = this.diaryForm.get("tags")?.value || [];
      if (!currentTags.includes(tag)) {
        this.diaryForm.patchValue({
          tags: [...currentTags, tag],
        });
        this.diaryForm.markAsDirty();
      }
      this.tagInput.set("");
    }
  }

  removeTag(tag: string): void {
    const currentTags = this.diaryForm.get("tags")?.value || [];
    this.diaryForm.patchValue({
      tags: currentTags.filter((t: string) => t !== tag),
    });
    this.diaryForm.markAsDirty();
  }

  onTagInputChange(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.tagInput.set(input.value);
  }

  onTagInputKeydown(event: KeyboardEvent): void {
    if (event.key === "Enter") {
      event.preventDefault();
      this.addTag();
    }
  }

  logout(): void {
    if (this.diaryForm.dirty) {
      if (
        confirm(
          "You have unsaved changes. Do you want to save before logging out?"
        )
      ) {
        this.saveEntry().subscribe(() => {
          this.authService.logout();
        });
      } else {
        this.authService.logout();
      }
    } else {
      this.authService.logout();
    }
  }

  get tags(): string[] {
    return this.diaryForm.get("tags")?.value || [];
  }

  toggleAssetUpload(): void {
    this.showAssetUpload.update((v) => !v);
  }

  toggleMarkdownPreview(): void {
    this.showMarkdownPreview.update((v) => !v);
  }

  private processAssetLinks(markdown: string): string {
    // Replace markdown image links ![](filename) with full asset URLs
    // Pattern: ![alt text](filename) or ![](filename)
    return markdown.replace(
      /!\[([^\]]*)\]\(([^)]+)\)/g,
      (match, altText, filename) => {
        // If it's already a full URL, don't modify it
        if (filename.startsWith("http://") || filename.startsWith("https://")) {
          return match;
        }
        // Convert to full asset URL
        const assetUrl = this.assetService.getAssetUrl(filename);
        return `![${altText}](${assetUrl})`;
      }
    );
  }

  onAssetsUploaded(paths: string[]): void {
    this.uploadedAssets.update((current) => [...current, ...paths]);
    this.showAssetUpload.set(false);
    this.saveMessage.set(`${paths.length} asset(s) uploaded successfully!`);
    setTimeout(() => this.saveMessage.set(""), 3000);
  }

  onAssetSelected(path: string): void {
    this.previewAssetPath.set(path);
  }

  onAssetDeleted(path: string): void {
    this.uploadedAssets.update((current) => current.filter((p) => p !== path));
    this.saveMessage.set("Asset deleted successfully!");
    setTimeout(() => this.saveMessage.set(""), 3000);
  }

  closePreview(): void {
    this.previewAssetPath.set(null);
  }

  downloadAsset(path: string): void {
    window.open(path, "_blank");
  }
}
