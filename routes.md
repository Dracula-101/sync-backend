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
- [X] `POST /auth/google/callback` - Handle Google OAuth callback
- [X] `POST /auth/logout` - User logout
- [ ] `POST /auth/forgot-password` - Request password reset
- [X] `POST /auth/refresh-token` - Refresh access token


### Posts
- [X] `POST /post/create` - Create a new post
- [X] `GET /post/:postId` - Get specific post
- [X] `POST /post/edit/:postId` - Edit post
- [X] `GET /post/get/user` - Get posts by user
- [X] `GET /post/get/community/:communityId` - Get posts in community
- [ ] `GET /post/get/trending` - Get trending posts
- [ ] `GET /post/get/popular` - Get popular posts
- [ ] `GET /post/get/featured` - Get featured posts
- [ ] `GET /post/get/featured/:communityId` - Get featured posts in community


### Communities
- [ ] `GET /communities` - List communities (paginated)
- [ ] `POST /communities` - Create new community
- [ ] `GET /communities/:communityId` - Get specific community
- [ ] `PUT /communities/:communityId` - Update community
- [ ] `DELETE /communities/:communityId` - Delete community
- [ ] `GET /communities/trending` - Get trending communities
- [ ] `GET /communities/search` - Search communities
- [ ] `GET /communities/:communityId/posts` - Get posts in community
- [ ] `POST /communities/:communityId/join` - Join a community
- [ ] `POST /communities/:communityId/leave` - Leave a community
- [ ] `GET /communities/:communityId/members` - Get community members
- [ ] `POST /communities/:communityId/report` - Report a community
- [ ] `GET /users/me/communities` - Get communities joined by current user

### User Management
- [ ] `GET /users/me` - Get current user profile
- [ ] `PUT /users/me` - Update current user profile
- [ ] `DELETE /users/me` - Delete current user account
- [ ] `GET /users/:username` - Get public user profile
- [ ] `GET /users/:userId` - Get user by ID
- [ ] `GET /users/search` - Search users
- [ ] `PUT /users/me/password` - Change password
- [ ] `PUT /users/me/settings` - Update user settings
- [ ] `GET /users/me/settings` - Get user settings
- [ ] `POST /users/me/avatar` - Upload profile avatar
- [ ] `DELETE /users/me/avatar` - Remove profile avatar


### Community Moderation
- [ ] `POST /communities/:communityId/moderators` - Add moderator
- [ ] `DELETE /communities/:communityId/moderators/:userId` - Remove moderator
- [ ] `GET /communities/:communityId/moderators` - List moderators
- [ ] `POST /communities/:communityId/rules` - Add community rule
- [ ] `PUT /communities/:communityId/rules/:ruleId` - Update community rule
- [ ] `DELETE /communities/:communityId/rules/:ruleId` - Delete community rule
- [ ] `GET /communities/:communityId/rules` - Get community rules
- [ ] `POST /communities/:communityId/ban/:userId` - Ban user from community
- [ ] `DELETE /communities/:communityId/ban/:userId` - Unban user from community
- [ ] `GET /communities/:communityId/banned` - List banned users

### Comments
- [ ] `GET /posts/:postId/comments` - Get comments for post
- [ ] `POST /posts/:postId/comments` - Add comment to post
- [ ] `GET /comments/:commentId` - Get specific comment
- [ ] `PUT /comments/:commentId` - Update comment
- [ ] `DELETE /comments/:commentId` - Delete comment
- [ ] `POST /comments/:commentId/reply` - Reply to comment
- [ ] `GET /comments/:commentId/replies` - Get replies to comment
- [ ] `POST /comments/:commentId/report` - Report a comment

### Votes
- [ ] `POST /posts/:postId/upvote` - Upvote a post
- [ ] `POST /posts/:postId/downvote` - Downvote a post
- [ ] `DELETE /posts/:postId/vote` - Remove vote from post
- [ ] `POST /comments/:commentId/upvote` - Upvote a comment
- [ ] `POST /comments/:commentId/downvote` - Downvote a comment
- [ ] `DELETE /comments/:commentId/vote` - Remove vote from comment
- [ ] `GET /users/me/voted` - Get posts/comments voted by current user

