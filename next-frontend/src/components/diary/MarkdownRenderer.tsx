'use client';

import { useMemo } from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { cn } from '@/lib/utils';
import { assetsApi } from '@/lib/api';

export interface MarkdownRendererProps {
  content: string;
  className?: string;
}

// Video file extensions
const VIDEO_EXTENSIONS = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv'];

function isVideoFile(src: string): boolean {
  const lowerSrc = src.toLowerCase();
  return VIDEO_EXTENSIONS.some(ext => lowerSrc.endsWith(ext));
}

function getAssetUrl(src: string): string {
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  return assetsApi.getAssetUrl(src);
}

export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
  const html = useMemo(() => {
    if (!content) return '';

    // Configure marked
    marked.setOptions({
      breaks: true,
      gfm: true,
    });

    // Parse markdown to HTML
    let html = marked.parse(content) as string;

    // Transform markdown images to videos if they're video files
    // Matches: <img src="..." alt="...">
    html = html.replace(
      /<img\s+src="([^"]+)"(?:\s+alt="([^"]*)")?[^>]*>/gi,
      (match, src, alt = '') => {
        const assetUrl = getAssetUrl(src);

        if (isVideoFile(src)) {
          // Convert to video tag
          return `<video controls class="rounded-lg shadow-md w-full max-w-2xl" preload="metadata">
            <source src="${assetUrl}" type="video/${src.split('.').pop()?.toLowerCase() || 'mp4'}">
            Your browser does not support the video tag.
          </video>`;
        }

        // Keep as image but with transformed URL
        return `<img src="${assetUrl}" alt="${alt}" class="rounded-lg shadow-md">`;
      }
    );

    // Sanitize HTML - allow video and source tags
    const sanitized = DOMPurify.sanitize(html, {
      ADD_TAGS: ['iframe', 'video', 'source'],
      ADD_ATTR: ['allow', 'allowfullscreen', 'frameborder', 'scrolling', 'controls', 'preload', 'autoplay', 'loop', 'muted', 'playsinline'],
    });

    return sanitized;
  }, [content]);

  return (
    <div
      className={cn(
        'prose prose-zinc dark:prose-invert max-w-none',
        // Typography adjustments
        'prose-headings:font-semibold',
        'prose-p:leading-relaxed',
        'prose-img:rounded-lg prose-img:shadow-md',
        'prose-a:text-blue-600 dark:prose-a:text-blue-400',
        // Video styling
        '[&_video]:rounded-lg [&_video]:shadow-md [&_video]:my-4',
        className
      )}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
