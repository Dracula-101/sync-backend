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
- [ ] `PUT /user/me` - Update current user profile
- [ ] `DELETE /user/me` - Delete current user account
- [ ] `GET /user/search` - Search users
- [ ] `PUT /user/password` - Change password
- [ ] `PUT /user/settings` - Update user settings
- [ ] `POST /user/avatar` - Upload profile avatar
- [ ] `DELETE /user/avatar` - Remove profile avatar
- [ ] `GET /user/notifications` - Get user notifications
- [ ] `PUT /user/notification-settings` - Update notification preferences

### Posts
- [X] `POST /post/create` - Create a new post
- [X] `GET /post/get/:postId` - Get specific post
- [X] `POST /post/edit/:postId` - Edit post
- [X] `POST /post/like/:postId` - Like a post
- [X] `POST /post/dislike/:postId` - Dislike a post
- [X] `POST /post/save/:postId` - Save a post
- [X] `GET /post/get/user` - Get posts by current user
- [X] `GET /post/get/community/:communityId` - Get posts in community
- [ ] `GET /post/feed` - Get personalized post feed
- [ ] `GET /post/trending` - Get trending posts
- [ ] `GET /post/popular` - Get popular posts
- [ ] `DELETE /post/:postId` - Delete a post
- [ ] `POST /post/share/:postId` - Share a post
- [ ] `GET /post/saved` - Get saved posts
- [ ] `POST /post/report/:postId` - Report a post
- [ ] `GET /post/tags/:tagName` - Get posts with specific tag

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
- [ ] `GET /comment/post/:postId` - Get comments for post
- [ ] `POST /comment/post/:postId` - Add comment to post
- [ ] `GET /comment/:commentId` - Get specific comment
- [ ] `PUT /comment/:commentId` - Update comment
- [ ] `DELETE /comment/:commentId` - Delete comment
- [ ] `POST /comment/:commentId/reply` - Reply to comment
- [ ] `POST /comment/:commentId/like` - Like a comment
- [ ] `POST /comment/:commentId/dislike` - Dislike a comment
- [ ] `GET /comment/user` - Get comments by current user

### Messaging
- [ ] `GET /message/conversations` - Get user conversations
- [ ] `GET /message/conversation/:userId` - Get messages with specific user
- [ ] `POST /message/send/:userId` - Send message to user
- [ ] `DELETE /message/:messageId` - Delete message
- [ ] `PUT /message/:messageId/read` - Mark message as read
- [ ] `POST /message/read-all` - Mark all messages as read

### Notifications
- [ ] `GET /notification` - Get user notifications
- [ ] `PUT /notification/:notificationId/read` - Mark notification as read
- [ ] `POST /notification/read-all` - Mark all notifications as read
- [ ] `DELETE /notification/:notificationId` - Delete notification
- [ ] `GET /notification/settings` - Get notification settings
- [ ] `PUT /notification/settings` - Update notification settings

### Search
- [ ] `GET /search` - Global search across posts, users, communities
- [ ] `GET /search/users` - Search users
- [ ] `GET /search/posts` - Search posts
- [ ] `GET /search/comments` - Search comments
- [ ] `GET /search/trending` - Get trending search terms

### Media
- [ ] `POST /media/upload` - Upload media files
- [ ] `GET /media/:mediaId` - Get media file metadata
- [ ] `DELETE /media/:mediaId` - Delete uploaded media

### Tags/Topics
- [ ] `GET /tag/trending` - Get trending tags
- [ ] `GET /tag/:tagName/posts` - Get posts with specific tag
- [ ] `GET /tag/follow` - Get tags followed by user
- [ ] `POST /tag/:tagName/follow` - Follow a tag
- [ ] `POST /tag/:tagName/unfollow` - Unfollow a tag

### Moderation (Admin/Moderator Only)
- [ ] `GET /moderation/reports` - View reported content
- [ ] `PUT /moderation/reports/:reportId` - Process report
- [ ] `POST /moderation/ban/:userId` - Ban user
- [ ] `DELETE /moderation/ban/:userId` - Unban user
- [ ] `GET /moderation/banned` - List banned users
- [ ] `POST /moderation/community/:communityId/feature` - Feature a community
- [ ] `DELETE /moderation/community/:communityId/feature` - Unfeature a community

### Analytics (Admin/Owner Only)
- [ ] `GET /analytics/overview` - Get platform overview statistics
- [ ] `GET /analytics/users` - Get user statistics
- [ ] `GET /analytics/content` - Get content statistics
- [ ] `GET /analytics/engagement` - Get engagement statistics
- [ ] `GET /analytics/communities` - Get community statistics

### System
- [ ] `GET /system/status` - API health check
- [ ] `GET /system/config` - Get public system configuration
- [ ] `POST /system/feedback` - Submit system feedback