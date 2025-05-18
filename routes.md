# Sync Social Media Platform - API Routes Documentation

## Base URL
`https://<base-url>/api/v1`

## Authentication
All authenticated routes require a valid JWT token in the Authorization header:
`Authorization: Bearer {token}`

## Routes

### Authentication
- [X] `POST /auth/login` - User login with credentials
- [X] `POST /auth/signup` - New user registration
- [X] `POST /auth/google` - Login with Google Token
- [X] `POST /auth/logout` - User logout
- [X] `POST /auth/forgot-password` - Request password reset
- [X] `POST /auth/refresh-token` - Refresh access token

### User Management
- [X] `GET /user/me` - Get current user profile
- [X] `GET /user/:userId` - Get user by ID
- [X] `POST /user/follow/:userId` - Follow a user
- [X] `POST /user/unfollow/:userId` - Unfollow a user
- [X] `POST /user/block/:userId` - Block a user
- [X] `POST /user/unblock/:userId` - Unblock a user
- [ ] `PUT /user/me` - Update current user profile (Not implemented)
- [ ] `DELETE /user/me` - Delete current user account (Not implemented)
- [ ] `GET /user/search` - Search users (Not implemented)
- [ ] `PUT /user/password` - Change password (Not implemented)
- [ ] `PUT /user/settings` - Update user settings (Not implemented)
- [ ] `POST /user/avatar` - Upload profile avatar (Not implemented)
- [ ] `DELETE /user/avatar` - Remove profile avatar (Not implemented)
- [ ] `GET /user/notifications` - Get user notifications (Not implemented)
- [ ] `PUT /user/notification-settings` - Update notification preferences (Not implemented)

### Posts
- [X] `POST /post/create` - Create a new post
- [X] `GET /post/get/:postId` - Get specific post
- [X] `POST /post/edit/:postId` - Edit post
- [X] `POST /post/like/:postId` - Like a post
- [X] `POST /post/dislike/:postId` - Dislike a post
- [X] `POST /post/save/:postId` - Save a post
- [X] `GET /post/get/user` - Get posts by current user
- [X] `GET /post/get/community/:communityId` - Get posts in community
- [X] `POST /post/share/:postId` - Share a post
- [ ] `GET /post/feed` - Get personalized post feed (Not implemented)
- [ ] `GET /post/trending` - Get trending posts (Not implemented)
- [ ] `GET /post/popular` - Get popular posts (Not implemented)
- [ ] `DELETE /post/:postId` - Delete a post (Not implemented)
- [ ] `GET /post/saved` - Get saved posts (Not implemented)
- [ ] `POST /post/report/:postId` - Report a post (Not implemented)
- [ ] `GET /post/tags/:tagName` - Get posts with specific tag (Not implemented)

### Communities
- [X] `POST /community/create` - Create new community
- [X] `GET /community/:communityId` - Get specific community
- [X] `GET /community/search` - Search communities
- [X] `GET /community/autocomplete` - Autocomplete community names
- [X] `GET /community/trending` - Get trending communities
- [X] `POST /user/join/:communityId` - Join a community
- [X] `POST /user/leave/:communityId` - Leave a community
- [X] `GET /user/communities/owner` - Get communities owned by current user
- [X] `GET /user/communities/joined` - Get communities joined by current user
- [ ] `PUT /community/:communityId` - Update community
- [ ] `DELETE /community/:communityId` - Delete community
- [ ] `GET /community/:communityId/members` - Get community members
- [ ] `POST /community/:communityId/report` - Report a community
- [ ] `POST /community/:communityId/avatar` - Update community avatar
- [ ] `POST /community/:communityId/cover` - Update community cover image
- [ ] `POST /community/:communityId/invite` - Invite user to community
- [ ] `GET /community/:communityId/invites` - Get community invites
- [ ] `POST /community/:communityId/invites/:inviteId/accept` - Accept community invite

