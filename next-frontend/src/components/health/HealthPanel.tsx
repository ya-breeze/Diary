'use client';

import { useState } from 'react';
import {
  AlertTriangle,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  Eye,
  EyeOff,
  Link,
  Trash2,
  Wrench,
} from 'lucide-react';
import { Drawer } from '@/components/ui/Drawer';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import {
  useHealthIssues,
  useFixHealthIssues,
  useDeleteOrphan,
  useAttachOrphan,
  useIgnoreOrphan,
  useUnignoreOrphan,
} from '@/hooks';
import { assetsApi } from '@/lib/api';
import type { HealthIssue } from '@/lib/api/health';

interface HealthPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

interface OrphanState {
  confirmDelete?: boolean;
  showAttach?: boolean;
  attachDate?: string;
}

const IMAGE_EXTENSIONS = /\.(jpe?g|png|gif|webp|avif|svg)$/i;

export function HealthPanel({ isOpen, onClose }: HealthPanelProps) {
  const { data, isLoading } = useHealthIssues();
  const fixMutation = useFixHealthIssues();
  const deleteOrphan = useDeleteOrphan();
  const attachOrphan = useAttachOrphan();
  const ignoreOrphan = useIgnoreOrphan();
  const unignoreOrphan = useUnignoreOrphan();

  const [fixing, setFixing] = useState<string | null>(null);
  const [orphanStates, setOrphanStates] = useState<Record<string, OrphanState>>({});
  const [ignoredExpanded, setIgnoredExpanded] = useState(false);

  const issues = data?.issues ?? [];
  const ignoredOrphans = data?.ignoredOrphans ?? [];
  const lastChecked = data?.lastChecked ? new Date(data.lastChecked) : null;

  const orphanIssues = issues.filter((i) => i.check === 'orphans');
  const grouped = issues.reduce<Record<string, HealthIssue[]>>((acc, issue) => {
    if (issue.check !== 'orphans') {
      acc[issue.check] = [...(acc[issue.check] ?? []), issue];
    }
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

  const setOrphanState = (filename: string, patch: Partial<OrphanState>) =>
    setOrphanStates((prev) => ({ ...prev, [filename]: { ...prev[filename], ...patch } }));

  const clearOrphanState = (filename: string) =>
    setOrphanStates((prev) => { const s = { ...prev }; delete s[filename]; return s; });

  const handleDelete = async (filename: string) => {
    if (!orphanStates[filename]?.confirmDelete) {
      setOrphanState(filename, { confirmDelete: true });
      return;
    }
    await deleteOrphan.mutateAsync(filename);
    clearOrphanState(filename);
  };

  const handleAttachSubmit = async (filename: string) => {
    const date = orphanStates[filename]?.attachDate ?? '';
    if (!date) return;
    await attachOrphan.mutateAsync({ filename, date });
    clearOrphanState(filename);
  };

  const handleIgnore = async (filename: string) => {
    await ignoreOrphan.mutateAsync(filename);
    clearOrphanState(filename);
  };

  const filenameFromPath = (path: string) => path.split('/').pop() ?? path;

  const isAnyOrphanBusy = deleteOrphan.isPending || attachOrphan.isPending || ignoreOrphan.isPending;

  return (
    <Drawer isOpen={isOpen} onClose={onClose} side="right" className="w-[400px]">
      <div className="flex h-full flex-col p-5 pt-12">
        <h2 className="mb-1 text-lg font-semibold text-zinc-900 dark:text-zinc-100">
          Storage Health
        </h2>

        {lastChecked && (
          <p className="mb-4 text-xs text-zinc-500">
            Last checked: {lastChecked.toLocaleString()}
          </p>
        )}

        {isLoading && <p className="text-sm text-zinc-500">Loading…</p>}

        {!isLoading && issues.length === 0 && ignoredOrphans.length === 0 && (
          <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
            <CheckCircle className="h-5 w-5" />
            <span className="text-sm">No issues found.</span>
          </div>
        )}

        <div className="flex flex-col gap-4 overflow-y-auto">
          {/* Refs / mime check groups */}
          {!isLoading &&
            Object.entries(grouped).map(([checkName, groupIssues]) => {
              const fixable = groupIssues.some((i) => i.fixable);
              return (
                <div
                  key={checkName}
                  className="rounded-lg border border-zinc-200 p-3 dark:border-zinc-700"
                >
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

          {/* Orphans — per-item */}
          {!isLoading && orphanIssues.length > 0 && (
            <div className="rounded-lg border border-zinc-200 p-3 dark:border-zinc-700">
              <div className="mb-3 flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-amber-500" />
                <span className="font-medium text-zinc-800 dark:text-zinc-200">orphans</span>
                <span className="text-xs text-zinc-500">
                  ({orphanIssues.length} {orphanIssues.length === 1 ? 'issue' : 'issues'})
                </span>
              </div>

              <div className="flex flex-col gap-3">
                {orphanIssues.map((issue) => {
                  const filename = filenameFromPath(issue.path);
                  const state = orphanStates[filename] ?? {};
                  const imgUrl = IMAGE_EXTENSIONS.test(filename)
                    ? assetsApi.getAssetUrl(filename)
                    : null;

                  return (
                    <div
                      key={issue.path}
                      className="rounded-md border border-zinc-100 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-800/50"
                    >
                      {/* Thumbnail + filename */}
                      <div className="mb-2 flex items-center gap-2">
                        {imgUrl ? (
                          <div className="group relative flex-shrink-0">
                            <img
                              src={imgUrl}
                              alt={filename}
                              className="h-10 w-10 rounded object-cover"
                              onError={(e) => {
                                (e.target as HTMLImageElement).closest('div')!.style.display = 'none';
                              }}
                            />
                            <div className="pointer-events-none absolute bottom-0 right-full z-50 mr-2 hidden w-64 group-hover:block">
                              <img
                                src={imgUrl}
                                alt={filename}
                                className="max-h-64 w-full rounded-md object-contain shadow-xl ring-1 ring-zinc-200 dark:ring-zinc-700"
                              />
                            </div>
                          </div>
                        ) : (
                          <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded bg-zinc-200 dark:bg-zinc-700">
                            <span className="text-[10px] uppercase text-zinc-400">
                              {filename.split('.').pop()?.slice(0, 4) ?? 'file'}
                            </span>
                          </div>
                        )}
                        <span className="flex-1 truncate font-mono text-xs text-zinc-600 dark:text-zinc-300">
                          {filename}
                        </span>
                      </div>

                      {/* Confirm-delete warning */}
                      {state.confirmDelete && (
                        <p className="mb-2 text-xs text-red-600 dark:text-red-400">
                          Permanently delete this file?
                        </p>
                      )}

                      {/* Attach-to-day row */}
                      {state.showAttach && (
                        <div className="mb-2 flex items-center gap-1">
                          <Input
                            type="date"
                            className="h-7 flex-1 text-xs"
                            value={state.attachDate ?? ''}
                            onChange={(e) =>
                              setOrphanState(filename, { attachDate: e.target.value })
                            }
                          />
                          <Button
                            size="sm"
                            variant="primary"
                            onClick={() => handleAttachSubmit(filename)}
                            disabled={!state.attachDate || isAnyOrphanBusy}
                          >
                            Add
                          </Button>
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() =>
                              setOrphanState(filename, {
                                showAttach: false,
                                attachDate: undefined,
                              })
                            }
                          >
                            ✕
                          </Button>
                        </div>
                      )}

                      {/* Action buttons */}
                      {!state.showAttach && (
                        <div className="flex flex-wrap gap-1">
                          {state.confirmDelete ? (
                            <>
                              <Button
                                size="sm"
                                variant="danger"
                                onClick={() => handleDelete(filename)}
                                disabled={isAnyOrphanBusy}
                              >
                                <Trash2 className="mr-1 h-3 w-3" />
                                Confirm delete
                              </Button>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => setOrphanState(filename, { confirmDelete: false })}
                              >
                                Cancel
                              </Button>
                            </>
                          ) : (
                            <>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => handleDelete(filename)}
                                disabled={isAnyOrphanBusy}
                              >
                                <Trash2 className="mr-1 h-3 w-3" />
                                Delete
                              </Button>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() =>
                                  setOrphanState(filename, {
                                    showAttach: true,
                                    attachDate: new Date().toISOString().split('T')[0],
                                  })
                                }
                                disabled={isAnyOrphanBusy}
                              >
                                <Link className="mr-1 h-3 w-3" />
                                Add to day
                              </Button>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => handleIgnore(filename)}
                                disabled={isAnyOrphanBusy}
                              >
                                <EyeOff className="mr-1 h-3 w-3" />
                                Keep
                              </Button>
                            </>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Ignored orphans — collapsed */}
          {!isLoading && ignoredOrphans.length > 0 && (
            <div className="rounded-lg border border-zinc-200 p-3 dark:border-zinc-700">
              <button
                className="flex w-full items-center gap-2 text-left"
                onClick={() => setIgnoredExpanded((v) => !v)}
              >
                {ignoredExpanded ? (
                  <ChevronDown className="h-4 w-4 text-zinc-400" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-zinc-400" />
                )}
                <EyeOff className="h-4 w-4 text-zinc-400" />
                <span className="text-sm text-zinc-500">
                  Ignored ({ignoredOrphans.length})
                </span>
              </button>

              {ignoredExpanded && (
                <div className="mt-2 flex flex-col gap-2">
                  {ignoredOrphans.map((filename) => {
                    const imgUrl = IMAGE_EXTENSIONS.test(filename)
                      ? assetsApi.getAssetUrl(filename)
                      : null;
                    return (
                      <div
                        key={filename}
                        className="flex items-center gap-2 rounded-md border border-zinc-100 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-800/50"
                      >
                        {imgUrl ? (
                          <div className="group relative flex-shrink-0">
                            <img
                              src={imgUrl}
                              alt={filename}
                              className="h-8 w-8 rounded object-cover opacity-50"
                              onError={(e) => {
                                (e.target as HTMLImageElement).closest('div')!.style.display = 'none';
                              }}
                            />
                            <div className="pointer-events-none absolute bottom-0 right-full z-50 mr-2 hidden w-64 group-hover:block">
                              <img
                                src={imgUrl}
                                alt={filename}
                                className="max-h-64 w-full rounded-md object-contain shadow-xl ring-1 ring-zinc-200 dark:ring-zinc-700"
                              />
                            </div>
                          </div>
                        ) : (
                          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded bg-zinc-200 opacity-50 dark:bg-zinc-700">
                            <span className="text-[9px] uppercase text-zinc-400">
                              {filename.split('.').pop()?.slice(0, 4) ?? 'file'}
                            </span>
                          </div>
                        )}
                        <span className="flex-1 truncate font-mono text-xs text-zinc-400 dark:text-zinc-500">
                          {filename}
                        </span>
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => unignoreOrphan.mutateAsync(filename)}
                          disabled={unignoreOrphan.isPending}
                        >
                          <Eye className="mr-1 h-3 w-3" />
                          Un-ignore
                        </Button>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </Drawer>
  );
}
