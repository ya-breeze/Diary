'use client';

import { Calendar, Clock } from 'lucide-react';
import { Badge } from '@/components/ui';
import { MarkdownRenderer } from './MarkdownRenderer';
import { formatFullDate, formatTime } from '@/lib/utils/date';
import { assetsApi } from '@/lib/api';
import type { DiaryEntry } from '@/types';

export interface EntryViewerProps {
  entry: DiaryEntry;
  className?: string;
}

// Video file extensions to exclude from featured image
const VIDEO_EXTENSIONS = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv'];

function isVideoFile(src: string): boolean {
  const lowerSrc = src.toLowerCase();
  return VIDEO_EXTENSIONS.some(ext => lowerSrc.endsWith(ext));
}

export function EntryViewer({ entry, className }: EntryViewerProps) {
  // Extract first image from body if present (for featured image)
  // Skip video files - only use actual images
  const imageMatches = entry.body.matchAll(/!\[.*?\]\(([^)]+)\)/g);
  let featuredImage: string | null = null;

  for (const match of imageMatches) {
    const src = match[1];
    if (!isVideoFile(src)) {
      featuredImage = src;
      break;
    }
  }

  // Get mood from first tag
  const mood = entry.tags?.[0];

  return (
    <article className={className}>
      {/* Header badges */}
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <Badge variant="outline" className="gap-1.5">
          <Calendar className="h-3 w-3" />
          {formatFullDate(entry.date)}
        </Badge>

        <Badge variant="outline" className="gap-1.5">
          <Clock className="h-3 w-3" />
          {formatTime(new Date())}
        </Badge>

        {mood && <Badge variant="mood">{mood}</Badge>}

        {entry.tags?.slice(1).map((tag) => (
          <Badge key={tag} variant="default">
            {tag}
          </Badge>
        ))}
      </div>

      {/* Title */}
      <h1 className="mb-6 font-serif text-3xl font-bold text-zinc-900 dark:text-white md:text-4xl">
        {entry.title || 'Untitled'}
      </h1>

      {/* Featured Image (only if it's an image, not a video) */}
      {featuredImage && (
        <div className="mb-6 overflow-hidden rounded-xl">
          <img
            src={
              featuredImage.startsWith('http')
                ? featuredImage
                : assetsApi.getAssetUrl(featuredImage)
            }
            alt="Featured"
            className="w-full object-cover"
          />
        </div>
      )}

      {/* Body */}
      <MarkdownRenderer content={entry.body} />
    </article>
  );
}
