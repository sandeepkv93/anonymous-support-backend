# API Examples

This document provides example workflows for common operations using the Anonymous Support API.

## Table of Contents
1. [Authentication](#authentication)
2. [Creating Posts](#creating-posts)
3. [Responding to Posts](#responding-to-posts)
4. [Circle Management](#circle-management)
5. [User Profile](#user-profile)

## Authentication

### Register Anonymous User

Create an anonymous account (no email required):

```bash
curl -X POST http://localhost:8080/api.auth.v1.AuthService/RegisterAnonymous \
  -H "Content-Type: application/json" \
  -d '{
    "username": "anonymous_user_123"
  }'
```

Response:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "anonymous_user_123",
    "isAnonymous": true,
    "strengthPoints": 0
  }
}
```

### Register with Email

Create a full account with email:

```bash
curl -X POST http://localhost:8080/api.auth.v1.AuthService/RegisterWithEmail \
  -H "Content-Type: application/json" \
  -d '{
    "username": "recovery_warrior",
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### Login

Authenticate with email and password:

```bash
curl -X POST http://localhost:8080/api.auth.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### Refresh Access Token

Get a new access token using refresh token:

```bash
curl -X POST http://localhost:8080/api.auth.v1.AuthService/RefreshToken \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

## Creating Posts

### Create SOS Post

Create an urgent support request:

```bash
curl -X POST http://localhost:8080/api.post.v1.PostService/CreatePost \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "SOS",
    "content": "Having strong cravings right now. Need support.",
    "categories": ["alcohol", "cravings"],
    "urgencyLevel": 5,
    "daysSinceRelapse": 30,
    "timeContext": "evening",
    "tags": ["trigger", "urgent"]
  }'
```

Response:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "username": "recovery_warrior",
  "type": "SOS",
  "content": "Having strong cravings right now. Need support.",
  "categories": ["alcohol", "cravings"],
  "urgencyLevel": 5,
  "responseCount": 0,
  "supportCount": 0,
  "createdAt": "2025-01-15T20:30:00Z"
}
```

### Create Victory Post

Share a milestone or achievement:

```bash
curl -X POST http://localhost:8080/api.post.v1.PostService/CreatePost \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Victory",
    "content": "90 days sober today! Feeling grateful.",
    "categories": ["alcohol", "milestone"],
    "urgencyLevel": 1,
    "daysSinceRelapse": 90,
    "tags": ["90days", "milestone"]
  }'
```

### Get Feed

Retrieve recent posts:

```bash
curl -X GET "http://localhost:8080/api.post.v1.PostService/GetFeed?limit=20&offset=0&category=alcohol" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Responding to Posts

### Quick Support

Send quick one-tap support:

```bash
curl -X POST http://localhost:8080/api.support.v1.SupportService/QuickSupport \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "postId": "507f1f77bcf86cd799439011"
  }'
```

### Create Text Response

Send a supportive message:

```bash
curl -X POST http://localhost:8080/api.support.v1.SupportService/CreateResponse \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "postId": "507f1f77bcf86cd799439011",
    "type": "Text",
    "content": "You got this! Remember why you started. Take it one hour at a time."
  }'
```

Response:
```json
{
  "id": "507f1f77bcf86cd799439012",
  "postId": "507f1f77bcf86cd799439011",
  "userId": "550e8400-e29b-41d4-a716-446655440001",
  "username": "supportive_friend",
  "type": "Text",
  "content": "You got this! Remember why you started. Take it one hour at a time.",
  "strengthPoints": 10,
  "createdAt": "2025-01-15T20:35:00Z"
}
```

### Get Responses for a Post

Retrieve all responses to a post:

```bash
curl -X GET "http://localhost:8080/api.support.v1.SupportService/GetResponses?postId=507f1f77bcf86cd799439011&limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Circle Management

### Create a Circle

Create a support community:

```bash
curl -X POST http://localhost:8080/api.circle.v1.CircleService/CreateCircle \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Evening Warriors",
    "description": "Support for those facing evening triggers",
    "category": "alcohol",
    "maxMembers": 100,
    "isPrivate": false
  }'
```

### Join a Circle

Join an existing circle:

```bash
curl -X POST http://localhost:8080/api.circle.v1.CircleService/JoinCircle \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "circleId": "650e8400-e29b-41d4-a716-446655440000"
  }'
```

### Get Circle Feed

Get posts from a specific circle:

```bash
curl -X GET "http://localhost:8080/api.circle.v1.CircleService/GetCircleFeed?circleId=650e8400-e29b-41d4-a716-446655440000&limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Browse Circles

Discover available circles:

```bash
curl -X GET "http://localhost:8080/api.circle.v1.CircleService/GetCircles?category=alcohol&limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## User Profile

### Get Profile

Retrieve user profile and streak data:

```bash
curl -X GET http://localhost:8080/api.user.v1.UserService/GetProfile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Response:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "recovery_warrior",
    "email": "user@example.com",
    "strengthPoints": 250,
    "isAnonymous": false
  },
  "streak": {
    "streakDays": 90,
    "totalCravings": 45,
    "cravingsResisted": 40,
    "lastRelapseDate": "2024-10-15T00:00:00Z",
    "goals": ["Stay sober", "Help others"],
    "milestones": ["30 days", "60 days", "90 days"]
  }
}
```

### Update Profile

Update username and avatar:

```bash
curl -X POST http://localhost:8080/api.user.v1.UserService/UpdateProfile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_username",
    "avatarId": "avatar_5"
  }'
```

### Update Streak

Record streak update or relapse:

```bash
curl -X POST http://localhost:8080/api.user.v1.UserService/UpdateStreak \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hasRelapsed": false
  }'
```

## Error Handling

All endpoints return structured errors with appropriate HTTP status codes:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid post content",
  "details": "Content must be between 1 and 5000 characters"
}
```

Common error codes:
- `VALIDATION_ERROR` (400) - Invalid request data
- `UNAUTHORIZED` (401) - Missing or invalid authentication
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource not found
- `CONFLICT` (409) - Resource already exists
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `INTERNAL_ERROR` (500) - Server error

## WebSocket Real-time Updates

Connect to WebSocket for real-time notifications:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  // Send authentication
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'YOUR_ACCESS_TOKEN'
  }));

  // Subscribe to channels
  ws.send(JSON.stringify({
    type: 'subscribe',
    channels: ['posts', 'responses']
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

Real-time event types:
- `new_post` - A new post was created
- `new_response` - A new response was added
- `post_deleted` - A post was deleted
- `supporter_joined` - Someone sent quick support

## Rate Limits

Default rate limits:
- Posts: 10 per hour
- Responses: 100 per hour

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1642350000
```
