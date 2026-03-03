import type { UserResponse } from './user';

export type NotificationType = 'LIKE' | 'RETWEET' | 'REPLY' | 'FOLLOW';

export interface NotificationResponse {
    id: number;
    type: NotificationType;
    actor: UserResponse;
    tweetId: number | null;
    tweetContent: string | null;
    tweetMediaUrl: string | null;
    originalTweetId?: number | null;
    originalTweetContent?: string | null;
    originalTweetMediaUrl?: string | null;
    isRead: boolean;
    createdAt: string;
}
