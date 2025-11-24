---
type: "always_apply"
---

# Frontend Development Guidelines (Angular)

This document defines the development standards, patterns, and best practices for the Diary frontend application written in Angular.

## Project Overview

- **Framework**: Angular 20+
- **Language**: TypeScript 5.8 (strict mode enabled)
- **State Management**: Angular Signals + RxJS 7.8
- **Architecture**: Standalone components (no NgModules)
- **Styling**: CSS with theme support (light/dark mode)
- **Testing**: Jasmine + Karma

## Architecture Principles

### Standalone Components Architecture

The frontend uses Angular's modern standalone components approach:

- **No NgModules**: All components are standalone
- **Direct imports**: Import dependencies directly in component metadata
- **Simplified structure**: Reduced boilerplate and clearer dependencies

### Layered Architecture

```
frontend/src/app/
├── core/                    # Singleton services, guards, interceptors
│   ├── guards/             # Route guards (auth)
│   ├── interceptors/       # HTTP interceptors (auth, error)
│   ├── services/           # Core services (auth, diary, asset, theme, toast, keyboard)
│   └── navbar/             # Core navigation component
├── shared/                 # Shared components and models
│   ├── components/         # Reusable components
│   └── models/             # TypeScript interfaces
├── auth/                   # Authentication feature
│   └── login/              # Login component
└── diary/                  # Diary feature
    ├── diary-editor/       # Main editor component
    ├── diary-list/         # List view component
    └── diary-search/       # Search component
```

## Code Standards

### Component Structure

All components should follow this pattern:

```typescript
import { Component, signal } from "@angular/core";
import { CommonModule } from "@angular/common";

@Component({
  selector: "app-component-name",
  standalone: true,
  imports: [CommonModule /* other imports */],
  templateUrl: "./component-name.component.html",
  styleUrl: "./component-name.component.css",
})
export class ComponentNameComponent {
  // Use signals for reactive state
  isLoading = signal<boolean>(false);
  data = signal<DataType | null>(null);

  constructor(private service: SomeService) {}

  // Methods
}
```

### State Management

#### Angular Signals (Preferred)

Use signals for component-local state:

```typescript
// Writable signals for mutable state
isLoading = signal<boolean>(false);
currentDate = signal<string>(this.getTodayDate());

// Update signals
this.isLoading.set(true);
this.currentDate.update(date => this.formatDate(date));

// Read signals in templates
@if (isLoading()) {
  <app-loading-spinner />
}
```

#### RxJS Observables

Use observables for:

- HTTP requests
- Event streams
- Complex async operations
- Service-to-component communication

```typescript
// BehaviorSubject for shared state
private currentUserSubject = new BehaviorSubject<User | null>(null);
public currentUser$ = this.currentUserSubject.asObservable();

// Observable patterns
this.authService.login(credentials).subscribe({
  next: (response) => { /* handle success */ },
  error: (error) => { /* handle error */ }
});
```

### Service Pattern

All services should be:

- Decorated with `@Injectable({ providedIn: 'root' })`
- Singleton (provided in root)
- Focused on a single responsibility

```typescript
@Injectable({
  providedIn: "root",
})
export class DiaryService {
  private readonly apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getItems(date?: string): Observable<DiaryItemsListResponse> {
    // Implementation
  }
}
```

## Authentication & Authorization

### JWT Token Management

- Store JWT tokens in `localStorage` with key `diary_auth_token`
- Include tokens in HTTP requests via `authInterceptor`
- Clear tokens on logout
- Validate authentication state before protected routes

### Auth Guard Pattern

```typescript
export const authGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.validateAuthentication().pipe(
    map((isAuthenticated) => {
      if (isAuthenticated) return true;
      router.navigate(["/login"], { queryParams: { returnUrl: state.url } });
      return false;
    })
  );
};
```

### HTTP Interceptors

#### Auth Interceptor

- Add `Authorization: Bearer <token>` header to all requests
- Include credentials (cookies) with `withCredentials: true`
- Skip token for `/authorize` endpoint

#### Error Interceptor

- Handle 401 errors globally
- Redirect to login on authentication failure
- Log errors for debugging

## Routing

### Route Configuration

```typescript
export const routes: Routes = [
  { path: "", redirectTo: "/diary", pathMatch: "full" },
  { path: "login", component: LoginComponent },
  {
    path: "diary",
    component: DiaryEditorComponent,
    canActivate: [authGuard],
  },
  { path: "**", redirectTo: "/diary" },
];
```

### Navigation Patterns

- Use `Router.navigate()` for programmatic navigation
- Use `routerLink` directive in templates
- Include return URLs for login redirects
- Use query parameters for state preservation

## Template Syntax

### Modern Control Flow

Use Angular's new control flow syntax (Angular 17+):

```html
<!-- Conditionals -->
@if (isLoading()) {
<app-loading-spinner />
} @else {
<div>Content</div>
}

<!-- Loops -->
@for (item of items(); track item.id) {
<div>{{ item.name }}</div>
} @empty {
<p>No items found</p>
}

<!-- Switch -->
@switch (status()) { @case ('loading') { <app-spinner /> } @case ('error') {
<app-error /> } @default { <app-content /> } }
```

