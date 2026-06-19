'use client';

import { useState, useCallback, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Save } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, Textarea } from '@/components/ui';
import { cn } from '@/lib/utils';
import { ImageGrid } from '@/components/assets';
import { useSaveEntry } from '@/hooks';
import { formatDateForApi, formatFullDate } from '@/lib/utils/date';
import { diaryApi, assetsApi } from '@/lib/api';
import type { DiaryEntry } from '@/types';

const entrySchema = z.object({
  title: z.string().min(1, 'Title is required'),
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
  const saveEntry = useSaveEntry();
  const [currentDate, setCurrentDate] = useState(
    entry?.date || initialDate || formatDateForApi(new Date())
  );
  const [pendingDate, setPendingDate] = useState<string | null>(null);
  const [attachedImages, setAttachedImages] = useState<string[]>([]);
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);

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

  // eslint-disable-next-line react-hooks/incompatible-library
  const bodyValue = watch('body');

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
      } else {
        reset({ title: '', body: '', tags: '' });
        setAttachedImages([]);
      }
      setCurrentDate(date);
    } catch (error) {
      console.error('Failed to load entry for date:', error);
    }
  }, [reset]);

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
      console.error('Save failed:', error);
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
        console.error('Upload failed:', error);
      } finally {
        setUploadProgress(null);
      }
    },
    [bodyValue, setValue]
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
      console.error('Save failed:', error);
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
            <Input
              label="Tags"
              placeholder="Enter tags separated by commas..."
              {...register('tags')}
            />

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
