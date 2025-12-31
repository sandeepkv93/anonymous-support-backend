# Anonymous Support API Documentation

## Overview

The Anonymous Support API provides a safe, anonymous platform for users to seek and provide support for addiction recovery and mental health challenges.

## Base URL

```
https://api.anonymous-support.com
```

## Authentication

All API requests require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <access_token>
```

### Register Anonymous User

**POST** `/auth.v1.AuthService/RegisterAnonymous`

Register a new anonymous user.

**Request:**
```json
{
  "username": "anonymous_user_123"
}
```

**Response:**
```json
{
  "userId": "uuid",
  "username": "anonymous_user_123",
  "accessToken": "jwt_token",
  "refreshToken": "jwt_refresh_token"
}
```

### Login

**POST** `/auth.v1.AuthService/Login`

Authenticate existing user.

**Request:**
```json
{
  "username": "user@example.com",
  "password": "secure_password"
}
```

## Posts

### Create Post

**POST** `/post.v1.PostService/CreatePost`

Create a support post.

**Request:**
```json
{
  "type": "POST_TYPE_SOS",
  "content": "I'm struggling today and need support",
  "categories": ["addiction", "alcohol"],
  "urgencyLevel": 3,
  "tags": ["urgent", "support-needed"]
}
```

### Get Feed

**POST** `/post.v1.PostService/GetFeed`

Retrieve personalized feed.

**Request:**
```json
{
  "limit": 20,
  "offset": 0,
  "typeFilter": "POST_TYPE_SOS"
}
```

## WebSocket Real-time

Connect to `wss://api.anonymous-support.com/ws`

**Authentication:**
```json
{
  "type": "auth",
  "token": "your_jwt_token"
}
```

**Subscribe to channels:**
```json
{
  "type": "subscribe",
  "channels": ["posts", "user:your_user_id"]
}
```

## Rate Limits

- Posts: 10 per hour
- Responses: 50 per hour
- API requests: 1000 per hour

## Error Codes

- `UNAUTHENTICATED`: Missing or invalid authentication
- `PERMISSION_DENIED`: Insufficient permissions
- `INVALID_ARGUMENT`: Invalid request parameters
- `NOT_FOUND`: Resource not found
- `RESOURCE_EXHAUSTED`: Rate limit exceeded
