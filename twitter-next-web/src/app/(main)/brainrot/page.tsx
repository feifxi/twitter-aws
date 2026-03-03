'use client';

import { Suspense, useEffect } from 'react';
import { useInView } from 'react-intersection-observer';
import { useBrainrotFeed } from '@/hooks/useBrainrot';
import { Loader2 } from 'lucide-react';
import Image from 'next/image';

import { TikTokEmbed } from '@/components/TikTokEmbed';

function BrainrotContent() {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = useBrainrotFeed();
  const { ref, inView } = useInView();

  useEffect(() => {
    if (inView && hasNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, fetchNextPage]);

  const items = data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <div className="h-[calc(100vh-60px)] xl:h-screen w-full overflow-y-scroll snap-y snap-mandatory scroll-smooth no-scrollbar bg-background">
      {/* 1. TikTok Video */}
      <div className="h-full w-full snap-start flex flex-col items-center justify-center p-4 relative">
         <div className="relative w-full h-full flex items-center justify-center">
            <TikTokEmbed />
            <div className="absolute bottom-20 left-4 right-4 p-4 bg-background/60 backdrop-blur-sm rounded-xl">
                <p className="text-foreground font-bold text-lg">Peak edit of the year</p>
                <p className="text-muted-foreground text-sm">@dwsy.ae</p>
            </div>
         </div>
      </div>

      {/* 2. Original Feed Items */}
      {items.map((item, i) => (
        <div 
            key={`${item.id}-${i}`} 
            className="h-full w-full snap-start flex flex-col items-center justify-center relative bg-background border-b border-border/30"
        >
            <div className="relative w-full h-full flex items-center justify-center p-2">
              <img
                src={item.url}
                alt={'Brainrot content'}
                className="w-full h-auto max-h-full object-contain"
                loading="lazy"
              />
            </div>
        </div>
      ))}

      {/* Loader / Infinite Scroll Trigger */}
       <div ref={ref} className="h-20 w-full snap-start flex items-center justify-center">
            {(isLoading || isFetchingNextPage) && (
             <Loader2 className="w-8 h-8 animate-spin text-primary" />
            )}
       </div>
    </div>
  );
}

export default function BrainrotPage() {
  return (
    <Suspense fallback={<div className="h-screen w-full flex items-center justify-center bg-background text-foreground">Loading brainrot...</div>}>
      <BrainrotContent />
    </Suspense>
  );
}
