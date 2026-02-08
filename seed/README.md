# Database Seed Data

This directory contains comprehensive seed data for the Sync database with interconnected relationships across all collections.

## Quick Start

### Import All Seed Data
```bash
./scripts/import.sh
```

### Clear All Seed Data
```bash
./scripts/clear.sh
```

## Scripts

Located in [../scripts](../scripts/) directory:
- **import.sh** - Automated script to import all seed data in correct order
- **clear.sh** - Drop all seed data collections (with confirmation prompt)

Both scripts automatically read database credentials from `.env` file.

## Seed Files

- **users.json** - 6 test user accounts with complete profiles, avatars, and relationships
- **communities.json** - 3 communities (TechHub, DesignersUnite, OpenSourceHub) with settings and metadata
- **posts.json** - 10 posts across different communities with various content types
- **comments.json** - 10 comments on posts with nested threading support
- **community_interactions.json** - User memberships in communities
- **post_interactions.json** - Likes and saves on posts
- **comment_interactions.json** - Likes on comments
- **moderators.json** - 4 community moderators with permissions and stats
- **moderation_logs.json** - 10 moderation actions (removals, bans, pins, locks)
- **community_tags.json** - Community category tags

## Data Relationships

### Users
- **john_doe** (550e8400-e29b-41d4-a716-446655440001) - Active member, moderator of TechHub
- **jane_smith** (550e8400-e29b-41d4-a716-446655440002) - Designer, owner of DesignersUnite
- **test_user** (550e8400-e29b-41d4-a716-446655440003) - Test account
- **pratik** (550e8400-e29b-41d4-a716-446655440004) - Developer, owner of TechHub
- **alice_dev** (550e8400-e29b-41d4-a716-446655440005) - Developer, owner of OpenSourceHub
- **bob_inactive** (550e8400-e29b-41d4-a716-446655440006) - Inactive user

All users use password: **password123**
Bcrypt hash: `$2a$10$Cv/Xb2ykZ9FLmWyB6vaPEueAzA51kkU2GDZj8C4hwgAH3gQhwIo.q`

### Communities
- **TechHub** (comm-550e8400-e29b-41d4-a716-446655440001)
  - Owner: pratik
  - Members: 5 users
  - Posts: 6 posts
  - Moderators: pratik (owner), john_doe (moderator)

- **DesignersUnite** (comm-550e8400-e29b-41d4-a716-446655440002)
  - Owner: jane_smith
  - Members: 4 users
  - Posts: 2 posts
  - Moderators: jane_smith (owner)

- **OpenSourceHub** (comm-550e8400-e29b-41d4-a716-446655440003)
  - Owner: alice_dev
  - Members: 3 users
  - Posts: 2 posts
  - Moderators: alice_dev (owner)

## Manual Import (Optional)

If you prefer manual control, you can use `mongoimport` directly:

### Using mongoimport

```bash
# Navigate to the seed directory
cd seed

# Set your MongoDB URI
MONGO_URI="mongodb://username:password@localhost:27017/your-database"

# Import each collection
mongoimport --uri="$MONGO_URI" --collection=users --file=users.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=communities --file=communities.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=posts --file=posts.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=comments --file=comments.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=community_interactions --file=community_interactions.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=post_interactions --file=post_interactions.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=comment_interactions --file=comment_interactions.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=moderators --file=moderators.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=moderation_logs --file=moderation_logs.json --jsonArray --drop
mongoimport --uri="$MONGO_URI" --collection=community_tags --file=community_tags.json --jsonArray --drop
```

### Import Order (Important)

To maintain referential integrity, import in this order:
1. **users.json** - Base user accounts
2. **communities.json** - Communities (reference users as owners)
3. **community_interactions.json** - User memberships
4. **moderators.json** - Community moderators
5. **posts.json** - Posts (reference users and communities)
6. **comments.json** - Comments (reference posts and users)
7. **post_interactions.json** - Post likes/saves
8. **comment_interactions.json** - Comment likes
9. **moderation_logs.json** - Moderation actions
10. **community_tags.json** - Tags

## Schema Documentation

The schema for all collections is documented in:
- **schemas_output.md** - Human-readable markdown format
- **schemas_output.json** - Machine-readable JSON format

These files contain:
- Collection names and document counts
- Field types and nested object structures
- Index definitions with constraints (unique, sparse, TTL)

## Data Statistics

- **Total Users**: 6
- **Total Communities**: 3
- **Total Posts**: 10
- **Total Comments**: 10
- **Total Interactions**: 40 (community joins, post likes/saves, comment likes)
- **Total Moderators**: 4
- **Total Moderation Actions**: 10

## Features Demonstrated

This seed data demonstrates:
- ✅ User authentication and profiles
- ✅ Community creation and management
- ✅ Post creation (text, link, media types)
- ✅ Nested comment threading
- ✅ Social interactions (likes, saves, follows)
- ✅ Community memberships
- ✅ Moderation system (moderators, logs, actions)
- ✅ Content status (published, removed, locked, pinned)
- ✅ User relationships (follows, followers)
- ✅ Rich metadata (device info, location, timestamps)

## Testing Scenarios

Use this seed data to test:
- User login/authentication flows
- Community browsing and joining
- Post creation and interactions
- Comment threading and replies
- Moderation workflows
- User profiles and social features
- Search and filtering
- Permissions and authorization
- Activity feeds and notifications
