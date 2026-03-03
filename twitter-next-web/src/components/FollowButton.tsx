'use client';

import { Button } from '@/components/ui/button';
import { useFollowUser, useUnfollowUser } from '@/hooks/useProfile';
import { useDebouncedToggle } from '@/hooks/useDebouncedToggle';
import { useAuth } from '@/hooks/useAuth';
import { useUIStore } from '@/store/useUIStore';

interface FollowButtonProps {
  userId: number;
  isFollowing: boolean;
  username?: string; // Optional for future use
}

export function FollowButton({ userId, isFollowing }: FollowButtonProps) {
  const followMutation = useFollowUser();
  const unfollowMutation = useUnfollowUser();
  const { isLoggedIn } = useAuth();
  const openSignInModal = useUIStore((s) => s.openSignInModal);

  const {
      optimisticState: following,
      toggle,
  } = useDebouncedToggle({
      initialState: isFollowing,
      onMutate: (newState) => {
          if (newState) followMutation.mutate(userId);
          else unfollowMutation.mutate(userId);
      },
  });

  const handleFollow = (e: React.MouseEvent) => {
    e.stopPropagation();
    e.preventDefault(); // Prevent link navigation if inside a link
    if (!isLoggedIn) {
        openSignInModal();
        return;
    }
    toggle();
  };

  return (
    <Button
      size="sm"
      variant={following ? "outline" : "secondary"}
      onClick={handleFollow}
      className={`rounded-full font-bold h-8 px-4 transition-colors cursor-pointer ${
        following
          ? 'hover:border-destructive hover:text-destructive hover:bg-destructive/10 border-border text-foreground'
          : 'bg-foreground text-background hover:bg-foreground/90'
      }`}
    >
      {following ? 'Following' : 'Follow'}
    </Button>
  );
}