### Template Best Practices

- Use `track` in `@for` loops for performance
- Prefer signals over observables in templates when possible
- Use `async` pipe for observables
- Keep templates simple - move complex logic to component
- Use semantic HTML elements
- Include ARIA attributes for accessibility

## Forms

### Reactive Forms Pattern

```typescript
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
} from "@angular/forms";

export class LoginComponent {
  loginForm: FormGroup;

  constructor(private fb: FormBuilder) {
    this.loginForm = this.fb.group({
      email: ["", [Validators.required, Validators.email]],
      password: ["", [Validators.required, Validators.minLength(3)]],
    });
  }

  onSubmit(): void {
    if (this.loginForm.valid) {
      // Process form
    }
  }

  get email() {
    return this.loginForm.get("email");
  }
}
```

### Form Validation

- Use built-in validators when possible
- Create custom validators for complex validation
- Display validation errors clearly
- Disable submit button when form is invalid
- Show validation feedback on blur or submit

## HTTP Communication

### API Service Pattern

```typescript
@Injectable({ providedIn: "root" })
export class DiaryService {
  private readonly apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getItems(date?: string): Observable<DiaryItemsListResponse> {
    let params = new HttpParams();
    if (date) {
      params = params.set("date", date);
    }
    return this.http.get<DiaryItemsListResponse>(`${this.apiUrl}/items`, {
      params,
    });
  }

  saveItem(item: DiaryItemRequest): Observable<DiaryItem> {
    return this.http.put<DiaryItem>(`${this.apiUrl}/items`, item).pipe(
      tap((savedItem) => {
        // Update local state
      })
    );
  }
}
```

### Error Handling

- Use RxJS `catchError` operator for request-specific errors
- Use error interceptor for global error handling
- Provide user-friendly error messages
- Log errors for debugging

## Asset Management

### File Upload Pattern

```typescript
uploadAsset(file: File): Observable<string> {
  const formData = new FormData();
  formData.append('asset', file);

  return this.http.post(`${this.apiUrl}/assets`, formData, {
    responseType: 'text'
  });
}

uploadAssetsBatch(files: File[]): Observable<AssetsBatchResponse> {
  const formData = new FormData();
  files.forEach(file => {
    formData.append('assets', file);
  });

  return this.http.post<AssetsBatchResponse>(
    `${this.apiUrl}/assets/batch`,
    formData
  );
}
```

### Asset URL Handling

- Use service methods to generate asset URLs
- Separate display URLs from API URLs
- Handle asset paths consistently
- Validate file types before upload

## Styling

### CSS Organization

- Use component-scoped CSS files
- Follow BEM naming convention for classes
- Use CSS custom properties for theming
- Keep styles modular and reusable

### Theme Support

```typescript
@Injectable({ providedIn: "root" })
export class ThemeService {
  private readonly THEME_KEY = "diary_theme";
  currentTheme = signal<"light" | "dark">("light");

  constructor() {
    this.loadTheme();
  }

  toggleTheme(): void {
    const newTheme = this.currentTheme() === "light" ? "dark" : "light";
    this.currentTheme.set(newTheme);
    this.applyTheme(newTheme);
    localStorage.setItem(this.THEME_KEY, newTheme);
  }
}
```

### CSS Variables for Themes

```css
:root {
  --bg-primary: #ffffff;
  --text-primary: #333333;
  --accent-color: #007bff;
}

[data-theme="dark"] {
  --bg-primary: #1a1a1a;
  --text-primary: #e0e0e0;
  --accent-color: #4a9eff;
}
```

## Testing

### Component Testing

```typescript
describe("LoginComponent", () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginComponent, HttpClientTestingModule],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should validate email format", () => {
    const emailControl = component.loginForm.get("email");
    emailControl?.setValue("invalid-email");
    expect(emailControl?.hasError("email")).toBeTruthy();
  });
});
```

### Service Testing

```typescript
describe("AuthService", () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should login successfully", () => {
    const mockResponse = { token: "test-token" };

    service
      .login({ email: "test@test.com", password: "test" })
      .subscribe((response) => expect(response.token).toBe("test-token"));

    const req = httpMock.expectOne(`${environment.apiUrl}/authorize`);
    expect(req.request.method).toBe("POST");
    req.flush(mockResponse);
  });
});
```

### Testing Best Practices

- Test component creation
- Test user interactions
- Test form validation
- Test service methods with mocked HTTP
- Test error handling
- Use `HttpClientTestingModule` for HTTP testing
- Clean up subscriptions in tests

## Performance Optimization

### Change Detection

- Use `OnPush` change detection strategy when appropriate
- Prefer signals over observables for better performance
- Avoid unnecessary re-renders
- Use `trackBy` functions in `@for` loops

### Lazy Loading

- Lazy load feature modules when possible
- Use route-based code splitting
- Preload critical routes
- Optimize bundle size

### Memory Management

- Unsubscribe from observables in `ngOnDestroy`
- Use `takeUntil` or `async` pipe for automatic cleanup
- Avoid memory leaks from event listeners
- Clean up timers and intervals