### Comments
- [X] `GET /comment/post/:postId` - Get comments for post
- [X] `POST /comment/post/create` - Add comment to post
- [X] `GET /comment/post/:postId/reply/:commentId` - Get comment replies
- [X] `POST /comment/post/edit/:commentId` - Update comment
- [X] `POST /comment/post/delete/:commentId` - Delete comment
- [X] `POST /comment/post/reply/create` - Reply to comment
- [X] `POST /comment/post/reply/edit/:commentId` - Edit comment reply
- [X] `POST /comment/post/reply/delete/:commentId` - Delete comment reply
- [X] `POST /comment/like/:commentId` - Like a comment
- [X] `POST /comment/dislike/:commentId` - Dislike a comment
- [X] `GET /comment/user/:userId` - Get comments by specific user
- [X] `GET /comment/user` - Get comments by current user

### Messaging
- [ ] `GET /message/conversations` - Get user conversations (Not implemented)
- [ ] `GET /message/conversation/:userId` - Get messages with specific user (Not implemented)
- [ ] `POST /message/send/:userId` - Send message to user (Not implemented)
- [ ] `DELETE /message/:messageId` - Delete message (Not implemented)
- [ ] `PUT /message/:messageId/read` - Mark message as read (Not implemented)
- [ ] `POST /message/read-all` - Mark all messages as read (Not implemented)

### Notifications
- [ ] `GET /notification` - Get user notifications (Not implemented)
- [ ] `PUT /notification/:notificationId/read` - Mark notification as read (Not implemented)
- [ ] `POST /notification/read-all` - Mark all notifications as read (Not implemented)
- [ ] `DELETE /notification/:notificationId` - Delete notification (Not implemented)
- [ ] `GET /notification/settings` - Get notification settings (Not implemented)
- [ ] `PUT /notification/settings` - Update notification settings (Not implemented)

### Search
- [ ] `GET /search` - Global search across posts, users, communities (Not implemented)
- [ ] `GET /search/users` - Search users (Not implemented)
- [ ] `GET /search/posts` - Search posts (Not implemented)
- [ ] `GET /search/comments` - Search comments (Not implemented)
- [ ] `GET /search/trending` - Get trending search terms (Not implemented)

### Media
- [ ] `POST /media/upload` - Upload media files (Not implemented)
- [ ] `GET /media/:mediaId` - Get media file metadata (Not implemented)
- [ ] `DELETE /media/:mediaId` - Delete uploaded media (Not implemented)

### Tags/Topics
- [ ] `GET /tag/trending` - Get trending tags (Not implemented)
- [ ] `GET /tag/:tagName/posts` - Get posts with specific tag (Not implemented)
- [ ] `GET /tag/follow` - Get tags followed by user (Not implemented)
- [ ] `POST /tag/:tagName/follow` - Follow a tag (Not implemented)
- [ ] `POST /tag/:tagName/unfollow` - Unfollow a tag (Not implemented)

### Moderation (Admin/Moderator Only)
- [ ] `GET /moderation/reports` - View reported content (Not implemented)
- [ ] `PUT /moderation/reports/:reportId` - Process report (Not implemented)
- [ ] `POST /moderation/ban/:userId` - Ban user (Not implemented)
- [ ] `DELETE /moderation/ban/:userId` - Unban user (Not implemented)
- [ ] `GET /moderation/banned` - List banned users (Not implemented)
- [ ] `POST /moderation/community/:communityId/feature` - Feature a community (Not implemented)
- [ ] `DELETE /moderation/community/:communityId/feature` - Unfeature a community (Not implemented)

### Analytics (Admin/Owner Only)
- [ ] `GET /analytics/overview` - Get platform overview statistics (Not implemented)
- [ ] `GET /analytics/users` - Get user statistics (Not implemented)
- [ ] `GET /analytics/content` - Get content statistics (Not implemented)
- [ ] `GET /analytics/engagement` - Get engagement statistics (Not implemented)
- [ ] `GET /analytics/communities` - Get community statistics (Not implemented)

### System
- [ ] `GET /system/status` - API health check (Not implemented)
- [ ] `GET /system/config` - Get public system configuration (Not implemented)
- [ ] `POST /system/feedback` - Submit system feedback (Not implemented)