'use client';

import { useEffect, useState } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { Edit, ChevronLeft, ChevronRight } from 'lucide-react';
import { Button, Modal } from '@/components/ui';
import { EntryViewer, EntryEditor } from '@/components/diary';
import { useDiaryEntry } from '@/hooks';
import { useIsMobile } from '@/hooks';

export default function DiaryEntryPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const isMobile = useIsMobile();

  const date = params.date as string;
  const editParam = searchParams.get('edit');

  const { data: entry, isLoading, refetch } = useDiaryEntry(date);
  const [isEditing, setIsEditing] = useState(editParam === 'true');

  // Sync edit state with URL
  useEffect(() => {
    setIsEditing(editParam === 'true');
  }, [editParam]);

  const handleEdit = () => {
    router.push(`/diary/${date}?edit=true`);
  };

  const handleCloseEdit = () => {
    router.push(`/diary/${date}`);
  };

  const handleSaved = () => {
    refetch();
    handleCloseEdit();
  };

  const handleNavigate = (targetDate: string | null | undefined) => {
    if (targetDate) {
      router.push(`/diary/${targetDate}`);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-900" />
      </div>
    );
  }

  // No entry found - show editor for new entry
  if (!entry) {
    return (
      <div className="h-full">
        <EntryEditor
          initialDate={date}
          onClose={() => router.push('/diary')}
          onSave={handleSaved}
        />
      </div>
    );
  }

  return (
    <>
      {/* Entry view */}
      <div className="mx-auto max-w-3xl px-4 py-6 md:px-8 md:py-8">
        {/* Action bar */}
        <div className="mb-6 flex items-center justify-between">
          {/* Navigation */}
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleNavigate(entry.previousDate)}
              disabled={!entry.previousDate}
              className="gap-1"
            >
              <ChevronLeft className="h-4 w-4" />
              <span className="hidden sm:inline">Previous</span>
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleNavigate(entry.nextDate)}
              disabled={!entry.nextDate}
              className="gap-1"
            >
              <span className="hidden sm:inline">Next</span>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          {/* Edit button */}
          <Button onClick={handleEdit} variant="secondary" size="sm" className="gap-2">
            <Edit className="h-4 w-4" />
            Edit
          </Button>
        </div>

        {/* Entry content */}
        <EntryViewer entry={entry} />
      </div>

      {/* Edit modal/page */}
      {isEditing && (
        isMobile ? (
          // Full-screen on mobile
          <div className="fixed inset-0 z-50 bg-white dark:bg-zinc-900">
            <EntryEditor
              entry={entry}
              onClose={handleCloseEdit}
              onSave={handleSaved}
            />
          </div>
        ) : (
          // Modal on tablet/desktop
          <Modal
            isOpen={isEditing}
            onClose={handleCloseEdit}
            fullScreenOnMobile={false}
            className="h-[90vh] max-h-[800px] w-full max-w-3xl overflow-hidden"
          >
            <EntryEditor
              entry={entry}
              onClose={handleCloseEdit}
              onSave={handleSaved}
            />
          </Modal>
        )
      )}
    </>
  );
}
