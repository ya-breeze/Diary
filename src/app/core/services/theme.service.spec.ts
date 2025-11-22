import { TestBed, fakeAsync, tick, flush } from "@angular/core/testing";
import { ThemeService } from "./theme.service";

describe("ThemeService", () => {
  let service: ThemeService;
  let localStorageSpy: jasmine.SpyObj<Storage>;

  beforeEach(() => {
    // Mock localStorage
    localStorageSpy = jasmine.createSpyObj("localStorage", [
      "getItem",
      "setItem",
      "removeItem",
    ]);
    Object.defineProperty(window, "localStorage", {
      value: localStorageSpy,
      writable: true,
    });

    // Mock matchMedia
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: jasmine.createSpy("matchMedia").and.returnValue({
        matches: false,
        media: "(prefers-color-scheme: dark)",
        addEventListener: jasmine.createSpy("addEventListener"),
        removeEventListener: jasmine.createSpy("removeEventListener"),
      }),
    });

    TestBed.configureTestingModule({});
  });

  afterEach(() => {
    // Clean up document attribute
    document.documentElement.removeAttribute("data-theme");
  });

  it("should be created", () => {
    service = TestBed.inject(ThemeService);
    expect(service).toBeTruthy();
  });

  it("should initialize with light theme by default", () => {
    localStorageSpy.getItem.and.returnValue(null);
    service = TestBed.inject(ThemeService);

    expect(service.currentTheme()).toBe("light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  it("should initialize with saved theme from localStorage", () => {
    localStorageSpy.getItem.and.returnValue("dark");
    service = TestBed.inject(ThemeService);

    expect(service.currentTheme()).toBe("dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
  });

  it("should initialize with system preference when no saved theme", () => {
    localStorageSpy.getItem.and.returnValue(null);
    (window.matchMedia as jasmine.Spy).and.returnValue({
      matches: true,
      media: "(prefers-color-scheme: dark)",
      addEventListener: jasmine.createSpy("addEventListener"),
      removeEventListener: jasmine.createSpy("removeEventListener"),
    });

    service = TestBed.inject(ThemeService);

    expect(service.currentTheme()).toBe("dark");
  });

  it("should toggle theme from light to dark", fakeAsync(() => {
    localStorageSpy.getItem.and.returnValue("light");
    service = TestBed.inject(ThemeService);
    flush(); // Flush initial effect

    service.toggleTheme();
    flush(); // Flush effect after toggle

    expect(service.currentTheme()).toBe("dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    expect(localStorageSpy.setItem).toHaveBeenCalledWith("diary-theme", "dark");
  }));

  it("should toggle theme from dark to light", fakeAsync(() => {
    localStorageSpy.getItem.and.returnValue("dark");
    service = TestBed.inject(ThemeService);
    flush(); // Flush initial effect

    service.toggleTheme();
    flush(); // Flush effect after toggle

    expect(service.currentTheme()).toBe("light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
    expect(localStorageSpy.setItem).toHaveBeenCalledWith(
      "diary-theme",
      "light"
    );
  }));

  it("should set theme directly", fakeAsync(() => {
    localStorageSpy.getItem.and.returnValue("light");
    service = TestBed.inject(ThemeService);
    flush(); // Flush initial effect

    service.setTheme("dark");
    flush(); // Flush effect after setTheme

    expect(service.currentTheme()).toBe("dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    expect(localStorageSpy.setItem).toHaveBeenCalledWith("diary-theme", "dark");
  }));

  it("should persist theme changes to localStorage", fakeAsync(() => {
    localStorageSpy.getItem.and.returnValue("light");
    service = TestBed.inject(ThemeService);
    flush(); // Flush initial effect

    service.setTheme("dark");
    flush(); // Flush effect after setTheme

    expect(localStorageSpy.setItem).toHaveBeenCalledWith("diary-theme", "dark");
  }));

  it("should apply theme to document root element", fakeAsync(() => {
    localStorageSpy.getItem.and.returnValue("light");
    service = TestBed.inject(ThemeService);
    flush(); // Flush initial effect

    expect(document.documentElement.getAttribute("data-theme")).toBe("light");

    service.setTheme("dark");
    flush(); // Flush effect after setTheme

    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
  }));
});
