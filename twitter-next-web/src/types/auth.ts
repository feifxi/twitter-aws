import type { UserResponse } from './user';

export interface GoogleAuthRequest {
    idToken: string;
}

export interface AuthResponse {
    accessToken: string;
    user: UserResponse;
}
