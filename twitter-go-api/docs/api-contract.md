# Twitter Go API Contract (Frontend)

Base URL: `/api/v1`

## Auth & Session
- Access token is accepted from either:
1. Cookie: `access_token`
2. Header: `Authorization: Bearer <token>`
- Login/refresh set HttpOnly cookies:
1. `access_token` (path `/`)
2. `refresh_token` (path `/api/v1/auth/refresh`)
- Send requests with `credentials: include` from frontend.

## Common Query Params
- Pagination (where supported):
1. `page` (default `0`, min `0`)
2. `size` (default `20`, max `50`)

## PageResponse<T>
```json
{
  "content": [],
  "page": 0,
  "size": 20,
  "totalElements": 100,
  "totalPages": 5,
  "first": true,
  "last": false
}
```

## Common Error Shape
```json
{
  "code": "BAD_REQUEST|UNAUTHORIZED|FORBIDDEN|NOT_FOUND|CONFLICT|INTERNAL_ERROR|VALIDATION_ERROR|TOO_MANY_REQUESTS",
  "message": "string",
  "details": [
    { "field": "string", "message": "string" }
  ]
}
```

## Data Models

### UserResponse
```json
{
  "id": 1,
  "username": "alice",
  "email": "alice@example.com",
  "displayName": "Alice",
  "bio": "hello",
  "avatarUrl": "https://...",
  "isFollowing": false,
  "followersCount": 10,
  "followingCount": 20
}
```

### TweetResponse
```json
{
  "id": 1,
  "content": "tweet text",
  "mediaType": "IMAGE|VIDEO|NONE",
  "mediaUrl": "https://...",
  "user": { "...UserResponse" },
  "replyCount": 0,
  "likeCount": 0,
  "retweetCount": 0,
  "isLiked": false,
  "isRetweeted": false,
  "retweetedTweet": { "...TweetResponse" },
  "replyToTweetId": 123,
  "replyToUsername": "alice",
  "createdAt": "2026-03-02T00:00:00Z"
}
```

### HashtagResponse
```json
{
  "id": 1,
  "text": "golang",
  "usageCount": 12,
  "lastUsedAt": "2026-03-02T00:00:00Z",
  "createdAt": "2026-03-01T00:00:00Z"
}
```

### NotificationResponse
```json
{
  "id": 1,
  "actor": { "...UserResponse" },
  "tweetId": 10,
  "tweetContent": "tweet text",
  "tweetMediaUrl": "https://...",
  "type": "LIKE|REPLY|RETWEET|FOLLOW",
  "isRead": false,
  "createdAt": "2026-03-02T00:00:00Z"
}
```

## Endpoints

## Auth

### POST `/auth/google`
Body:
```json
{ "idToken": "google_id_token" }
```
Response 200:
```json
{
  "accessToken": "jwt",
  "user": { "...UserResponse" }
}
```

### POST `/auth/refresh`
- Requires `refresh_token` cookie.
Response 200:
```json
{ "accessToken": "jwt" }
```

### POST `/auth/logout`
Response 200:
```json
{ "success": true }
```

### GET `/auth/me` (private)
Response 200: `UserResponse`

## Users

### GET `/users/:id` (optional auth)
Response 200: `UserResponse`

### PUT `/users/profile` (private)
Supports:
1. `application/json`
```json
{ "bio": "...", "displayName": "..." }
```
2. `multipart/form-data`
- `data`: JSON string of `{ bio?, displayName? }`
- `avatar`: image file

Response 200: `UserResponse`

### POST `/users/:id/follow` (private)
Response 200:
```json
{ "success": true }
```

### DELETE `/users/:id/follow` (private)
Response 200:
```json
{ "success": true }
```

### GET `/users/:id/followers` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<UserResponse>`

### GET `/users/:id/following` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<UserResponse>`

## Tweets

### POST `/tweets` (private)
`multipart/form-data`:
- `data` (required): JSON string
```json
{ "content": "text", "parentId": 123 }
```
- `media` (optional): image/video file

Response 201: `TweetResponse`

### DELETE `/tweets/:id` (private)
Response 200:
```json
{ "success": true }
```

### GET `/tweets/:id` (optional auth)
Response 200: `TweetResponse`

### GET `/tweets/:id/replies` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<TweetResponse>`

### POST `/tweets/:id/like` (private)
Response 200:
```json
{ "success": true }
```

### DELETE `/tweets/:id/like` (private)
Response 200:
```json
{ "success": true }
```

### POST `/tweets/:id/retweet` (private)
Response 200: `TweetResponse`

### DELETE `/tweets/:id/retweet` (private)
Response 200:
```json
{ "success": true }
```

## Feeds

### GET `/feeds/global` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<TweetResponse>`

### GET `/feeds/user/:id` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<TweetResponse>`

### GET `/feeds/following` (private)
Query: `page`, `size`
Response 200: `PageResponse<TweetResponse>`

## Search

### GET `/search/users` (optional auth)
Query:
1. `q` (required)
2. `page`, `size`
Response 200: `PageResponse<UserResponse>`

### GET `/search/tweets` (optional auth)
Query:
1. `q` (required)
2. `page`, `size`
Response 200: `PageResponse<TweetResponse>`

### GET `/search/hashtags` (optional auth)
Query:
1. `q` (optional, empty -> `[]`)
2. `limit` (default `5`, max `50`)
Response 200: `HashtagResponse[]`

## Discovery

### GET `/discovery/trending` (optional auth)
Query:
1. `limit` (default `10`, max `50`)
Response 200: `HashtagResponse[]`

### GET `/discovery/users` (optional auth)
Query: `page`, `size`
Response 200: `PageResponse<UserResponse>`

## Notifications

### GET `/notifications` (private)
Query: `page`, `size`
Response 200: `PageResponse<NotificationResponse>`

### GET `/notifications/unread-count` (private)
Response 200: number

### POST `/notifications/mark-read` (private)
Response 200:
```json
{ "success": true }
```

### GET `/notifications/stream` (private, SSE)
- Content-Type: `text/event-stream`
- Events:
1. `connected`
2. `ping`
3. `notification` (payload `NotificationResponse`)
