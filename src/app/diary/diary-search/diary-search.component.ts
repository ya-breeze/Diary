import { Component, OnInit, signal } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormBuilder, FormGroup, ReactiveFormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { DiaryService } from "../../core/services/diary.service";
import { AuthService } from "../../core/services/auth.service";
import { DiaryItem } from "../../shared/models";

@Component({
  selector: "app-diary-search",
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: "./diary-search.component.html",
  styleUrl: "./diary-search.component.css",
})
export class DiarySearchComponent implements OnInit {
  searchForm: FormGroup;
  searchResults = signal<DiaryItem[]>([]);
  isSearching = signal<boolean>(false);
  userEmail = signal<string>("");
  hasSearched = signal<boolean>(false);

  constructor(
    private fb: FormBuilder,
    private diaryService: DiaryService,
    private authService: AuthService,
    private router: Router
  ) {
    this.searchForm = this.fb.group({
      query: [""],
      tags: [""],
    });
  }

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    if (user) {
      this.userEmail.set(user.email);
    }
  }

  onSearch(): void {
    const query = this.searchForm.get("query")?.value?.trim() || "";
    const tagsInput = this.searchForm.get("tags")?.value?.trim() || "";
    const tags = tagsInput
      ? tagsInput.split(",").map((t: string) => t.trim())
      : [];

    if (!query && tags.length === 0) {
      return;
    }

    this.isSearching.set(true);
    this.hasSearched.set(true);

    this.diaryService.searchItems(query, tags).subscribe({
      next: (response) => {
        this.searchResults.set(response.items);
        this.isSearching.set(false);
      },
      error: (error) => {
        console.error("Search error:", error);
        this.isSearching.set(false);
        if (error.status === 401) {
          this.authService.logout();
        }
      },
    });
  }

  clearSearch(): void {
    this.searchForm.reset();
    this.searchResults.set([]);
    this.hasSearched.set(false);
  }

  viewEntry(date: string): void {
    this.router.navigate(["/diary"], { queryParams: { date } });
  }

  logout(): void {
    this.authService.logout();
  }
}