### Messages
- [ ] `GET /messages` - Get user conversations
- [ ] `GET /messages/:conversationId` - Get messages in conversation
- [ ] `POST /messages/:userId` - Start/continue conversation with user
- [ ] `DELETE /messages/:messageId` - Delete message
- [ ] `PUT /messages/:messageId/read` - Mark message as read
- [ ] `POST /messages/read-all` - Mark all messages as read

### Notifications
- [ ] `GET /notifications` - Get user notifications
- [ ] `PUT /notifications/:notificationId/read` - Mark notification as read
- [ ] `POST /notifications/read-all` - Mark all notifications as read
- [ ] `DELETE /notifications/:notificationId` - Delete notification
- [ ] `PUT /users/me/notification-settings` - Update notification preferences
- [ ] `GET /users/me/notification-settings` - Get notification preferences

### Search
- [ ] `GET /search` - Global search across posts, comments, communities, and users
- [ ] `GET /search/posts` - Search posts
- [ ] `GET /search/comments` - Search comments
- [ ] `GET /search/communities` - Search communities
- [ ] `GET /search/users` - Search users
- [ ] `GET /search/trending` - Get trending search terms

### Hashtags and Topics
- [ ] `GET /hashtags/trending` - Get trending hashtags
- [ ] `GET /hashtags/:tag/posts` - Get posts with specific hashtag
- [ ] `GET /topics` - Get list of topics
- [ ] `GET /topics/:topicId/posts` - Get posts for specific topic
- [ ] `POST /users/me/topics` - Follow topics
- [ ] `DELETE /users/me/topics/:topicId` - Unfollow topic
- [ ] `GET /users/me/topics` - Get topics followed by user

### Bookmarks
- [ ] `GET /users/me/bookmarks` - Get bookmarked posts
- [ ] `POST /posts/:postId/bookmark` - Bookmark a post
- [ ] `DELETE /posts/:postId/bookmark` - Remove bookmark

### Content Discovery
- [ ] `GET /discover/feed` - Get personalized discovery feed
- [ ] `GET /discover/trending` - Get trending content across platform
- [ ] `GET /discover/recommended/users` - Get recommended users to follow
- [ ] `GET /discover/recommended/communities` - Get recommended communities

### Following/Followers
- [ ] `GET /users/:userId/followers` - Get user followers
- [ ] `GET /users/:userId/following` - Get users followed by user
- [ ] `POST /users/:userId/follow` - Follow a user
- [ ] `DELETE /users/:userId/follow` - Unfollow a user

### Content Moderation
- [ ] `POST /moderation/reports` - Report content (admin endpoint)
- [ ] `GET /moderation/reports` - Get reported content (admin endpoint)
- [ ] `PUT /moderation/reports/:reportId` - Process report (admin endpoint)
- [ ] `POST /users/:userId/ban` - Ban user from platform (admin endpoint)
- [ ] `DELETE /users/:userId/ban` - Unban user (admin endpoint)
- [ ] `GET /moderation/banned` - List banned users (admin endpoint)

### Analytics (Admin/Owner Only)
- [ ] `GET /analytics/overview` - Get platform overview statistics
- [ ] `GET /analytics/users` - Get user statistics
- [ ] `GET /analytics/content` - Get content statistics
- [ ] `GET /analytics/engagement` - Get engagement statistics
- [ ] `GET /analytics/growth` - Get platform growth statistics
- [ ] `GET /analytics/communities` - Get community statistics

### System
- [ ] `GET /system/status` - API health check
- [ ] `GET /system/config` - Get public system configuration
- [ ] `POST /system/feedback` - Submit system feedback

### Media
- [ ] `POST /media/upload` - Upload media files
- [ ] `GET /media/:mediaId` - Get media file metadata
- [ ] `DELETE /media/:mediaId` - Delete uploaded media

### Feeds
- [ ] `GET /feeds/home` - Get personalized home feed
- [ ] `GET /feeds/all` - Get all content feed
- [ ] `GET /feeds/popular` - Get popular content feed
- [ ] `GET /feeds/community/:communityId` - Get community feed