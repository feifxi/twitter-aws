'use client';

import { GoogleLogin, CredentialResponse } from '@react-oauth/google';
import { useState, useRef, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { axiosInstance } from '@/api/axiosInstance';
import { useAuthStore } from '@/store/useAuthStore';
import type { AuthResponse } from '@/types';
import type { GoogleAuthRequestInput } from '@/lib/validation';

import { useUIStore } from '@/store/useUIStore';

import { useQueryClient } from '@tanstack/react-query';

export default function GoogleLoginBtn() {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const closeSignInModal = useUIStore((s) => s.closeSignInModal);
  const queryClient = useQueryClient();

  const [isLoading, setIsLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const [buttonWidth, setButtonWidth] = useState<number>(364); // Default to PC width

  useEffect(() => {
    const updateWidth = () => {
      if (containerRef.current) {
        setButtonWidth(containerRef.current.offsetWidth);
      }
    };
    
    updateWidth();
    const observer = new ResizeObserver(updateWidth);
    if (containerRef.current) observer.observe(containerRef.current);
    
    return () => observer.disconnect();
  }, []);

  const handleSuccess = async (credentialResponse: CredentialResponse) => {
    const googleToken = credentialResponse.credential;
    if (!googleToken) return;
    
    setIsLoading(true);
    try {
      const payload: GoogleAuthRequestInput = { idToken: googleToken };
      const { data } = await axiosInstance.post<AuthResponse>('/auth/google', payload);
      setAuth(data.accessToken, data.user);
      
      await queryClient.invalidateQueries({ queryKey: ['feeds'] });
      await queryClient.invalidateQueries({ queryKey: ['discovery'] });
      await queryClient.invalidateQueries({ queryKey: ['users'] });
      await queryClient.invalidateQueries({ queryKey: ['search'] });

      closeSignInModal();
      router.push('/');
    } catch (error) {
      console.error('Login Failed:', error);
      alert('Login failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div ref={containerRef} className="w-full h-[40px] relative overflow-hidden flex justify-center rounded-full">
      <GoogleLogin
        key={buttonWidth}
        onSuccess={handleSuccess}
        onError={() => console.log('Google Login Failed')}
        theme="filled_black"
        shape="pill"
        size="large"
        width={buttonWidth.toString()}
        text="signup_with"
      />
      {isLoading && (
        <div className="absolute inset-0 bg-black/50 flex items-center justify-center rounded-full z-10 cursor-not-allowed">
           <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
        </div>
      )}
    </div>
  );
}
