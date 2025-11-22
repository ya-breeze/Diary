import { Component, inject, ChangeDetectionStrategy } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ToastService } from "../../../core/services/toast.service";

@Component({
  selector: "app-toast-container",
  standalone: true,
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="toast-container">
      @for (toast of toastService.toasts(); track toast.id) {
      <div
        class="toast toast-{{ toast.type }}"
        (click)="toastService.remove(toast.id)"
      >
        <div class="toast-icon">
          @switch (toast.type) { @case ('success') { ✓ } @case ('error') { ✕ }
          @case ('warning') { ⚠ } @case ('info') { ℹ } }
        </div>
        <div class="toast-message">{{ toast.message }}</div>
        <button
          class="toast-close"
          (click)="toastService.remove(toast.id)"
          aria-label="Close notification"
        >
          ×
        </button>
      </div>
      }
    </div>
  `,
  styles: [
    `
      .toast-container {
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 10000;
        display: flex;
        flex-direction: column;
        gap: 10px;
        max-width: 400px;
      }

      .toast {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 16px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        background-color: white;
        cursor: pointer;
        animation: slideIn 0.3s ease-out;
        transition: transform 0.2s, opacity 0.2s;
      }

      .toast:hover {
        transform: translateX(-5px);
      }

      @keyframes slideIn {
        from {
          transform: translateX(100%);
          opacity: 0;
        }
        to {
          transform: translateX(0);
          opacity: 1;
        }
      }

      .toast-icon {
        font-size: 20px;
        font-weight: bold;
        flex-shrink: 0;
      }

      .toast-message {
        flex: 1;
        font-size: 14px;
        line-height: 1.4;
      }

      .toast-close {
        background: none;
        border: none;
        font-size: 24px;
        cursor: pointer;
        padding: 0;
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: inherit;
        opacity: 0.6;
        transition: opacity 0.2s;
      }

      .toast-close:hover {
        opacity: 1;
      }

      .toast-success {
        background-color: #d4edda;
        border-left: 4px solid #28a745;
        color: #155724;
      }

      .toast-error {
        background-color: #f8d7da;
        border-left: 4px solid #dc3545;
        color: #721c24;
      }

      .toast-warning {
        background-color: #fff3cd;
        border-left: 4px solid #ffc107;
        color: #856404;
      }

      .toast-info {
        background-color: #d1ecf1;
        border-left: 4px solid #17a2b8;
        color: #0c5460;
      }

      @media (max-width: 768px) {
        .toast-container {
          top: 10px;
          right: 10px;
          left: 10px;
          max-width: none;
        }
      }
    `,
  ],
})
export class ToastContainerComponent {
  protected readonly toastService = inject(ToastService);
}
