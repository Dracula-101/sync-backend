# Sync Social Media Platform - API Routes Documentation

## Base URL
`https://<base-url>/api/v1`

## Authentication
All authenticated routes require a valid JWT token in the Authorization header:
`Authorization: Bearer {token}`

## Routes

### Authentication
- [X] `POST /auth/signup` - New user registration
- [X] `POST /auth/login` - User login with credentials
- [X] `POST /auth/google` - Login with Google Token
- [X] `POST /auth/logout` - User logout
- [X] `POST /auth/forgot-password` - Request password reset
- [X] `POST /auth/refresh-token` - Refresh access token
- [ ] `PUT /auth/reset-password` - Reset password with token (Not implemented)
- [ ] `GET /auth/verify-email/:token` - Verify email address (Not implemented)

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
- [X] `PUT /post/:postId` - Edit post
- [X] `DELETE /post/:postId` - Delete a post
- [X] `POST /post/like/:postId` - Like a post
- [X] `POST /post/dislike/:postId` - Dislike a post
- [X] `POST /post/save/:postId` - Save a post
- [X] `GET /post/get/user` - Get posts by current user
- [X] `GET /post/get/community/:communityId` - Get posts in community
- [X] `POST /post/share/:postId` - Share a post
- [ ] `GET /post/feed` - Get personalized post feed (Not implemented)
- [ ] `GET /post/trending` - Get trending posts (Not implemented)
- [ ] `GET /post/popular` - Get popular posts (Not implemented)
- [ ] `GET /post/saved` - Get saved posts (Not implemented)
- [ ] `POST /post/report/:postId` - Report a post (Not implemented)
- [ ] `GET /post/tags/:tagName` - Get posts with specific tag (Not implemented)

### Communities
- [X] `POST /community/create` - Create new community
- [X] `GET /community/:communityId` - Get specific community
- [X] `PUT /community/:communityId` - Update community
- [X] `DELETE /community/:communityId` - Delete community
- [X] `GET /community/search` - Search communities
- [X] `GET /community/autocomplete` - Autocomplete community names
- [X] `GET /community/trending` - Get trending communities
- [X] `POST /user/join/:communityId` - Join a community
- [X] `POST /user/leave/:communityId` - Leave a community
- [X] `GET /user/communities/owner` - Get communities owned by current user
- [X] `GET /user/communities/joined` - Get communities joined by current user
- [ ] `GET /community/:communityId/members` - Get community members (Not implemented)
- [ ] `POST /community/:communityId/report` - Report a community (Not implemented)
- [ ] `POST /community/:communityId/invite` - Invite user to community (Not implemented)
- [ ] `GET /community/:communityId/invites` - Get community invites (Not implemented)
- [ ] `POST /community/:communityId/invites/:inviteId/accept` - Accept community invite (Not implemented)

### Comments
- [X] `GET /comment/post/:postId` - Get comments for post
- [X] `POST /comment/post/create` - Add comment to post
- [X] `PUT /comment/post/:commentId` - Edit comment
- [X] `DELETE /comment/post/:commentId` - Delete comment
- [X] `GET /comment/post/:postId/reply/:commentId` - Get comment replies
- [X] `POST /comment/post/reply/create` - Reply to comment
- [X] `POST /comment/post/reply/edit/:commentId` - Edit comment reply
- [X] `POST /comment/post/reply/delete/:commentId` - Delete comment reply
- [X] `POST /comment/like/:commentId` - Like a comment
- [X] `POST /comment/dislike/:commentId` - Dislike a comment
- [X] `GET /comment/user/:userId` - Get comments by specific user
- [X] `GET /comment/user` - Get comments by current user

### Community Moderation
- [X] `POST /community/moderator/:communityId/add` - Add moderator
- [X] `DELETE /community/moderator/:communityId/remove` - Remove moderator
- [X] `POST /community/moderator/:communityId/update` - Update moderator permissions
- [X] `GET /community/moderator/:communityId/list` - List moderators
- [X] `GET /community/moderator/:communityId/get/:userId` - Get moderator
- [X] `GET /community/moderator/:communityId/check-permission/:permission` - Check moderator permission
- [X] `POST /community/moderator/:communityId/ban/:userId` - Ban user from community
- [X] `POST /community/moderator/:communityId/unban/:userId` - Unban user from community
- [X] `POST /community/moderator/report/create` - Create report
- [X] `PATCH /community/moderator/report/:reportId/process` - Process report
- [X] `GET /community/moderator/report/:reportId` - Get report details
- [X] `GET /community/moderator/:communityId/reports` - List reports for community
- [X] `GET /community/moderator/:communityId/logs` - Get moderation logs for community

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

### System
- [X] `GET /status/status` - API health check & overall system status
- [X] `GET /status/health` - Get detailed health check information
- [X] `GET /status/routes` - Get all registered API routes
- [ ] `GET /system/config` - Get public system configuration (Not implemented)
- [ ] `POST /system/feedback` - Submit system feedback (Not implemented)

### Analytics
- [ ] `GET /analytics/overview` - Get platform overview statistics (Not implemented)
- [ ] `GET /analytics/users` - Get user statistics (Not implemented)
- [ ] `GET /analytics/content` - Get content statistics (Not implemented)
- [ ] `GET /analytics/engagement` - Get engagement statistics (Not implemented)
- [ ] `GET /analytics/communities` - Get community statistics (Not implemented)

### Platform Administration (Admin Only)
- [ ] `GET /admin/users` - List all users (Not implemented)
- [ ] `GET /admin/communities` - List all communities (Not implemented)
- [ ] `POST /admin/users/:userId/ban` - Ban user platform-wide (Not implemented)
- [ ] `DELETE /admin/users/:userId/ban` - Unban user (Not implemented)
- [ ] `GET /admin/reports` - View all reported content (Not implemented)
- [ ] `PUT /admin/reports/:reportId` - Process report (Not implemented)
- [ ] `GET /admin/logs` - View admin action logs (Not implemented)
- [ ] `GET /admin/metrics` - Get platform health metrics (Not implemented)
