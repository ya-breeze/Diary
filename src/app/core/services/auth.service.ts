import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, BehaviorSubject, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { User, AuthData, AuthResponse } from '../../shared/models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly TOKEN_KEY = 'diary_auth_token';
  private readonly apiUrl = environment.apiUrl;
  
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();
  
  public isAuthenticated = signal<boolean>(false);

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    // Check if token exists on initialization
    const token = this.getToken();
    if (token) {
      this.isAuthenticated.set(true);
      this.loadUserProfile();
    }
  }

  login(credentials: AuthData): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/authorize`, credentials).pipe(
      tap(response => {
        this.setToken(response.token);
        this.isAuthenticated.set(true);
        this.loadUserProfile();
      })
    );
  }

  logout(): void {
    this.removeToken();
    this.isAuthenticated.set(false);
    this.currentUserSubject.next(null);
    this.router.navigate(['/login']);
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  private setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  private removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
  }

  private loadUserProfile(): void {
    this.http.get<User>(`${this.apiUrl}/user`).subscribe({
      next: (user) => {
        this.currentUserSubject.next(user);
      },
      error: (error) => {
        console.error('Failed to load user profile:', error);
        this.logout();
      }
    });
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }
}

