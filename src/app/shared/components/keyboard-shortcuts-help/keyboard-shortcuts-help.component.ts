import { Component, EventEmitter, Input, Output } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  KeyboardShortcutsService,
  KeyboardShortcut,
} from "../../../core/services/keyboard-shortcuts.service";

@Component({
  selector: "app-keyboard-shortcuts-help",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./keyboard-shortcuts-help.component.html",
  styleUrl: "./keyboard-shortcuts-help.component.css",
})
export class KeyboardShortcutsHelpComponent {
  @Input() isOpen = false;
  @Output() close = new EventEmitter<void>();

  shortcuts: KeyboardShortcut[] = [];

  constructor(private keyboardShortcutsService: KeyboardShortcutsService) {
    this.shortcuts = this.keyboardShortcutsService.getShortcuts();
  }

  onClose(): void {
    this.close.emit();
  }

  getShortcutLabel(shortcut: KeyboardShortcut): string {
    return this.keyboardShortcutsService.getShortcutLabel(shortcut);
  }

  onBackdropClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      this.onClose();
    }
  }
}

