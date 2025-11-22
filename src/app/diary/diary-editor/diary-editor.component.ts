import {
  Component,
  OnInit,
  OnDestroy,
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
import { marked } from "marked";
import { DomSanitizer, SafeHtml } from "@angular/platform-browser";

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
  ],
  templateUrl: "./diary-editor.component.html",
  styleUrl: "./diary-editor.component.css",
})
export class DiaryEditorComponent implements OnInit, OnDestroy {
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
  private shortcutsSubscription?: Subscription;
  private bodySubscription?: Subscription;
  private cursorPosition: { start: number; end: number } | null = null;
  private bodyText = signal<string>("");
  markdownHtml = computed<SafeHtml>(() => {
    const body = this.bodyText();
    // Replace asset references with full URLs
    const processedBody = this.processAssetLinks(body);
    const html = marked.parse(processedBody, { async: false }) as string;
    return this.sanitizer.bypassSecurityTrustHtml(html);
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

  ngOnDestroy(): void {
    this.shortcutsSubscription?.unsubscribe();
    this.bodySubscription?.unsubscribe();
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
}
