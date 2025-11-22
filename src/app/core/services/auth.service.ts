import { Injectable, signal } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Router } from "@angular/router";
import { Observable, BehaviorSubject, tap, of, catchError, map } from "rxjs";
import { environment } from "../../../environments/environment";
import { User, AuthData, AuthResponse } from "../../shared/models";

@Injectable({
  providedIn: "root",
})
export class AuthService {
  private readonly TOKEN_KEY = "diary_auth_token";
  private readonly apiUrl = environment.apiUrl;

  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  public isAuthenticated = signal<boolean>(false);
  private authCheckInProgress = false;

  constructor(private http: HttpClient, private router: Router) {
    // Don't automatically load profile on initialization
    // Let the guard handle it to avoid race conditions
    const token = this.getToken();
    if (token) {
      this.isAuthenticated.set(true);
    }
  }

  login(credentials: AuthData): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(`${this.apiUrl}/authorize`, credentials)
      .pipe(
        tap((response) => {
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
    this.router.navigate(["/login"]);
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
        console.error("Failed to load user profile:", error);
        this.logout();
      },
    });
  }

  /**
   * Validates authentication by checking token and loading user profile if needed.
   * Returns an Observable that emits true if authenticated, false otherwise.
   * This should be used by the auth guard to ensure proper authentication before navigation.
   */
  validateAuthentication(): Observable<boolean> {
    const token = this.getToken();

    // No token means not authenticated
    if (!token) {
      this.isAuthenticated.set(false);
      return of(false);
    }

    // If user is already loaded, we're authenticated
    if (this.currentUserSubject.value) {
      this.isAuthenticated.set(true);
      return of(true);
    }

    // Load user profile to validate the token
    return this.http.get<User>(`${this.apiUrl}/user`).pipe(
      map((user) => {
        this.currentUserSubject.next(user);
        this.isAuthenticated.set(true);
        return true;
      }),
      catchError((error) => {
        console.error("Authentication validation failed:", error);
        this.removeToken();
        this.isAuthenticated.set(false);
        this.currentUserSubject.next(null);
        return of(false);
      })
    );
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }
}
