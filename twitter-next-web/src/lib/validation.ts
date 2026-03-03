import { z } from 'zod';

/** Matches backend @Size(max = 280) on TweetRequest.content */
export const tweetContentSchema = z
  .string()
  .max(280, 'Tweet content must be under 280 characters')
  .optional();

/** Matches backend TweetRequest: content + parentId */
export const tweetRequestSchema = z.object({
  content: z.string().max(280, 'Tweet content must be under 280 characters').optional(),
  parentId: z.number().nullable().optional(),
});

/** Matches backend UpdateProfileRequest @Size constraints */
export const updateProfileSchema = z.object({
  displayName: z.string().max(30, 'Display name cannot exceed 30 characters').optional(),
  bio: z.string().max(160, 'Bio cannot exceed 160 characters').optional(),
});

/** Google auth: token required */
export const googleAuthRequestSchema = z.object({
  idToken: z.string().min(1, 'Token is required'),
});

export type TweetRequestInput = z.infer<typeof tweetRequestSchema>;
export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;
export type GoogleAuthRequestInput = z.infer<typeof googleAuthRequestSchema>;