```typescript
export class MyComponent implements OnDestroy {
  private destroy$ = new Subject<void>();

  ngOnInit() {
    this.service
      .getData()
      .pipe(takeUntil(this.destroy$))
      .subscribe((data) => {
        /* handle data */
      });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
```

## Accessibility

### ARIA Attributes

- Use semantic HTML elements first
- Add ARIA labels for screen readers
- Ensure keyboard navigation works
- Test with screen readers
- Maintain proper heading hierarchy

### Accessibility Checklist

- [ ] All interactive elements are keyboard accessible
- [ ] Form inputs have associated labels
- [ ] Images have alt text
- [ ] Color contrast meets WCAG standards
- [ ] Focus indicators are visible
- [ ] Error messages are announced to screen readers

## Development Workflow

### Pre-Development Checklist

1. ✅ Verify existing code builds: `npm run build`
2. ✅ Verify tests pass: `npm test`
3. ✅ Review component architecture

### Implementation Process

1. **Design Phase**

   - Plan component structure
   - Design state management approach
   - Plan routing and navigation
   - Design API integration

2. **Implementation Phase**

   - Create components using Angular CLI
   - Implement services for business logic
   - Add routing and guards
   - Implement forms and validation
   - Add styling and themes

3. **Testing Phase**

   - Write unit tests for components
   - Write unit tests for services
   - Test user interactions
   - Test error scenarios

4. **Quality Assurance**
   - Run `npm run build` - must succeed
   - Run `npm test` - all tests must pass
   - Run `npm run lint` - fix all linting issues
   - Test in multiple browsers
   - Test accessibility

### Build Commands

```bash
# Install dependencies
npm install

# Start development server
npm start
# or
ng serve

# Build for production
npm run build
# or
ng build

# Run tests
npm test
# or
ng test

# Run linting
npm run lint
# or
ng lint
```

## Code Quality Standards

### TypeScript Best Practices

- Enable strict mode in `tsconfig.json`
- Use explicit types (avoid `any`)
- Use interfaces for data models
- Use enums for constants
- Use type guards for type safety

### Linting

- Follow Angular style guide
- Use ESLint for code quality
- Fix all linting errors before committing
- Configure IDE to show linting errors

### Code Review Checklist

- [ ] Components follow standalone pattern
- [ ] Services are singleton (providedIn: 'root')
- [ ] Forms use reactive forms pattern
- [ ] HTTP requests have error handling
- [ ] Signals used for local state
- [ ] Observables cleaned up properly
- [ ] Tests written and passing
- [ ] Accessibility considered
- [ ] Styling follows theme system

## Common Patterns

### Toast Notifications

```typescript
@Injectable({ providedIn: "root" })
export class ToastService {
  private toasts = signal<Toast[]>([]);

  success(message: string): void {
    this.show({ message, type: "success" });
  }

  error(message: string): void {
    this.show({ message, type: "error" });
  }

  private show(toast: Toast): void {
    this.toasts.update((toasts) => [...toasts, toast]);
    setTimeout(() => this.remove(toast), 3000);
  }
}
```

### Keyboard Shortcuts

```typescript
@Injectable({ providedIn: "root" })
export class KeyboardService {
  registerShortcut(key: string, callback: () => void): void {
    // Implementation
  }

  unregisterShortcut(key: string): void {
    // Implementation
  }
}
```

### Loading States

```typescript
export class MyComponent {
  isLoading = signal<boolean>(false);

  loadData(): void {
    this.isLoading.set(true);
    this.service.getData().subscribe({
      next: (data) => {
        // Handle data
        this.isLoading.set(false);
      },
      error: (error) => {
        // Handle error
        this.isLoading.set(false);
      },
    });
  }
}
```

## Common Pitfalls to Avoid

### ❌ Don't

- Use NgModules (use standalone components)
- Mutate signals directly (use `.set()` or `.update()`)
- Forget to unsubscribe from observables
- Use `any` type
- Put business logic in components
- Ignore accessibility
- Skip error handling
- Use inline styles
- Hardcode API URLs

### ✅ Do

- Use standalone components
- Use signals for reactive state
- Clean up subscriptions
- Use explicit types
- Keep components focused on presentation
- Test accessibility
- Handle errors gracefully
- Use component-scoped CSS
- Use environment files for configuration

## Troubleshooting

### Build Failures

- Clear `node_modules` and reinstall: `rm -rf node_modules && npm install`
- Clear Angular cache: `ng cache clean`
- Check TypeScript version compatibility
- Verify all imports are correct

### Runtime Errors

- Check browser console for errors
- Verify API endpoints are correct
- Check authentication token is valid
- Verify CORS configuration on backend

### Test Failures

- Check test setup and teardown
- Verify mocks are configured correctly
- Check for async timing issues
- Use `fakeAsync` and `tick` for async tests

## Additional Resources

- Angular Documentation: https://angular.dev
- RxJS Documentation: https://rxjs.dev
- TypeScript Documentation: https://www.typescriptlang.org
- Frontend README: `frontend/README.md`
- OpenAPI Specification: `api/openapi.yaml`
