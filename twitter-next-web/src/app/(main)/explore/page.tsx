'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search, ArrowLeft } from 'lucide-react';

import { useTrendingHashtags } from '@/hooks/useDiscovery';
import Link from 'next/link';

export default function ExplorePage() {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const { data: trending, isLoading: trendingLoading } = useTrendingHashtags(20);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      router.push(`/search?q=${encodeURIComponent(query)}`);
    }
  };

  return (
    <div className="min-h-screen">
      {/* Search Bar Header */}
      <div className="sticky top-0 z-30 bg-background/60 backdrop-blur-md border-b border-border px-4 py-3 flex items-center gap-4">
        <div 
            onClick={() => router.back()} 
            className="p-2 rounded-full hover:bg-card cursor-pointer transition-colors -ml-2"
        >
            <ArrowLeft className="w-5 h-5 text-foreground" />
        </div>
        <form onSubmit={handleSearch} className="relative flex-1">
          <div className="flex items-center gap-2 rounded-full bg-secondary/50 border border-border focus-within:border-primary focus-within:bg-background px-4 py-2.5 transition-colors">
            <Search className="w-5 h-5 text-muted-foreground shrink-0 focus-within:text-primary" />
            <input
              type="search"
              placeholder="Search Twitter"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="flex-1 min-w-0 bg-transparent text-foreground text-[15px] placeholder:text-muted-foreground outline-none border-none"
            />
          </div>
        </form>
      </div>

      {/* Trending List */}
      <h1 className="text-xl font-bold p-4 border-b border-border">Trends for you</h1>
      
      {trendingLoading ? (
         <div className="p-4 space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-16 bg-card/30 rounded-lg animate-pulse" />
            ))}
         </div>
      ) : (
          <div>
              {trending?.map((item) => (
                <Link
                    key={item.text}
                    href={`/search?q=${encodeURIComponent('#' + item.text)}`}
                    className="block px-4 py-3 hover:bg-card transition-colors border-b border-border"
                >
                    <div className="text-[13px] text-muted-foreground">Trending</div>
                    <div className="font-bold text-foreground text-[16px]">#{item.text}</div>
                    <div className="text-[13px] text-muted-foreground">{item.usageCount} posts</div>
                </Link>
              ))}
          </div>
      )}
    </div>
  );
}
