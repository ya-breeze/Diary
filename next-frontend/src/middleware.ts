import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const publicPaths = ['/login', '/api'];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public paths
  if (publicPaths.some((path) => pathname.startsWith(path))) {
    return NextResponse.next();
  }

  // Allow static files
  if (pathname.startsWith('/_next') || pathname.includes('.')) {
    return NextResponse.next();
  }

  // Middlewareprimarily handles initial navigation protection
  // Client-side will handle the actual auth validation

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
