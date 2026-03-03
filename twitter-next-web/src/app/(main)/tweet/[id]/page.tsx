'use client';

import { use } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { TweetCard, TweetCardSkeleton } from '@/components/TweetCard';
import { CreateTweet } from '@/components/CreateTweet';
import { useTweet, useRepliesInfinite } from '@/hooks/useTweet';

export default function TweetDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const router = useRouter();
  const { id } = use(params);
  const tweetId = parseInt(id, 10);
  const { data: tweet, isLoading: tweetLoading } = useTweet(
    Number.isNaN(tweetId) ? null : tweetId
  );
  const repliesQuery = useRepliesInfinite(
    Number.isNaN(tweetId) ? null : tweetId
  );
  const replyList =
    repliesQuery.data?.pages.flatMap((p) => p.content) ?? [];

  if (Number.isNaN(tweetId)) {
    return (
      <div className="p-4 text-destructive text-[15px]">Invalid tweet ID.</div>
    );
  }

  return (
    <div className="min-h-screen">
      {/* Sticky Header */}
      <div className="sticky top-0 z-30 bg-background/60 backdrop-blur-md border-b border-border px-4 py-3 flex items-center gap-6">
        <button
          onClick={() => router.back()}
          className="p-2 -ml-2 rounded-full hover:bg-card transition-colors"
          aria-label="Back"
        >
          <ArrowLeft className="w-5 h-5 text-foreground" />
        </button>
        <h1 className="text-[20px] font-bold text-foreground">Post</h1>
      </div>

      {tweetLoading || !tweet ? (
        <div className="border-b border-border">
          <TweetCardSkeleton />
        </div>
      ) : (
        <>
          <TweetCard tweet={tweet} disableClick />
          <div className="border-b border-border">
            <CreateTweet
              isReply
              replyToId={tweet.id}
              placeholder="Post your reply"
              className="px-4 py-2 border-none"
            />
          </div>
          <p className="px-4 py-2 text-muted-foreground text-[15px] font-medium">
            Replies
          </p>
          {repliesQuery.isLoading ? (
            <div>
              <TweetCardSkeleton />
              <TweetCardSkeleton />
            </div>
          ) : replyList.length === 0 ? (
            <p className="p-4 text-muted-foreground text-[15px]">No replies yet.</p>
          ) : (
            <div>
              {replyList.map((r) => (
                <TweetCard key={r.id} tweet={r} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
