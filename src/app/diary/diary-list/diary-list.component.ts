import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { DiaryService } from '../../core/services/diary.service';
import { AuthService } from '../../core/services/auth.service';
import { DiaryItem } from '../../shared/models';

@Component({
  selector: 'app-diary-list',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './diary-list.component.html',
  styleUrl: './diary-list.component.css'
})
export class DiaryListComponent implements OnInit {
  currentItem = signal<DiaryItem | null>(null);
  currentDate = signal<string>('');
  isLoading = signal<boolean>(false);
  userEmail = signal<string>('');

  constructor(
    private diaryService: DiaryService,
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    if (user) {
      this.userEmail.set(user.email);
    }

    this.currentDate.set(this.diaryService.currentDate());
    this.loadDiaryEntry(this.currentDate());

    // Subscribe to current item changes
    this.diaryService.currentItem$.subscribe(item => {
      this.currentItem.set(item);
    });
  }

  loadDiaryEntry(date: string): void {
    this.isLoading.set(true);
    this.diaryService.getItemByDate(date).subscribe({
      next: () => {
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Error loading diary entry:', error);
        this.isLoading.set(false);
      }
    });
  }

  onDateChange(event: Event): void {
    const input = event.target as HTMLInputElement;
    const newDate = input.value;
    this.currentDate.set(newDate);
    this.loadDiaryEntry(newDate);
  }

  goToPreviousDate(): void {
    const item = this.currentItem();
    if (item?.previousDate) {
      this.currentDate.set(item.previousDate);
      this.loadDiaryEntry(item.previousDate);
    }
  }

  goToNextDate(): void {
    const item = this.currentItem();
    if (item?.nextDate) {
      this.currentDate.set(item.nextDate);
      this.loadDiaryEntry(item.nextDate);
    }
  }

  logout(): void {
    this.authService.logout();
  }
}

