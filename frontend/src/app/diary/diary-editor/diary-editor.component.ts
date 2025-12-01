import {
  Component,
  OnInit,
  OnDestroy,
  AfterViewChecked,
  HostListener,
  signal,
  effect,
  computed,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
} from "@angular/forms";
import { Router, RouterModule } from "@angular/router";
import { Subscription } from "rxjs";
import { DiaryService } from "../../core/services/diary.service";
import { AuthService } from "../../core/services/auth.service";
import { AssetService } from "../../core/services/asset.service";
import { ToastService } from "../../core/services/toast.service";
import { KeyboardShortcutsService } from "../../core/services/keyboard-shortcuts.service";
import { DiaryItem } from "../../shared/models";
import { AssetUploadComponent } from "../../shared/components/asset-upload/asset-upload.component";
import { AssetGalleryComponent } from "../../shared/components/asset-gallery/asset-gallery.component";
import { AssetPreviewModalComponent } from "../../shared/components/asset-preview-modal/asset-preview-modal.component";
import { LoadingSpinnerComponent } from "../../shared/components/loading-spinner/loading-spinner.component";
import { KeyboardShortcutsHelpComponent } from "../../shared/components/keyboard-shortcuts-help/keyboard-shortcuts-help.component";
import { ThemeToggleComponent } from "../../shared/components/theme-toggle/theme-toggle.component";
import { AssetsPanelComponent } from "../../shared/components/assets-panel/assets-panel.component";
import { marked } from "marked";
import { DomSanitizer, SafeHtml } from "@angular/platform-browser";
import DOMPurify from "dompurify";
import {
  extractAssetsFromMarkdown,
  appendAssetsToMarkdown,
} from "../../shared/utils/markdown-parser";

export type EditorMode = "view" | "edit";

