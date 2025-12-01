import { ComponentFixture, TestBed } from "@angular/core/testing";
import { DiaryEditorComponent } from "./diary-editor.component";
import { DiaryService } from "../../core/services/diary.service";
import { AuthService } from "../../core/services/auth.service";
import { AssetService } from "../../core/services/asset.service";
import { ToastService } from "../../core/services/toast.service";
import { KeyboardShortcutsService } from "../../core/services/keyboard-shortcuts.service";
import { Router, ActivatedRoute } from "@angular/router";
import { of } from "rxjs";
import { User, DiaryItemsListResponse } from "../../shared/models";

describe("DiaryEditorComponent - XSS Prevention", () => {
  let component: DiaryEditorComponent;

  beforeEach(async () => {
    const diaryServiceSpy = jasmine.createSpyObj("DiaryService", [
      "getItemByDate",
      "saveItem",
    ]);
    const authServiceSpy = jasmine.createSpyObj("AuthService", [
      "getCurrentUser",
      "logout",
    ]);
    const assetServiceSpy = jasmine.createSpyObj("AssetService", [
      "getAssetUrl",
      "uploadAssetsBatch",
    ]);
    const toastServiceSpy = jasmine.createSpyObj("ToastService", [
      "success",
      "error",
    ]);
    const keyboardServiceSpy = jasmine.createSpyObj(
      "KeyboardShortcutsService",
      ["handleKeyboardEvent", "getShortcuts"]
    );
    const routerSpy = jasmine.createSpyObj("Router", [
      "navigate",
      "createUrlTree",
    ]);

    diaryServiceSpy.currentItem$ = of(null);
    keyboardServiceSpy.shortcuts$ = of("");
    keyboardServiceSpy.getShortcuts.and.returnValue([]);

    const activatedRouteSpy = jasmine.createSpyObj("ActivatedRoute", [], {
      snapshot: { params: {} },
    });

    await TestBed.configureTestingModule({
      imports: [DiaryEditorComponent],
      providers: [
        { provide: DiaryService, useValue: diaryServiceSpy },
        { provide: AuthService, useValue: authServiceSpy },
        { provide: AssetService, useValue: assetServiceSpy },
        { provide: ToastService, useValue: toastServiceSpy },
        { provide: KeyboardShortcutsService, useValue: keyboardServiceSpy },
        { provide: Router, useValue: routerSpy },
        { provide: ActivatedRoute, useValue: activatedRouteSpy },
      ],
    }).compileComponents();

    const mockResponse: DiaryItemsListResponse = {
      items: [],
      totalCount: 0,
    };
    diaryServiceSpy.getItemByDate.and.returnValue(of(mockResponse));

    const mockUser: User = {
      id: "test-user-id",
      email: "test@test.com",
      startDate: "2024-01-01",
    };
    authServiceSpy.getCurrentUser.and.returnValue(mockUser);

    component = TestBed.createComponent(DiaryEditorComponent).componentInstance;
  });

  it("should sanitize malicious script tags in markdown", () => {
    const maliciousMarkdown = `# Test\n<script>alert('XSS')</script>`;
    component["bodyText"].set(maliciousMarkdown);

    const sanitizedHtml = component.markdownHtml();
    const htmlString = sanitizedHtml.toString();

    expect(htmlString).not.toContain("<script>");
    expect(htmlString).not.toContain("alert");
  });

  it("should sanitize event handlers in HTML", () => {
    const maliciousMarkdown = `<img src="x" onerror="alert('XSS')">`;
    component["bodyText"].set(maliciousMarkdown);

    const sanitizedHtml = component.markdownHtml();
    const htmlString = sanitizedHtml.toString();

    expect(htmlString).not.toContain("onerror");
  });

  it("should allow safe markdown formatting", () => {
    const safeMarkdown = `# Heading\n**bold** *italic* [link](http://example.com)`;
    component["bodyText"].set(safeMarkdown);

    const sanitizedHtml = component.markdownHtml();
    const htmlString = sanitizedHtml.toString();

    expect(htmlString).toContain("<h1>");
    expect(htmlString).toContain("<strong>");
    expect(htmlString).toContain("<em>");
    expect(htmlString).toContain("<a");
  });

  it("should allow safe image tags", () => {
    const markdownWithImage = `![alt text](http://example.com/image.jpg)`;
    component["bodyText"].set(markdownWithImage);

    const sanitizedHtml = component.markdownHtml();
    const htmlString = sanitizedHtml.toString();

    expect(htmlString).toContain("<img");
    expect(htmlString).toContain("alt");
  });

  it("should remove data attributes", () => {
    const maliciousMarkdown = `<div data-evil="malicious">content</div>`;
    component["bodyText"].set(maliciousMarkdown);

    const sanitizedHtml = component.markdownHtml();
    const htmlString = sanitizedHtml.toString();

    expect(htmlString).not.toContain("data-evil");
  });
});

