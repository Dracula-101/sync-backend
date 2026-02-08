# Database Scripts

Utility scripts for managing seed data and database operations.

## Scripts

### seed/main.go

Go-based seed data management with subcommands.

**Direct Usage:**
```bash
# Import seed data
go run scripts/seed/main.go import

# Clear seed data
go run scripts/seed/main.go clear
```

**Shell Wrapper Scripts:**
```bash
# Import using shell wrapper
./scripts/import.sh

# Clear using shell wrapper
./scripts/clear.sh
```

### Features

**Import Command:**
- Automatically reads database credentials from `.env` file
- Imports collections in the correct order to maintain referential integrity
- Drops existing collections before importing (fresh start)
- Colored output with progress indicators
- Error handling for missing files
- Shows document count for each collection

**Clear Command:**
- Automatically reads database credentials from `.env` file
- Confirmation prompt before deletion (type "yes" to confirm)
- Shows which database and host will be affected
- Drops collections individually with status feedback
- Safe operation with explicit confirmation required

**Import Order:**
1. users
2. communities
3. community_interactions
4. moderators
5. posts
6. comments
7. post_interactions
8. comment_interactions
9. moderation_logs
10. community_tags

**Collections Managed:**
- users
- communities
- posts
- comments
- community_interactions
- post_interactions
- comment_interactions
- moderators
- moderation_logs
- community_tags

## Requirements

- **Go 1.21+** - Required to run the scripts
- **.env file** - Must exist in project root with database credentials

## Environment Variables

Both scripts expect these variables in `.env`:
```bash
DB_USER=your_username
DB_PASSWORD=your_password
DB_HOST=localhost:27017
DB_NAME=your_database_name
```

## Examples

### Import seed data
```bash
# From project root
./scripts/import.sh

# Output:
# 🌱 Starting seed data import...
# ═══════════════════════════════════════════════════
# 📦 Importing users...
# ✅ users imported successfully
# ...
# 🎉 Seed data import completed!
```

### Clear all seed data
```bash
# From project root
./scripts/clear.sh

# Output:
# ⚠️  WARNING: This will delete all seed data from the database!
# Database: test-db
# Host: localhost:27017
#
# Are you sure you want to continue? (yes/no): yes
# 🗑️  Clearing seed data...
# ...
# 🎉 Database cleared successfully!
```

## Troubleshooting

### "mongoimport: command not found"
Install MongoDB Database Tools:
```bash
# macOS
brew install mongodb-database-tools

# Ubuntu/Debian
sudo apt-get install mongodb-database-tools

# Windows
# Download from: https://www.mongodb.com/try/download/database-tools
```

### "mongosh: command not found"
Install MongoDB Shell:
```bash
# macOS
brew install mongosh

# Ubuntu/Debian
sudo apt-get install mongodb-mongosh

# Windows
# Download from: https://www.mongodb.com/try/download/shell
```

### ".env file not found"
Create a `.env` file in the project root with your database credentials:
```bash
cp .env.example .env
# Edit .env with your credentials
```

### "Authentication failed"
Verify your database credentials in `.env` are correct and that your MongoDB user has the necessary permissions.
