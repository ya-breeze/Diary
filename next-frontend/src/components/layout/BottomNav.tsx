'use client';

import { useRouter, usePathname } from 'next/navigation';
import { Home, Search, Plus, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { getTodayString } from '@/lib/utils/date';

export interface BottomNavProps {
  className?: string;
}

export function BottomNav({ className }: BottomNavProps) {
  const router = useRouter();
  const pathname = usePathname();

  const navItems = [
    {
      icon: Home,
      label: 'Home',
      href: '/diary',
      isActive: pathname === '/diary' || pathname?.startsWith('/diary/'),
    },
    {
      icon: Search,
      label: 'Search',
      href: '/search',
      isActive: pathname === '/search',
    },
    {
      icon: Plus,
      label: 'New',
      href: `/diary/${getTodayString()}?edit=true`,
      isActive: false,
      isPrimary: true,
    },
    {
      icon: User,
      label: 'Profile',
      href: '/profile',
      isActive: pathname === '/profile',
    },
  ];

  return (
    <nav
      className={cn(
        'fixed bottom-0 left-0 right-0 z-40 border-t border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900',
        'pb-safe', // Safe area padding for notched devices
        className
      )}
    >
      <div className="flex h-16 items-center justify-around">
        {navItems.map((item) => (
          <button
            key={item.label}
            onClick={() => router.push(item.href)}
            className={cn(
              'flex flex-col items-center justify-center gap-1 px-4 py-2',
              item.isPrimary
                ? 'text-white'
                : item.isActive
                  ? 'text-zinc-900 dark:text-white'
                  : 'text-zinc-500 dark:text-zinc-400'
            )}
          >
            {item.isPrimary ? (
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-zinc-900 dark:bg-white">
                <item.icon className="h-5 w-5 text-white dark:text-zinc-900" />
              </div>
            ) : (
              <>
                <item.icon className="h-5 w-5" />
                <span className="text-xs">{item.label}</span>
              </>
            )}
          </button>
        ))}
      </div>
    </nav>
  );
}
