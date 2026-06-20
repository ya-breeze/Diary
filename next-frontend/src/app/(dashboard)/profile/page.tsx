'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { LogOut, BookOpen, Flame, Tag, Users, Sparkles } from 'lucide-react';
import { Badge } from '@/components/ui';
import { useAuthStore } from '@/store';
import { useDiaryEntries } from '@/hooks';
import { authApi } from '@/lib/api/auth';
import type { Family } from '@/types';

function computeStreak(dates: string[]): number {
  if (dates.length === 0) return 0;

  const sorted = [...dates].sort().reverse();
  const today = new Date().toISOString().slice(0, 10);

  // Streak must start from today or yesterday (allow for timezone flexibility)
  const msPerDay = 86400000;
  const todayMs = new Date(today).getTime();
  const mostRecentMs = new Date(sorted[0]).getTime();
  if (todayMs - mostRecentMs > msPerDay) return 0;

  let streak = 1;
  for (let i = 1; i < sorted.length; i++) {
    const prev = new Date(sorted[i - 1]).getTime();
    const curr = new Date(sorted[i]).getTime();
    if (prev - curr === msPerDay) {
      streak++;
    } else {
      break;
    }
  }
  return streak;
}

function computeTopTags(entries: { tags: string[] | null }[], limit = 5): string[] {
  const counts: Record<string, number> = {};
  for (const entry of entries) {
    for (const tag of (entry.tags ?? [])) {
      counts[tag] = (counts[tag] ?? 0) + 1;
    }
  }
  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, limit)
    .map(([tag]) => tag);
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'long' });
}

export default function ProfilePage() {
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { data, isLoading } = useDiaryEntries();
  const [family, setFamily] = useState<Family | null>(null);
  const [savingAi, setSavingAi] = useState(false);

  useEffect(() => {
    authApi.getFamily().then(setFamily).catch(() => {});
  }, []);

  const toggleAiTagging = async () => {
    if (!family) return;
    setSavingAi(true);
    try {
      const updated = await authApi.updateFamilySettings({
        aiTaggingEnabled: !family.aiTaggingEnabled,
      });
      setFamily(updated);
    } catch (error) {
      console.error('Failed to update AI tagging setting:', error);
    } finally {
      setSavingAi(false);
    }
  };

  const entries = data?.items ?? [];
  const totalCount = data?.totalCount ?? 0;
  const uniqueTags = new Set(entries.flatMap((e) => e.tags ?? [])).size;
  const streak = computeStreak(entries.map((e) => e.date));
  const topTags = computeTopTags(entries);

  const initials = user?.email
    ? user.email.slice(0, 2).toUpperCase()
    : '??';

  const handleLogout = async () => {
    await logout();
    router.push('/login');
  };

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      {/* Header */}
      <div className="border-b border-zinc-200 bg-white px-6 py-8 dark:border-zinc-800 dark:bg-zinc-900">
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-zinc-900 text-lg font-semibold text-white dark:bg-zinc-100 dark:text-zinc-900">
            {initials}
          </div>
          <div>
            <p className="font-medium text-zinc-900 dark:text-zinc-100">{user?.email}</p>
            {user?.startDate && (
              <p className="text-sm text-zinc-500 dark:text-zinc-400">
                Member since {formatDate(user.startDate)}
              </p>
            )}
          </div>
        </div>
      </div>

      <div className="space-y-6 p-6">
        {/* Stat cards */}
        <div className="grid grid-cols-3 gap-3">
          {[
            { icon: BookOpen, label: 'Entries', value: isLoading ? '—' : String(totalCount) },
            { icon: Flame, label: 'Streak', value: isLoading ? '—' : `${streak}d` },
            { icon: Tag, label: 'Tags', value: isLoading ? '—' : String(uniqueTags) },
          ].map(({ icon: Icon, label, value }) => (
            <div
              key={label}
              className="flex flex-col items-center gap-1 rounded-xl border border-zinc-200 bg-white py-4 dark:border-zinc-800 dark:bg-zinc-900"
            >
              <Icon className="h-4 w-4 text-zinc-400" />
              <span className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{value}</span>
              <span className="text-xs text-zinc-500 dark:text-zinc-400">{label}</span>
            </div>
          ))}
        </div>

        {/* Family */}
        {family && (
          <div className="rounded-xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
            <div className="mb-3 flex items-center gap-2">
              <Users className="h-4 w-4 text-zinc-400" />
              <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400">{family.name}</p>
            </div>
            <div className="space-y-1">
              {family.members.map((member) => (
                <p key={member.email} className="text-sm text-zinc-700 dark:text-zinc-300">
                  {member.email}
                </p>
              ))}
            </div>
          </div>
        )}

        {/* AI settings */}
        {family && (
          <div className="rounded-xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
            <div className="mb-3 flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-zinc-400" />
              <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400">AI tagging</p>
            </div>
            <label className="flex items-center justify-between gap-4">
              <span className="text-sm text-zinc-700 dark:text-zinc-300">
                Suggest tags for entries from their text
              </span>
              <input
                type="checkbox"
                role="switch"
                checked={!!family.aiTaggingEnabled}
                disabled={savingAi}
                onChange={toggleAiTagging}
                className="h-5 w-5 cursor-pointer accent-blue-600 disabled:opacity-50"
                data-testid="ai-tagging-toggle"
              />
            </label>
          </div>
        )}

        {/* Top tags */}
        {topTags.length > 0 && (
          <div className="rounded-xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
            <p className="mb-3 text-sm font-medium text-zinc-500 dark:text-zinc-400">Top tags</p>
            <div className="flex flex-wrap gap-2">
              {topTags.map((tag) => (
                <Badge key={tag} variant="default">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Logout */}
        <button
          onClick={handleLogout}
          className="flex w-full items-center justify-center gap-2 rounded-xl border border-zinc-200 bg-white px-4 py-3 text-sm font-medium text-red-600 hover:bg-red-50 dark:border-zinc-800 dark:bg-zinc-900 dark:text-red-400 dark:hover:bg-red-950/30"
        >
          <LogOut className="h-4 w-4" />
          Log out
        </button>
      </div>
    </div>
  );
}
