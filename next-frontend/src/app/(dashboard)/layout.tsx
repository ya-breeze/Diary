'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ShieldAlert } from 'lucide-react';
import { Sidebar, MobileHeader, BottomNav } from '@/components/layout';
import { Drawer } from '@/components/ui';
import { HealthPanel } from '@/components/health/HealthPanel';
import { useAuthStore, useUIStore } from '@/store';
import { useIsDesktop, useIsMobile, useHealthIssues } from '@/hooks';

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
  const [healthOpen, setHealthOpen] = useState(false);
  const { data: healthData } = useHealthIssues();
  const issueCount = healthData?.issues?.length ?? 0;

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
          <MobileHeader
            onMenuClick={toggleSidebar}
            rightContent={
              issueCount > 0 ? (
                <button
                  onClick={() => setHealthOpen(true)}
                  className="relative flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-amber-600 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-900/20"
                  title={`${issueCount} storage issue${issueCount > 1 ? 's' : ''}`}
                >
                  <ShieldAlert className="h-4 w-4" />
                  <span>{issueCount}</span>
                </button>
              ) : undefined
            }
          />
        )}

        {/* Main content */}
        <main className="flex-1 overflow-y-auto">
          <div className={isMobile ? 'pb-20' : ''}>{children}</div>
        </main>

        {/* Mobile bottom nav */}
        {isMobile && <BottomNav />}
      </div>

      <HealthPanel isOpen={healthOpen} onClose={() => setHealthOpen(false)} />
    </div>
  );
}
