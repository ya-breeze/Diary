import { Injectable } from "@angular/core";
import { Subject } from "rxjs";

export interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  altKey?: boolean;
  shiftKey?: boolean;
  metaKey?: boolean;
  description: string;
  action: string;
}

@Injectable({
  providedIn: "root",
})
export class KeyboardShortcutsService {
  private shortcuts = new Subject<string>();
  public shortcuts$ = this.shortcuts.asObservable();

  private registeredShortcuts: KeyboardShortcut[] = [
    {
      key: "s",
      ctrlKey: true,
      description: "Save diary entry",
      action: "save",
    },
    {
      key: "ArrowLeft",
      altKey: true,
      description: "Go to previous entry",
      action: "previous",
    },
    {
      key: "ArrowRight",
      altKey: true,
      description: "Go to next entry",
      action: "next",
    },
    {
      key: "p",
      ctrlKey: true,
      description: "Toggle markdown preview",
      action: "preview",
    },
    {
      key: "f",
      ctrlKey: true,
      description: "Focus search",
      action: "search",
    },
    {
      key: "/",
      description: "Show keyboard shortcuts help",
      action: "help",
    },
    {
      key: "Escape",
      description: "Close modal or cancel action",
      action: "escape",
    },
  ];

  constructor() {}

  handleKeyboardEvent(event: KeyboardEvent): void {
    const matchedShortcut = this.registeredShortcuts.find((shortcut) => {
      return (
        shortcut.key.toLowerCase() === event.key.toLowerCase() &&
        (shortcut.ctrlKey === undefined || shortcut.ctrlKey === event.ctrlKey) &&
        (shortcut.altKey === undefined || shortcut.altKey === event.altKey) &&
        (shortcut.shiftKey === undefined || shortcut.shiftKey === event.shiftKey) &&
        (shortcut.metaKey === undefined || shortcut.metaKey === (event.metaKey || event.ctrlKey))
      );
    });

    if (matchedShortcut) {
      // Don't prevent default for help shortcut
      if (matchedShortcut.action !== "help") {
        event.preventDefault();
      }
      this.shortcuts.next(matchedShortcut.action);
    }
  }

  getShortcuts(): KeyboardShortcut[] {
    return this.registeredShortcuts;
  }

  getShortcutLabel(shortcut: KeyboardShortcut): string {
    const keys: string[] = [];
    
    if (shortcut.ctrlKey || shortcut.metaKey) {
      keys.push("Ctrl");
    }
    if (shortcut.altKey) {
      keys.push("Alt");
    }
    if (shortcut.shiftKey) {
      keys.push("Shift");
    }
    
    // Format special keys
    let keyLabel = shortcut.key;
    if (shortcut.key === "ArrowLeft") keyLabel = "←";
    if (shortcut.key === "ArrowRight") keyLabel = "→";
    if (shortcut.key === "ArrowUp") keyLabel = "↑";
    if (shortcut.key === "ArrowDown") keyLabel = "↓";
    if (shortcut.key === "Escape") keyLabel = "Esc";
    
    keys.push(keyLabel.toUpperCase());
    
    return keys.join(" + ");
  }
}

