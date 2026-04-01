'use client';

import { X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { assetsApi } from '@/lib/api';

const VIDEO_EXTENSIONS = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv'];

function isVideoFile(src: string): boolean {
  const lower = src.toLowerCase();
  return VIDEO_EXTENSIONS.some((ext) => lower.endsWith(ext));
}

export interface ImageGridProps {
  images: string[];
  onRemove?: (index: number) => void;
  onImageClick?: (index: number) => void;
  className?: string;
  columns?: 2 | 3 | 4;
}

export function ImageGrid({
  images,
  onRemove,
  onImageClick,
  className,
  columns = 4,
}: ImageGridProps) {
  const getImageUrl = (src: string) => {
    if (src.startsWith('http://') || src.startsWith('https://')) {
      return src;
    }
    return assetsApi.getAssetUrl(src);
  };

  const gridCols = {
    2: 'grid-cols-2',
    3: 'grid-cols-2 md:grid-cols-3',
    4: 'grid-cols-2 md:grid-cols-3 lg:grid-cols-4',
  };

  return (
    <div className={cn('grid gap-3', gridCols[columns], className)}>
      {images.map((src, index) => (
        <div
          key={`${src}-${index}`}
          className="group relative aspect-square overflow-hidden rounded-lg bg-zinc-100 dark:bg-zinc-800"
        >
          {isVideoFile(src) ? (
            <video
              src={getImageUrl(src)}
              className="h-full w-full object-cover"
              onClick={() => onImageClick?.(index)}
              controls
              preload="metadata"
            />
          ) : (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={getImageUrl(src)}
              alt={`Attached image ${index + 1}`}
              className={cn(
                'h-full w-full object-cover transition-transform',
                onImageClick && 'cursor-pointer hover:scale-105'
              )}
              onClick={() => onImageClick?.(index)}
            />
          )}

          {onRemove && (
            <button
              type="button"
              onClick={() => onRemove(index)}
              className="absolute right-2 top-2 rounded-full bg-black/50 p-1 text-white opacity-0 transition-opacity hover:bg-black/70 group-hover:opacity-100"
              aria-label="Remove image"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
