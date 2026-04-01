'use client';

import { useState } from 'react';
import { AlertTriangle, CheckCircle, Wrench } from 'lucide-react';
import { Drawer } from '@/components/ui/Drawer';
import { Button } from '@/components/ui/Button';
import { useHealthIssues, useFixHealthIssues } from '@/hooks';
import type { HealthIssue } from '@/lib/api/health';

interface HealthPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function HealthPanel({ isOpen, onClose }: HealthPanelProps) {
  const { data, isLoading } = useHealthIssues();
  const fixMutation = useFixHealthIssues();
  const [fixing, setFixing] = useState<string | null>(null);

  const issues = data?.issues ?? [];
  const lastChecked = data?.lastChecked ? new Date(data.lastChecked) : null;

  // Group issues by check type
  const grouped = issues.reduce<Record<string, HealthIssue[]>>((acc, issue) => {
    acc[issue.check] = [...(acc[issue.check] ?? []), issue];
    return acc;
  }, {});

  const handleFix = async (checkName: string) => {
    setFixing(checkName);
    try {
      await fixMutation.mutateAsync([checkName]);
    } finally {
      setFixing(null);
    }
  };

  return (
    <Drawer isOpen={isOpen} onClose={onClose} side="right" className="w-[360px]">
      <div className="flex h-full flex-col p-5 pt-12">
        <h2 className="mb-1 text-lg font-semibold text-zinc-900 dark:text-zinc-100">
          Storage Health
        </h2>

        {lastChecked && (
          <p className="mb-4 text-xs text-zinc-500">
            Last checked: {lastChecked.toLocaleString()}
          </p>
        )}

        {isLoading && (
          <p className="text-sm text-zinc-500">Loading…</p>
        )}

        {!isLoading && issues.length === 0 && (
          <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
            <CheckCircle className="h-5 w-5" />
            <span className="text-sm">No issues found.</span>
          </div>
        )}

        {!isLoading && issues.length > 0 && (
          <div className="flex flex-col gap-4 overflow-y-auto">
            {Object.entries(grouped).map(([checkName, groupIssues]) => {
              const fixable = groupIssues.some((i) => i.fixable);
              return (
                <div key={checkName} className="rounded-lg border border-zinc-200 p-3 dark:border-zinc-700">
                  <div className="mb-2 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <AlertTriangle className="h-4 w-4 text-amber-500" />
                      <span className="font-medium text-zinc-800 dark:text-zinc-200">
                        {checkName}
                      </span>
                      <span className="text-xs text-zinc-500">
                        ({groupIssues.length} {groupIssues.length === 1 ? 'issue' : 'issues'})
                      </span>
                    </div>
                    {fixable && (
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => handleFix(checkName)}
                        disabled={fixing === checkName}
                      >
                        <Wrench className="mr-1 h-3 w-3" />
                        {fixing === checkName ? 'Fixing…' : 'Fix'}
                      </Button>
                    )}
                  </div>
                  <ul className="space-y-1">
                    {groupIssues.map((issue, idx) => (
                      <li key={idx} className="text-xs text-zinc-600 dark:text-zinc-400">
                        • {issue.message}
                        {issue.path && (
                          <span className="ml-1 font-mono text-zinc-400 dark:text-zinc-500">
                            ({issue.path.split('/').pop()})
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </Drawer>
  );
}
