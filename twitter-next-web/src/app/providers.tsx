'use client';

import { GoogleOAuthProvider } from '@react-oauth/google';
import { SignInModal } from '@/components/auth/SignInModal';
import { SignUpModal } from '@/components/auth/SignUpModal';
import { AuthInitializer } from '@/components/auth/AuthInitializer';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

import { ThemeProvider as NextThemesProvider } from "next-themes";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider attribute="class" defaultTheme="dark" enableSystem disableTransitionOnChange>
      <QueryClientProvider client={queryClient}>
      <GoogleOAuthProvider clientId={process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID!}>
        <AuthInitializer>
          {children}
          <SignInModal />
          <SignUpModal />
        </AuthInitializer>
      </GoogleOAuthProvider>
    </QueryClientProvider>
    </NextThemesProvider>
  );
}