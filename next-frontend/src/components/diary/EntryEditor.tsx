'use client';

import { useState, useCallback, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Save, ImagePlus } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Textarea } from '@/components/ui';
import { ImageGrid } from '@/components/assets';
import { useSaveEntry } from '@/hooks';
import { formatDateForApi } from '@/lib/utils/date';
import { assetsApi } from '@/lib/api';
import type { DiaryEntry } from '@/types';

const entrySchema = z.object({
  title: z.string().min(1, 'Title is required'),
  date: z.string().min(1, 'Date is required'),
  body: z.string(),
  tags: z.string(), // Comma-separated tags
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
  const [imageUrl, setImageUrl] = useState('');
  const [attachedImages, setAttachedImages] = useState<string[]>([]);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<EntryFormData>({
    resolver: zodResolver(entrySchema),
    defaultValues: {
      title: entry?.title || '',
      date: entry?.date || initialDate || formatDateForApi(new Date()),
      body: entry?.body || '',
      tags: entry?.tags?.join(', ') || '',
    },
  });

  // eslint-disable-next-line react-hooks/incompatible-library
  const bodyValue = watch('body');

  // Extract images from body on mount
  useEffect(() => {
    if (entry?.body) {
      const imageMatches = entry.body.matchAll(/!\[.*?\]\(([^)]+)\)/g);
      const images = Array.from(imageMatches, (m) => m[1]);
      setAttachedImages(images);
    }
  }, [entry?.body]);

  const handleAddImage = useCallback(() => {
    if (imageUrl.trim()) {
      setAttachedImages((prev) => [...prev, imageUrl.trim()]);
      // Add image to body
      setValue('body', `${bodyValue}\n\n![](${imageUrl.trim()})`, { shouldDirty: true });
      setImageUrl('');
    }
  }, [imageUrl, bodyValue, setValue]);

  const handleRemoveImage = useCallback(
    (index: number) => {
      const imageToRemove = attachedImages[index];
      setAttachedImages((prev) => prev.filter((_, i) => i !== index));
      // Remove from body
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
      try {
        const uploadedUrls: string[] = [];
        for (const file of Array.from(files)) {
          const savedName = await assetsApi.uploadAsset(file);
          uploadedUrls.push(savedName);
        }

        setAttachedImages((prev) => [...prev, ...uploadedUrls]);
        const imageMarkdown = uploadedUrls.map((url) => `![](${url})`).join('\n\n');
        setValue('body', `${bodyValue}\n\n${imageMarkdown}`, { shouldDirty: true });
      } catch (error) {
        console.error('Upload failed:', error);
      }
    },
    [bodyValue, setValue]
  );

  const onSubmit = async (data: EntryFormData) => {
    try {
      await saveEntry.mutateAsync({
        title: data.title,
        date: data.date,
        body: data.body,
        tags: data.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      });

      onSave?.();
      if (onClose) {
        onClose();
      } else {
        router.push(`/diary/${data.date}`);
      }
    } catch (error) {
      console.error('Save failed:', error);
    }
  };

  const handleCancel = () => {
    if (onClose) {
      onClose();
    } else {
      router.back();
    }
  };

  return (
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
              error={errors.date?.message}
              {...register('date')}
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

            <div className="mb-4 flex gap-2">
              <Input
                placeholder="Paste image URL here..."
                value={imageUrl}
                onChange={(e) => setImageUrl(e.target.value)}
                className="flex-1"
              />
              <Button type="button" variant="secondary" onClick={handleAddImage} className="gap-2">
                <ImagePlus className="h-4 w-4" />
                Add
              </Button>
            </div>

            {/* Drop zone */}
            <div
              className="mb-4 rounded-lg border-2 border-dashed border-zinc-300 p-4 text-center dark:border-zinc-700"
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
                    accept="image/*"
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

            {/* Image Grid */}
            {attachedImages.length > 0 && (
              <ImageGrid images={attachedImages} onRemove={handleRemoveImage} />
            )}
          </div>
        </div>
      </div>
    </form>
  );
}
