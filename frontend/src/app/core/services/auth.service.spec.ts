import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { Router } from "@angular/router";
import { AuthService } from "./auth.service";
import { User, AuthData, AuthResponse } from "../../shared/models";
import { environment } from "../../../environments/environment";

describe("AuthService", () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let routerSpy: jasmine.SpyObj<Router>;
  let localStorageSpy: jasmine.SpyObj<Storage>;

  const mockUser: User = {
    id: "test-uuid-123",
    email: "test@example.com",
    startDate: "2024-01-01",
  };

  const mockAuthResponse: AuthResponse = {
    token: "mock-jwt-token",
  };

  beforeEach(() => {
    routerSpy = jasmine.createSpyObj("Router", ["navigate"]);
    localStorageSpy = jasmine.createSpyObj("localStorage", [
      "getItem",
      "setItem",
      "removeItem",
    ]);

    Object.defineProperty(window, "localStorage", {
      value: localStorageSpy,
      writable: true,
    });

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        AuthService,
        { provide: Router, useValue: routerSpy },
      ],
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should set isAuthenticated to true if token exists on initialization", () => {
    // Need to set up localStorage before creating the service
    localStorageSpy.getItem.and.returnValue("existing-token");

    // Create a new TestBed with fresh service instance
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AuthService, { provide: Router, useValue: routerSpy }],
    });

    const newService = TestBed.inject(AuthService);

    expect(newService.isAuthenticated()).toBe(true);
  });

  it("should login successfully and store token", (done) => {
    const credentials: AuthData = {
      email: "test@example.com",
      password: "password123",
    };

    service.login(credentials).subscribe((response) => {
      expect(response).toEqual(mockAuthResponse);
      expect(localStorageSpy.setItem).toHaveBeenCalledWith(
        "diary_auth_token",
        "mock-jwt-token"
      );
      expect(service.isAuthenticated()).toBe(true);
      done();
    });

    const loginReq = httpMock.expectOne(`${environment.apiUrl}/authorize`);
    expect(loginReq.request.method).toBe("POST");
    expect(loginReq.request.body).toEqual(credentials);
    loginReq.flush(mockAuthResponse);

    const userReq = httpMock.expectOne(`${environment.apiUrl}/user`);
    expect(userReq.request.method).toBe("GET");
    userReq.flush(mockUser);
  });

  it("should logout and clear token", () => {
    service.logout();

    const req = httpMock.expectOne(`${environment.apiUrl}/logout`);
    expect(req.request.method).toBe("POST");
    req.flush({});

    expect(localStorageSpy.removeItem).toHaveBeenCalledWith("diary_auth_token");
    expect(service.getCurrentUser()).toBeNull();
    expect(service.isAuthenticated()).toBeFalse();
    expect(routerSpy.navigate).toHaveBeenCalledWith(["/login"]);
  });

  it("should get token from localStorage", () => {
    localStorageSpy.getItem.and.returnValue("test-token");

    const token = service.getToken();

    expect(token).toBe("test-token");
    expect(localStorageSpy.getItem).toHaveBeenCalledWith("diary_auth_token");
  });

  it("should return null when no token exists", () => {
    localStorageSpy.getItem.and.returnValue(null);

    const token = service.getToken();

    expect(token).toBeNull();
  });

  it("should validate authentication with valid token and user", (done) => {
    localStorageSpy.getItem.and.returnValue("valid-token");

    service.validateAuthentication().subscribe((isValid) => {
      expect(isValid).toBe(true);
      expect(service.isAuthenticated()).toBe(true);
      expect(service.getCurrentUser()).toEqual(mockUser);
      done();
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/user`);
    expect(req.request.method).toBe("GET");
    req.flush(mockUser);
  });

  it("should return false when validating with no token", (done) => {
    localStorageSpy.getItem.and.returnValue(null);

    service.validateAuthentication().subscribe((isValid) => {
      expect(isValid).toBe(false);
      expect(service.isAuthenticated()).toBe(false);
      done();
    });

    httpMock.expectNone(`${environment.apiUrl}/user`);
  });

  it("should handle validation failure and clear token", (done) => {
    localStorageSpy.getItem.and.returnValue("invalid-token");

    service.validateAuthentication().subscribe((isValid) => {
      expect(isValid).toBe(false);
      expect(service.isAuthenticated()).toBe(false);
      expect(localStorageSpy.removeItem).toHaveBeenCalledWith(
        "diary_auth_token"
      );
      done();
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/user`);
    req.flush("Unauthorized", { status: 401, statusText: "Unauthorized" });
  });

  it("should return current user", (done) => {
    const credentials: AuthData = {
      email: "test@example.com",
      password: "password123",
    };

    service.login(credentials).subscribe(() => {
      // Need to wait for the user profile to load
      setTimeout(() => {
        const user = service.getCurrentUser();
        expect(user).toEqual(mockUser);
        done();
      }, 0);
    });

    httpMock
      .expectOne(`${environment.apiUrl}/authorize`)
      .flush(mockAuthResponse);
    httpMock.expectOne(`${environment.apiUrl}/user`).flush(mockUser);
  });
});
