'use client';

import { create } from 'zustand';

interface UIState {
  sidebarOpen: boolean;
  editMode: boolean;
  previewAsset: string | null;

  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  setEditMode: (mode: boolean) => void;
  toggleEditMode: () => void;
  setPreviewAsset: (path: string | null) => void;
}

export const useUIStore = create<UIState>()((set) => ({
  sidebarOpen: false,
  editMode: false,
  previewAsset: null,

  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setEditMode: (mode) => set({ editMode: mode }),
  toggleEditMode: () => set((state) => ({ editMode: !state.editMode })),
  setPreviewAsset: (path) => set({ previewAsset: path }),
}));