@Component({
  selector: "app-diary-editor",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterModule,
    AssetUploadComponent,
    AssetGalleryComponent,
    AssetPreviewModalComponent,
    LoadingSpinnerComponent,
    KeyboardShortcutsHelpComponent,
    ThemeToggleComponent,
    AssetsPanelComponent,
  ],
  templateUrl: "./diary-editor.component.html",
  styleUrl: "./diary-editor.component.css",
})
export class DiaryEditorComponent
  implements OnInit, OnDestroy, AfterViewChecked
{
  diaryForm: FormGroup;
  currentDate = signal<string>("");
  currentItem = signal<DiaryItem | null>(null);
  isLoading = signal<boolean>(false);
  isSaving = signal<boolean>(false);
  userEmail = signal<string>("");
  tagInput = signal<string>("");
  showAssetUpload = signal<boolean>(false);
  uploadedAssets = signal<string[]>([]);
  previewAssetPath = signal<string | null>(null);
  showMarkdownPreview = signal<boolean>(false);
  showKeyboardHelp = signal<boolean>(false);
  currentMode = signal<EditorMode>("view"); // Default to view mode
  mediaList = signal<string[]>([]); // List of all images and videos in current entry
  currentMediaIndex = signal<number>(0); // Current media index in the list
  private shortcutsSubscription?: Subscription;
  private bodySubscription?: Subscription;
  private cursorPosition: { start: number; end: number } | null = null;
  private bodyText = signal<string>("");
  markdownHtml = computed<SafeHtml>(() => {
    const body = this.bodyText();
    // Replace asset references with full URLs
    const processedBody = this.processAssetLinks(body);
    const html = marked.parse(processedBody, { async: false }) as string;
    // Sanitize HTML using DOMPurify to prevent XSS attacks
    // Whitelist safe HTML tags and attributes commonly used in markdown
    const sanitizedHtml = DOMPurify.sanitize(html, {
      ALLOWED_TAGS: [
        "p",
        "br",
        "strong",
        "em",
        "u",
        "h1",
        "h2",
        "h3",
        "h4",
        "h5",
        "h6",
        "ul",
        "ol",
        "li",
        "blockquote",
        "code",
        "pre",
        "a",
        "img",
        "video",
        "source",
        "table",
        "thead",
        "tbody",
        "tr",
        "th",
        "td",
        "hr",
      ],
      ALLOWED_ATTR: [
        "href",
        "title",
        "src",
        "alt",
        "controls",
        "type",
        "colspan",
        "rowspan",
      ],
      ALLOW_DATA_ATTR: false,
    });
    return this.sanitizer.bypassSecurityTrustHtml(sanitizedHtml);
  });
  // Computed signal for assets in current entry
  currentAssets = computed<string[]>(() => {
    const body = this.bodyText();
    return extractAssetsFromMarkdown(body);
  });

  @HostListener("window:keydown", ["$event"])
  handleKeyDown(event: KeyboardEvent): void {
    // Don't handle shortcuts when typing in input fields
    const target = event.target as HTMLElement;
    if (
      target.tagName === "INPUT" ||
      target.tagName === "TEXTAREA" ||
      target.isContentEditable
    ) {
      // Allow Ctrl+S even in input fields
      if (event.key === "s" && event.ctrlKey) {
        event.preventDefault();
        this.onSave();
      }
      // Allow Ctrl+P for preview toggle
      if (event.key === "p" && event.ctrlKey) {
        event.preventDefault();
        this.toggleMarkdownPreview();
      }
      return;
    }

    this.keyboardShortcutsService.handleKeyboardEvent(event);
  }

  constructor(
    private fb: FormBuilder,
    private diaryService: DiaryService,
    private authService: AuthService,
    private assetService: AssetService,
    private toastService: ToastService,
    private keyboardShortcutsService: KeyboardShortcutsService,
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

    // Load mode preference (defaults to 'view')
    const savedMode = this.loadModePref();
    this.currentMode.set(savedMode);

    // Keep a live copy of the body text for preview so it always reflects
    // the current editor content, not only what was last saved/loaded.
    this.bodySubscription = this.diaryForm
      .get("body")!
      .valueChanges.subscribe((value: string) => {
        this.bodyText.set(value || "");
      });

    // Initialize bodyText with the current form value
    const initialBody = this.diaryForm.get("body")?.value || "";
    this.bodyText.set(initialBody);

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

        // Also update live body text for preview when a new item loads
        this.bodyText.set(item.body || "");
      }
    });

    // Subscribe to keyboard shortcuts
    this.shortcutsSubscription =
      this.keyboardShortcutsService.shortcuts$.subscribe((action) => {
        this.handleShortcutAction(action);
      });
  }

  ngAfterViewChecked(): void {
    // Add click handlers to images in view mode
    if (this.currentMode() === "view") {
      this.attachImageClickHandlers();
    }
  }

  ngOnDestroy(): void {
    this.shortcutsSubscription?.unsubscribe();
    this.bodySubscription?.unsubscribe();
  }

  private attachImageClickHandlers(): void {
    const viewBody = document.querySelector(".view-body");
    if (!viewBody) return;

    const images = viewBody.querySelectorAll("img");
    const videos = viewBody.querySelectorAll("video");

    // Build list of all media (images and videos) filenames
    const mediaFilenames: string[] = [];
    const mediaElements: (HTMLImageElement | HTMLVideoElement)[] = [];

    // Add images
    images.forEach((img) => {
      const src = img.getAttribute("src");
      if (src) {
        const filename = src.split("/").pop();
        if (filename) {
          mediaFilenames.push(filename);
          mediaElements.push(img);
        }
      }
    });

    // Add videos
    videos.forEach((video) => {
      const source = video.querySelector("source");
      if (source) {
        const src = source.getAttribute("src");
        if (src) {
          const filename = src.split("/").pop();
          if (filename) {
            mediaFilenames.push(filename);
            mediaElements.push(video);
          }
        }
      }
    });

    this.mediaList.set(mediaFilenames);

    // Add click handlers to images
    images.forEach((img) => {
      // Find the index in the combined media list
      const mediaIndex = mediaElements.indexOf(img);

      // Remove existing listener if any
      const clone = img.cloneNode(true) as HTMLImageElement;
      img.parentNode?.replaceChild(clone, img);

      // Add click handler
      clone.style.cursor = "pointer";
      clone.addEventListener("click", (event) => {
        event.preventDefault();
        const src = clone.getAttribute("src");
        if (src) {
          // Extract just the filename from the URL
          // URL format: http://localhost:8080/v1/assets/9232aa57-eab9-4b2f-975b-3bfd6d421474.jpg
          const filename = src.split("/").pop();
          if (filename) {
            this.currentMediaIndex.set(mediaIndex);
            this.openImagePreview(filename);
          }
        }
      });
    });

    // Add click handlers to videos
    videos.forEach((video) => {
      const mediaIndex = mediaElements.indexOf(video);

      // Remove existing listener if any
      const clone = video.cloneNode(true) as HTMLVideoElement;
      video.parentNode?.replaceChild(clone, video);

      // Add click handler
      clone.style.cursor = "pointer";
      clone.addEventListener("click", (event) => {
        event.preventDefault();
        const source = clone.querySelector("source");
        if (source) {
          const src = source.getAttribute("src");
          if (src) {
            const filename = src.split("/").pop();
            if (filename) {
              this.currentMediaIndex.set(mediaIndex);
              this.openImagePreview(filename);
            }
          }
        }
      });
    });
  }

  private openImagePreview(imagePath: string): void {
    this.previewAssetPath.set(imagePath);
  }

  navigateImage(direction: "next" | "previous"): void {
    const media = this.mediaList();
    if (media.length === 0) return;

    let newIndex = this.currentMediaIndex();
    if (direction === "next") {
      newIndex = (newIndex + 1) % media.length;
    } else {
      newIndex = (newIndex - 1 + media.length) % media.length;
    }

    this.currentMediaIndex.set(newIndex);
    this.previewAssetPath.set(media[newIndex]);
  }

  handleShortcutAction(action: string): void {
    switch (action) {
      case "save":
        this.onSave();
        break;
      case "previous":
        this.goToPreviousDate();
        break;
      case "next":
        this.goToNextDate();
        break;
      case "preview":
        this.toggleMarkdownPreview();
        break;
      case "toggleMode":
        this.toggleMode();
        break;
      case "search":
        this.router.navigate(["/diary/search"]);
        break;
      case "help":
        this.toggleKeyboardHelp();
        break;
      case "escape":
        if (this.showKeyboardHelp()) {
          this.toggleKeyboardHelp();
        } else if (this.previewAssetPath()) {
          this.closePreview();
        }
        break;
    }
  }

  toggleKeyboardHelp(): void {
    this.showKeyboardHelp.set(!this.showKeyboardHelp());
  }

  toggleMode(): void {
    const newMode: EditorMode = this.currentMode() === "view" ? "edit" : "view";

    // If switching from edit to view and there are unsaved changes, prompt user
    if (this.currentMode() === "edit" && this.diaryForm.dirty) {
      if (
        confirm(
          "You have unsaved changes. Do you want to save before switching to view mode?"
        )
      ) {
        this.saveEntry().subscribe(() => {
          this.currentMode.set(newMode);
          this.saveModePref(newMode);
        });
      } else {
        this.currentMode.set(newMode);
        this.saveModePref(newMode);
      }
    } else {
      this.currentMode.set(newMode);
      this.saveModePref(newMode);
    }
  }

  switchToEditMode(): void {
    this.currentMode.set("edit");
    this.saveModePref("edit");
  }

  switchToViewMode(): void {
    if (this.diaryForm.dirty) {
      if (
        confirm(
          "You have unsaved changes. Do you want to save before switching to view mode?"
        )
      ) {
        this.saveEntry().subscribe(() => {
          this.currentMode.set("view");
          this.saveModePref("view");
        });
      } else {
        this.currentMode.set("view");
        this.saveModePref("view");
      }
    } else {
      this.currentMode.set("view");
      this.saveModePref("view");
    }
  }

  private saveModePref(mode: EditorMode): void {
    localStorage.setItem("diary-editor-mode", mode);
  }

  private loadModePref(): EditorMode {
    const saved = localStorage.getItem("diary-editor-mode");
    // Default to 'view' mode as per requirements
    return saved === "edit" || saved === "view" ? saved : "view";
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
        this.toastService.success("Diary entry saved successfully!");
        this.diaryForm.markAsPristine();
      },
      error: (error) => {
        this.isSaving.set(false);
        this.toastService.error("Error saving entry. Please try again.");
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
    const wasShowingPreview = this.showMarkdownPreview();

    if (!wasShowingPreview) {
      // Switching to preview mode - save cursor position
      const textarea = document.getElementById("body") as HTMLTextAreaElement;
      if (textarea) {
        this.cursorPosition = {
          start: textarea.selectionStart,
          end: textarea.selectionEnd,
        };
      }
    }

    this.showMarkdownPreview.update((v) => !v);

    if (wasShowingPreview) {
      // Switching back to edit mode - restore cursor position
      setTimeout(() => {
        const textarea = document.getElementById("body") as HTMLTextAreaElement;
        if (textarea && this.cursorPosition) {
          textarea.focus();
          textarea.setSelectionRange(
            this.cursorPosition.start,
            this.cursorPosition.end
          );
        }
      }, 0);
    }
  }

  private processAssetLinks(markdown: string): string {
    // Replace markdown image syntax ![](filename) with either:
    // - Full asset URL for images
    // - HTML5 video tag for videos
    // Pattern: ![alt text](filename) or ![](filename)
    return markdown.replace(
      /!\[([^\]]*)\]\(([^)]+)\)/g,
      (match, altText, filename) => {
        // If it's already a full URL, don't modify it
        if (filename.startsWith("http://") || filename.startsWith("https://")) {
          return match;
        }

        // Check if it's a video file
        const ext = filename.split(".").pop()?.toLowerCase();
        const videoExts = ["mp4", "webm", "ogg", "mov", "avi"];

        if (ext && videoExts.includes(ext)) {
          // Convert to HTML5 video tag for videos
          const videoUrl = this.assetService.getAssetUrl(filename);
          return `<video controls><source src="${videoUrl}" type="video/${ext}"></video>`;
        } else {
          // Convert to full asset URL for images
          const assetUrl = this.assetService.getAssetUrl(filename);
          return `![${altText}](${assetUrl})`;
        }
      }
    );
  }

  onAssetsUploaded(paths: string[]): void {
    this.uploadedAssets.update((current) => [...current, ...paths]);
    this.showAssetUpload.set(false);
    this.toastService.success(
      `${paths.length} asset(s) uploaded successfully!`
    );
  }

  onAssetSelected(path: string): void {
    this.previewAssetPath.set(path);
  }

  onAssetDeleted(path: string): void {
    this.uploadedAssets.update((current) => current.filter((p) => p !== path));
    this.toastService.success("Asset deleted successfully!");
  }

  closePreview(): void {
    this.previewAssetPath.set(null);
  }

  downloadAsset(path: string): void {
    window.open(path, "_blank");
  }

  onBatchUpload(files: File[]): void {
    if (!files || files.length === 0) {
      return;
    }

    // Upload files using batch endpoint
    this.assetService.uploadAssetsBatch(files).subscribe({
      next: (response) => {
        // Extract saved filenames from the response
        const savedNames = response.files.map((file) => file.savedName);

        // Append asset references to the body text
        const currentBody = this.diaryForm.get("body")?.value || "";
        const newBody = appendAssetsToMarkdown(currentBody, savedNames);
        this.diaryForm.patchValue({ body: newBody });
        this.diaryForm.markAsDirty();

        this.toastService.success(
          `${response.count} asset(s) uploaded and added to entry!`
        );
      },
      error: (error) => {
        console.error("Error uploading assets:", error);
        this.toastService.error("Error uploading assets. Please try again.");
      },
    });
  }
}
