export type Role = 'USER' | 'ADMIN';

export interface UserResponse {
    id: number;
    username: string;
    email: string;
    displayName: string | null;
    avatarUrl: string | null;
    bio: string | null;
    followersCount: number;
    followingCount: number;
    isFollowing: boolean;
}

export interface UpdateProfileRequest {
    displayName?: string;
    bio?: string;
}
