'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Sidebar, MobileHeader, BottomNav } from '@/components/layout';
import { Drawer } from '@/components/ui';
import { useAuthStore, useUIStore } from '@/store';
import { useIsDesktop, useIsMobile } from '@/hooks';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { isAuthenticated, validateSession } = useAuthStore();
  const { sidebarOpen, setSidebarOpen, toggleSidebar } = useUIStore();
  const isDesktop = useIsDesktop();
  const isMobile = useIsMobile();

  // Check authentication
  useEffect(() => {
    const checkAuth = async () => {
      const valid = await validateSession();
      if (!valid) {
        router.push('/login');
      }
    };
    checkAuth();
  }, [validateSession, router]);

  // Close sidebar on desktop
  useEffect(() => {
    if (isDesktop) {
      setSidebarOpen(false);
    }
  }, [isDesktop, setSidebarOpen]);

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-900" />
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-white dark:bg-zinc-950">
      {/* Desktop Sidebar - always visible */}
      {isDesktop && (
        <Sidebar className="hidden w-[300px] flex-shrink-0 flex-col border-r border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900 lg:flex" />
      )}

      {/* Tablet/Mobile Sidebar - drawer */}
      {!isDesktop && (
        <Drawer isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)}>
          <Sidebar
            className="flex h-full flex-col"
            onSelectEntry={() => setSidebarOpen(false)}
          />
        </Drawer>
      )}

      {/* Main content area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Mobile/Tablet header */}
        {!isDesktop && (
          <MobileHeader onMenuClick={toggleSidebar} />
        )}

        {/* Main content */}
        <main className="flex-1 overflow-y-auto">
          <div className={isMobile ? 'pb-20' : ''}>{children}</div>
        </main>

        {/* Mobile bottom nav */}
        {isMobile && <BottomNav />}
      </div>
    </div>
  );
}
