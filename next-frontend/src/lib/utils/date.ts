import { format, parseISO, isValid } from 'date-fns';

export function formatDate(date: string | Date, formatStr: string = 'MMMM d, yyyy'): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return isValid(d) ? format(d, formatStr) : '';
}

export function formatDateForApi(date: Date): string {
  return format(date, 'yyyy-MM-dd');
}

export function formatDayOfWeek(date: string | Date): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return isValid(d) ? format(d, 'EEEE') : '';
}

export function formatTime(date: string | Date): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return isValid(d) ? format(d, 'hh:mm a') : '';
}

export function formatFullDate(date: string | Date): string {
  const d = typeof date === 'string' ? parseISO(date) : date;
  if (!isValid(d)) return '';
  return `${format(d, 'EEEE')}, ${format(d, 'MMMM d, yyyy')}`;
}

export function getTodayString(): string {
  return formatDateForApi(new Date());
}
