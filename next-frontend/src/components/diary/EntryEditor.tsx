'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Save, Sparkles, Plus, X } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, Textarea } from '@/components/ui';
import { cn } from '@/lib/utils';
import { ImageGrid } from '@/components/assets';
import { useSaveEntry } from '@/hooks';
import { formatDateForApi, formatFullDate } from '@/lib/utils/date';
import { diaryApi, assetsApi, authApi, getErrorMessage } from '@/lib/api';
import { useToast } from '@/providers';
import { ApiError, type DiaryEntry } from '@/types';

// Max existing-tag autocomplete options shown at once.
const MAX_TAG_SUGGESTIONS = 8;

function parseTags(value: string): string[] {
  return value.split(',').map((t) => t.trim()).filter(Boolean);
}

const entrySchema = z.object({
  title: z.string(),
  body: z.string(),
  tags: z.string(),
});

type EntryFormData = z.infer<typeof entrySchema>;

export interface EntryEditorProps {
  entry?: DiaryEntry | null;
  initialDate?: string;
  onClose?: () => void;
  onSave?: () => void;
}

export function EntryEditor({ entry, initialDate, onClose, onSave }: EntryEditorProps) {
  const router = useRouter();
  const toast = useToast();
  const saveEntry = useSaveEntry();
  const [currentDate, setCurrentDate] = useState(
    entry?.date || initialDate || formatDateForApi(new Date())
  );
  const [pendingDate, setPendingDate] = useState<string | null>(null);
  const [attachedImages, setAttachedImages] = useState<string[]>([]);
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);

  // AI tag suggestions
  const [aiEnabled, setAiEnabled] = useState(false);
  const [suggestedTags, setSuggestedTags] = useState<string[]>(entry?.pendingTags ?? []);
  const [suggesting, setSuggesting] = useState(false);

  // Tag autocomplete (non-AI): the family's existing vocabulary + dropdown state.
  const [knownTags, setKnownTags] = useState<string[]>([]);
  const [tagsFocused, setTagsFocused] = useState(false);
  // The inline "add tag" input value (separate from committed chips).
  const [tagInput, setTagInput] = useState('');
  const tagInputRef = useRef<HTMLInputElement>(null);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    getValues,
    trigger,
    formState: { errors, isDirty },
  } = useForm<EntryFormData>({
    resolver: zodResolver(entrySchema),
    defaultValues: {
      title: entry?.title || '',
      body: entry?.body || '',
      tags: entry?.tags?.join(', ') || '',
    },
  });

  const bodyValue = watch('body');
  const tagsValue = watch('tags');

  // Check whether the family has AI tagging enabled (controls suggest UI).
  useEffect(() => {
    let cancelled = false;
    authApi
      .getFamily()
      .then((f) => { if (!cancelled) setAiEnabled(!!f.aiTaggingEnabled); })
      .catch(() => { if (!cancelled) setAiEnabled(false); });
    return () => { cancelled = true; };
  }, []);

  // Fetch suggestions for the current draft (explicit button only).
  // Never writes tags — only surfaces accept-able chips.
  const fetchSuggestions = useCallback(async () => {
    const title = getValues('title');
    const body = getValues('body');
    if (!title.trim() && !body.trim()) return;
    setSuggesting(true);
    try {
      const res = await diaryApi.suggestTags({ date: currentDate, title, body });
      setSuggestedTags(res.tags.map((t) => t.name));
      // A successful-but-empty result is a normal outcome (the model had
      // nothing, was blocked, or hit a token limit). Tell the user rather than
      // leaving the button appearing to do nothing.
      if (res.tags.length === 0) {
        toast.show('No tag suggestions for this entry.', 'info');
      }
    } catch (error) {
      toast.error(getErrorMessage(error));
    } finally {
      setSuggesting(false);
    }
  }, [currentDate, getValues, toast]);

  // Accept a suggested tag: add it to the tags field and remove the chip. For an
  // existing entry, also persist it immediately (mirrors dismiss) so it sticks
  // even without saving; for a brand-new unsaved entry it persists on save.
  const acceptSuggestion = useCallback((name: string) => {
    const current = parseTags(getValues('tags'));
    if (!current.some((t) => t.toLowerCase() === name.toLowerCase())) {
      setValue('tags', [...current, name].join(', '), { shouldDirty: true });
    }
    setSuggestedTags((prev) => prev.filter((t) => t !== name));
    diaryApi.acceptTag(currentDate, name).catch((error) => {
      // 404 for a not-yet-saved entry is expected — it persists on save.
      if (!(error instanceof ApiError && error.status === 404)) {
        toast.error(getErrorMessage(error));
      }
    });
  }, [currentDate, getValues, setValue, toast]);

  // Dismiss a suggested tag: drop it locally and clear it from the entry's
  // persisted pending tags (best-effort; the chip is already gone visually).
  const dismissSuggestion = useCallback((name: string) => {
    setSuggestedTags((prev) => prev.filter((t) => t !== name));
    diaryApi.dismissTag(currentDate, name).catch((error) => {
      // 404 for a not-yet-saved entry is expected — nothing to dismiss yet.
      if (!(error instanceof ApiError && error.status === 404)) {
        toast.error(getErrorMessage(error));
      }
    });
  }, [currentDate, toast]);

  // Hide suggestions already present (case-insensitive) in the tags field.
  const visibleSuggestions = suggestedTags.filter(
    (name) => !parseTags(tagsValue).some((t) => t.toLowerCase() === name.toLowerCase())
  );

  // Load the family's existing tag vocabulary once for autocomplete.
  useEffect(() => {
    let cancelled = false;
    diaryApi
      .getTags()
      .then((res) => { if (!cancelled) setKnownTags(res.tags); })
      .catch(() => { if (!cancelled) setKnownTags([]); });
    return () => { cancelled = true; };
  }, []);

  // Confirmed tags rendered as chips, derived from the form's comma-separated
  // value (kept as the source of truth so the save path is unchanged).
  const tags = parseTags(tagsValue);

  // Add a tag chip from the inline input (de-duplicated, case-insensitive).
  const addTag = useCallback((raw: string) => {
    const name = raw.trim();
    setTagInput('');
    if (!name) return;
    const current = parseTags(getValues('tags'));
    if (current.some((t) => t.toLowerCase() === name.toLowerCase())) return;
    setValue('tags', [...current, name].join(', '), { shouldDirty: true });
  }, [getValues, setValue]);

  // Remove a single tag chip.
  const removeTag = useCallback((name: string) => {
    const current = parseTags(getValues('tags'));
    setValue('tags', current.filter((t) => t !== name).join(', '), { shouldDirty: true });
  }, [getValues, setValue]);

  // Autocomplete matches for the inline input: existing tags matching it
  // (case-insensitive substring), excluding tags already added as chips.
  const trimmedInput = tagInput.trim();
  const alreadyEntered = new Set(tags.map((t) => t.toLowerCase()));
  const tagMatches = knownTags
    .filter((t) => !alreadyEntered.has(t.toLowerCase()))
    .filter((t) =>
      trimmedInput === ''
        ? true
        : t.toLowerCase().includes(trimmedInput.toLowerCase()) &&
          t.toLowerCase() !== trimmedInput.toLowerCase()
    )
    .slice(0, MAX_TAG_SUGGESTIONS);

  useEffect(() => {
    if (entry?.body) {
      const imageMatches = entry.body.matchAll(/!\[.*?\]\(([^)]+)\)/g);
      const images = Array.from(imageMatches, (m) => m[1]);
      setAttachedImages(images);
    }
  }, [entry?.body]);

  const reloadForDate = useCallback(async (date: string) => {
    try {
      const fetched = await diaryApi.getItemByDate(date);
      if (fetched) {
        reset({
          title: fetched.title || '',
          body: fetched.body || '',
          tags: fetched.tags?.join(', ') || '',
        });
        const imageMatches = (fetched.body || '').matchAll(/!\[.*?\]\(([^)]+)\)/g);
        setAttachedImages(Array.from(imageMatches, (m) => m[1]));
        setSuggestedTags(fetched.pendingTags ?? []);
      } else {
        reset({ title: '', body: '', tags: '' });
        setAttachedImages([]);
        setSuggestedTags([]);
      }
      setCurrentDate(date);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }, [reset, toast]);

  const handleDateChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const newDate = e.target.value;
    if (!newDate || newDate === currentDate) return;

    if (isDirty) {
      setPendingDate(newDate);
    } else {
      await reloadForDate(newDate);
    }
  }, [currentDate, isDirty, reloadForDate]);

  const handleSaveAndSwitch = async () => {
    const isValid = await trigger();
    if (!isValid) return;
    const data = getValues();
    try {
      await saveEntry.mutateAsync({
        title: data.title,
        date: currentDate,
        body: data.body,
        tags: data.tags.split(',').map((t) => t.trim()).filter(Boolean),
      });
    } catch (error) {
      toast.error(getErrorMessage(error));
      return;
    }
    const target = pendingDate!;
    setPendingDate(null);
    await reloadForDate(target);
  };

  const handleMoveDraft = () => {
    setCurrentDate(pendingDate!);
    setPendingDate(null);
  };

  const handleCancelDateChange = () => {
    setPendingDate(null);
  };

  const handleRemoveImage = useCallback(
    (index: number) => {
      const imageToRemove = attachedImages[index];
      setAttachedImages((prev) => prev.filter((_, i) => i !== index));
      setValue(
        'body',
        bodyValue.replace(new RegExp(`!\\[.*?\\]\\(${imageToRemove.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\)\\n*`, 'g'), ''),
        { shouldDirty: true }
      );
    },
    [attachedImages, bodyValue, setValue]
  );

  const handleFileUpload = useCallback(
    async (files: FileList) => {
      setUploadProgress(0);
      try {
        const response = await assetsApi.uploadAssetsBatch(Array.from(files), setUploadProgress);
        const uploadedUrls = response.files.map((f) => f.savedName);

        setAttachedImages((prev) => [...prev, ...uploadedUrls]);
        const imageMarkdown = uploadedUrls.map((url) => `![](${url})`).join('\n\n');
        setValue('body', `${bodyValue}\n\n${imageMarkdown}`, { shouldDirty: true });
      } catch (error) {
        toast.error(getErrorMessage(error));
      } finally {
        setUploadProgress(null);
      }
    },
    [bodyValue, setValue, toast]
  );

  const onSubmit = async (data: EntryFormData) => {
    try {
      await saveEntry.mutateAsync({
        title: data.title,
        date: currentDate,
        body: data.body,
        tags: data.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      });

      onSave?.();
      router.replace(`/diary/${currentDate}`);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  };

  const handleCancel = () => {
    if (onClose) {
      onClose();
    } else {
      // Fallback: close to the viewer for the current date, consistent with
      // the replace-based navigation model (router.back() could leave the page).
      router.replace(`/diary/${currentDate}`);
    }
  };

  return (
    <>
      <form onSubmit={handleSubmit(onSubmit)} className="flex h-full flex-col">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-zinc-200 px-4 py-3 dark:border-zinc-700 md:px-6">
          <h2 className="text-lg font-semibold text-zinc-900 dark:text-white">
            {entry ? 'Edit Entry' : 'New Entry'}
          </h2>

          <div className="flex items-center gap-2">
            <Button type="button" variant="ghost" onClick={handleCancel}>
              Cancel
            </Button>
            <Button type="submit" isLoading={saveEntry.isPending} className="gap-2">
              <Save className="h-4 w-4" />
              Save Changes
            </Button>
          </div>
        </div>

        {/* Form */}
        <div className="flex-1 overflow-y-auto p-4 md:p-6">
          <div className="mx-auto max-w-3xl space-y-6">
            {/* Title & Date */}
            <div className="grid gap-4 md:grid-cols-2">
              <Input
                label="Title"
                placeholder="Enter a title..."
                error={errors.title?.message}
                {...register('title')}
              />

              <Input
                label="Date"
                type="date"
                value={currentDate}
                onChange={handleDateChange}
              />
            </div>

            {/* Tags */}
            <div>
              {/* Keep the form field registered (and submitted) while it is
                  edited through the chip UI below. */}
              <input type="hidden" {...register('tags')} />
              <label className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                Tags
              </label>
              <div className="flex items-end gap-2">
                <div className="relative flex-1">
                  <div
                    className="flex min-h-[2.5rem] flex-wrap items-center gap-1.5 rounded-lg border border-zinc-300 bg-white px-2 py-1.5 focus-within:border-zinc-500 focus-within:ring-1 focus-within:ring-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:focus-within:border-zinc-500 dark:focus-within:ring-zinc-500"
                    data-testid="tags-chip-field"
                    onClick={() => tagInputRef.current?.focus()}
                  >
                    {tags.map((tag) => (
                      <span
                        key={tag}
                        className="inline-flex items-center gap-1 rounded-full bg-zinc-100 py-0.5 pl-2.5 pr-1 text-sm text-zinc-700 dark:bg-zinc-700 dark:text-zinc-200"
                        data-testid="tag-chip"
                      >
                        {tag}
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            removeTag(tag);
                          }}
                          className="rounded-full p-0.5 text-zinc-400 transition hover:bg-zinc-200 hover:text-zinc-700 dark:hover:bg-zinc-600 dark:hover:text-zinc-100"
                          title={`Remove "${tag}"`}
                          aria-label={`Remove ${tag}`}
                          data-testid="tag-chip-remove"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </span>
                    ))}
                    <input
                      ref={tagInputRef}
                      type="text"
                      value={tagInput}
                      placeholder={tags.length === 0 ? 'Add tags...' : ''}
                      autoComplete="off"
                      className="min-w-[6rem] flex-1 border-0 bg-transparent px-1 py-0.5 text-sm text-zinc-900 placeholder:text-zinc-400 focus:outline-none focus:ring-0 dark:text-white dark:placeholder:text-zinc-500"
                      data-testid="tags-input"
                      onChange={(e) => setTagInput(e.target.value)}
                      onFocus={() => setTagsFocused(true)}
                      onBlur={() => {
                        // Delay so a mousedown on an option registers before close.
                        setTimeout(() => setTagsFocused(false), 150);
                      }}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ',') {
                          e.preventDefault();
                          addTag(tagInput);
                        } else if (e.key === 'Backspace' && tagInput === '' && tags.length > 0) {
                          removeTag(tags[tags.length - 1]);
                        }
                      }}
                    />
                  </div>
                  {tagsFocused && tagMatches.length > 0 && (
                    <ul
                      className="absolute z-20 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-zinc-200 bg-white py-1 shadow-lg dark:border-zinc-700 dark:bg-zinc-800"
                      data-testid="tag-autocomplete"
                    >
                      {tagMatches.map((tag) => (
                        <li key={tag}>
                          <button
                            type="button"
                            // onMouseDown (not onClick) so it fires before input blur.
                            onMouseDown={(e) => {
                              e.preventDefault();
                              addTag(tag);
                            }}
                            className="block w-full px-3 py-1.5 text-left text-sm text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-700"
                            data-testid="tag-autocomplete-option"
                          >
                            {tag}
                          </button>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
                {aiEnabled && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => void fetchSuggestions()}
                    isLoading={suggesting}
                    className="gap-2"
                    data-testid="suggest-tags-button"
                  >
                    <Sparkles className="h-4 w-4" />
                    Suggest tags
                  </Button>
                )}
              </div>

              {/* Staged suggestions are reviewable regardless of the AI toggle —
                  they may have been generated earlier (e.g. by the backfill). */}
              {visibleSuggestions.length > 0 && (
                <div
                  className="mt-2 flex flex-wrap items-center gap-2"
                  data-testid="tag-suggestions"
                >
                  <span className="text-xs text-zinc-500 dark:text-zinc-400">Suggested:</span>
                  {visibleSuggestions.map((name) => (
                    <span
                      key={name}
                      className="inline-flex items-center rounded-full border border-dashed border-blue-400 bg-blue-50 text-sm text-blue-700 dark:border-blue-500 dark:bg-blue-950 dark:text-blue-300"
                      data-testid="tag-suggestion-chip"
                    >
                      <button
                        type="button"
                        onClick={() => acceptSuggestion(name)}
                        className="inline-flex items-center gap-1 rounded-l-full py-0.5 pl-2.5 pr-1 transition hover:bg-blue-100 dark:hover:bg-blue-900"
                        title={`Accept "${name}"`}
                        data-testid="tag-suggestion-accept"
                      >
                        <Plus className="h-3 w-3" />
                        {name}
                      </button>
                      <button
                        type="button"
                        onClick={() => dismissSuggestion(name)}
                        className="rounded-r-full py-0.5 pl-1 pr-2 text-blue-400 transition hover:bg-blue-100 hover:text-blue-700 dark:hover:bg-blue-900 dark:hover:text-blue-200"
                        title={`Dismiss "${name}"`}
                        aria-label={`Dismiss ${name}`}
                        data-testid="tag-suggestion-dismiss"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </span>
                  ))}
                </div>
              )}
            </div>

            {/* Content */}
            <Textarea
              label="Content (Markdown Supported)"
              placeholder="Write your thoughts..."
              monospace
              className="min-h-[300px]"
              error={errors.body?.message}
              {...register('body')}
            />

            {/* Attached Images */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                Attached Images
              </label>

              {uploadProgress !== null && (
                <div className="mb-4 rounded-lg border border-zinc-200 p-4 dark:border-zinc-700">
                  <div className="mb-2 flex justify-between text-sm text-zinc-600 dark:text-zinc-400">
                    <span>Uploading...</span>
                    <span>{uploadProgress}%</span>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-700">
                    <div
                      className="h-full rounded-full bg-blue-500 transition-all duration-200"
                      style={{ width: `${uploadProgress}%` }}
                    />
                  </div>
                </div>
              )}

              <div
                className={cn(
                  'mb-4 rounded-lg border-2 border-dashed border-zinc-300 p-4 text-center dark:border-zinc-700',
                  uploadProgress !== null && 'pointer-events-none opacity-50'
                )}
                onDragOver={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                }}
                onDrop={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  if (e.dataTransfer.files.length > 0) {
                    handleFileUpload(e.dataTransfer.files);
                  }
                }}
              >
                <p className="text-sm text-zinc-500 dark:text-zinc-400">
                  Drag and drop images here, or{' '}
                  <label className="cursor-pointer text-blue-600 hover:underline dark:text-blue-400">
                    browse
                    <input
                      type="file"
                      accept="image/*,video/*"
                      multiple
                      className="hidden"
                      onChange={(e) => {
                        if (e.target.files && e.target.files.length > 0) {
                          handleFileUpload(e.target.files);
                        }
                      }}
                    />
                  </label>
                </p>
              </div>

              {attachedImages.length > 0 && (
                <ImageGrid images={attachedImages} onRemove={handleRemoveImage} />
              )}
            </div>
          </div>
        </div>
      </form>

      {/* Unsaved changes dialog when date is changed */}
      <Modal
        isOpen={!!pendingDate}
        onClose={handleCancelDateChange}
        title="Unsaved Changes"
        fullScreenOnMobile={false}
      >
        <div className="p-6">
          <p className="mb-6 text-zinc-600 dark:text-zinc-400">
            You have unsaved changes to <span className="font-medium text-zinc-900 dark:text-white">{formatFullDate(currentDate)}</span>.
            What would you like to do?
          </p>
          {saveEntry.isError && (
            <p className="mb-2 text-sm text-red-600 dark:text-red-400">Save failed. Please try again.</p>
          )}
          <div className="flex flex-col gap-3">
            <Button onClick={handleSaveAndSwitch} isLoading={saveEntry.isPending} className="w-full">
              Save to {formatFullDate(currentDate)}
            </Button>
            <Button onClick={handleMoveDraft} variant="secondary" className="w-full">
              Move draft to {pendingDate ? formatFullDate(pendingDate) : ''}
            </Button>
            <Button onClick={handleCancelDateChange} variant="ghost" className="w-full">
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
