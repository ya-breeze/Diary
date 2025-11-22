import { TestBed } from "@angular/core/testing";
import {
  KeyboardShortcutsService,
  KeyboardShortcut,
} from "./keyboard-shortcuts.service";

describe("KeyboardShortcutsService", () => {
  let service: KeyboardShortcutsService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(KeyboardShortcutsService);
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should return registered shortcuts", () => {
    const shortcuts = service.getShortcuts();

    expect(shortcuts.length).toBeGreaterThan(0);
    expect(shortcuts.some((s) => s.action === "save")).toBe(true);
    expect(shortcuts.some((s) => s.action === "preview")).toBe(true);
    expect(shortcuts.some((s) => s.action === "help")).toBe(true);
  });

  it("should emit action when Ctrl+S is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("save");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "s",
      ctrlKey: true,
    });

    service.handleKeyboardEvent(event);
  });

  it("should emit action when Ctrl+P is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("preview");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "p",
      ctrlKey: true,
    });

    service.handleKeyboardEvent(event);
  });

  it("should emit action when Alt+ArrowLeft is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("previous");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "ArrowLeft",
      altKey: true,
    });

    service.handleKeyboardEvent(event);
  });

  it("should emit action when Alt+ArrowRight is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("next");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "ArrowRight",
      altKey: true,
    });

    service.handleKeyboardEvent(event);
  });

  it("should emit action when / is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("help");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "/",
    });

    service.handleKeyboardEvent(event);
  });

  it("should emit action when Escape is pressed", (done) => {
    service.shortcuts$.subscribe((action) => {
      expect(action).toBe("escape");
      done();
    });

    const event = new KeyboardEvent("keydown", {
      key: "Escape",
    });

    service.handleKeyboardEvent(event);
  });

  it("should not emit action for unregistered shortcuts", () => {
    let emitted = false;
    service.shortcuts$.subscribe(() => {
      emitted = true;
    });

    const event = new KeyboardEvent("keydown", {
      key: "x",
      ctrlKey: true,
    });

    service.handleKeyboardEvent(event);

    expect(emitted).toBe(false);
  });

  it("should generate correct label for Ctrl+S", () => {
    const shortcut: KeyboardShortcut = {
      key: "s",
      ctrlKey: true,
      description: "Save",
      action: "save",
    };

    const label = service.getShortcutLabel(shortcut);

    expect(label).toBe("Ctrl + S");
  });

  it("should generate correct label for Alt+ArrowLeft", () => {
    const shortcut: KeyboardShortcut = {
      key: "ArrowLeft",
      altKey: true,
      description: "Previous",
      action: "previous",
    };

    const label = service.getShortcutLabel(shortcut);

    expect(label).toBe("Alt + ←");
  });

  it("should generate correct label for Escape", () => {
    const shortcut: KeyboardShortcut = {
      key: "Escape",
      description: "Close",
      action: "escape",
    };

    const label = service.getShortcutLabel(shortcut);

    expect(label).toBe("ESC");
  });
});
