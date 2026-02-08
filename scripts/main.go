package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"sync-backend/arch/config"
	"sync-backend/arch/mongo"
	"sync-backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorBlue   = "\033[0;34m"
	ColorYellow = "\033[1;33m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "import":
		importData()
	case "clear":
		clearData()
	default:
		fmt.Printf("%sUnknown command: %s%s\n\n", ColorRed, command, ColorReset)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run scripts/main.go import   - Import seed data")
	fmt.Println("  go run scripts/main.go clear    - Clear seed data")
}

func importData() {
	fmt.Printf("%s🌱 Starting seed data import...%s\n\n", ColorBlue, ColorReset)

	// Load environment
	env := config.NewEnv(".env")

	// Create logger
	appLogger := utils.DefaultAppLogger("development", "info", "IMPORT")

	// Connect to database
	ctx := context.Background()
	db := connectDatabase(ctx, &env, appLogger)

	// Get database instance
	dbInstance := db.GetInstance()
	mongoDb := dbInstance.Database

	// Define import order
	imports := []struct {
		File       string
		Collection string
	}{
		{"seed/users.json", "users"},
		{"seed/communities.json", "communities"},
		{"seed/community_interactions.json", "community_interactions"},
		{"seed/moderators.json", "moderators"},
		{"seed/posts.json", "posts"},
		{"seed/comments.json", "comments"},
		{"seed/post_interactions.json", "post_interactions"},
		{"seed/comment_interactions.json", "comment_interactions"},
		{"seed/moderation_logs.json", "moderation_logs"},
		{"seed/community_tags.json", "community_tags"},
	}

	fmt.Printf("%s═══════════════════════════════════════════════════%s\n", ColorBlue, ColorReset)

	totalImported := 0
	totalFailed := 0

	for _, imp := range imports {
		imported, failed := importCollection(ctx, mongoDb, imp.File, imp.Collection)
		totalImported += imported
		totalFailed += failed
	}

	fmt.Printf("%s═══════════════════════════════════════════════════%s\n", ColorBlue, ColorReset)

	fmt.Printf("\n%s🎉 Seed data import completed!%s\n", ColorGreen, ColorReset)
	fmt.Printf("%s📊 Database: %s%s\n", ColorBlue, env.DBName, ColorReset)
	fmt.Printf("%s🔗 Host: %s%s\n\n", ColorBlue, env.DBHost, ColorReset)

	fmt.Printf("%s📋 Imported Collections:%s\n", ColorBlue, ColorReset)
	fmt.Println("   • users (6 accounts)")
	fmt.Println("   • communities (3 communities)")
	fmt.Println("   • posts (10 posts)")
	fmt.Println("   • comments (10 comments)")
	fmt.Println("   • community_interactions (12 memberships)")
	fmt.Println("   • post_interactions (18 likes/saves)")
	fmt.Println("   • comment_interactions (10 likes)")
	fmt.Println("   • moderators (4 moderators)")
	fmt.Println("   • moderation_logs (10 actions)")
	fmt.Println("   • community_tags (73 tags)")

	fmt.Printf("\n%s💡 All test accounts use password: password123%s\n", ColorGreen, ColorReset)
	fmt.Printf("\n%s📊 Total: %d documents imported, %d failed%s\n",
		ColorBlue, totalImported, totalFailed, ColorReset)
}

func clearData() {
	fmt.Printf("%s⚠️  WARNING: This will delete all seed data from the database!%s\n", ColorYellow, ColorReset)

	env := config.NewEnv(".env")

	fmt.Printf("%sDatabase: %s%s\n", ColorBlue, env.DBName, ColorReset)
	fmt.Printf("%sHost: %s%s\n\n", ColorBlue, env.DBHost, ColorReset)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to continue? (yes/no): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "yes" {
		fmt.Printf("%s❌ Operation cancelled%s\n", ColorBlue, ColorReset)
		os.Exit(0)
	}

	fmt.Printf("\n%s🗑️  Clearing seed data...%s\n\n", ColorRed, ColorReset)

	appLogger := utils.DefaultAppLogger("development", "info", "CLEAR")

	ctx := context.Background()
	db := connectDatabase(ctx, &env, appLogger)

	dbInstance := db.GetInstance()
	mongoDb := dbInstance.Database

	collections := []string{
		"users",
		"communities",
		"posts",
		"comments",
		"community_interactions",
		"post_interactions",
		"comment_interactions",
		"moderators",
		"moderation_logs",
		"community_tags",
	}

	for _, collName := range collections {
		fmt.Printf("%sDropping %s...%s\n", ColorBlue, collName, ColorReset)
		err := mongoDb.Collection(collName).Drop(ctx)
		if err != nil {
			fmt.Printf("%s❌ Failed to drop %s: %v%s\n", ColorRed, collName, err, ColorReset)
		} else {
			fmt.Printf("%s✅ %s dropped%s\n", ColorGreen, collName, ColorReset)
		}
	}

	fmt.Printf("\n%s🎉 Database cleared successfully!%s\n", ColorGreen, ColorReset)
}

func connectDatabase(ctx context.Context, env *config.Env, appLogger utils.AppLogger) mongo.Database {
	dbConfig := mongo.DbConfig{
		User:        env.DBUser,
		Pwd:         env.DBPassword,
		Host:        env.DBHost,
		Name:        env.DBName,
		MinPoolSize: 10,
		MaxPoolSize: 100,
		Timeout:     10 * time.Second,
	}
	db := mongo.NewDatabase(ctx, appLogger, dbConfig)
	db.Connect()
	return db
}

func importCollection(
	ctx context.Context,
	mongoDb *mongodriver.Database,
	filePath, collectionName string,
) (int, int) {

	fmt.Printf("%s📦 Importing %s...%s\n", ColorBlue, collectionName, ColorReset)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("%s⚠️  %s not found, skipping...%s\n\n", ColorRed, filePath, ColorReset)
		return 0, 0
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%s❌ Failed to read %s: %v%s\n\n", ColorRed, filePath, err, ColorReset)
		return 0, 0
	}

	var documents []bson.M
	if err := json.Unmarshal(fileData, &documents); err != nil {
		fmt.Printf("%s❌ Failed to parse %s: %v%s\n\n", ColorRed, filePath, err, ColorReset)
		return 0, 0
	}

	if len(documents) == 0 {
		fmt.Printf("%s⚠️  No documents found in %s%s\n\n", ColorRed, filePath, ColorReset)
		return 0, 0
	}

	// 🔹 ONLY ADDITION: convert date strings to Mongo Timestamp
	for _, doc := range documents {
		normalizeDatesToMongoTimestamp(doc)
	}

	coll := mongoDb.Collection(collectionName)
	_ = coll.Drop(ctx)

	docsInterface := make([]interface{}, len(documents))
	for i, doc := range documents {
		docsInterface[i] = doc
	}

	_, err = coll.InsertMany(ctx, docsInterface)
	if err != nil {
		fmt.Printf("%s❌ Failed to import %s: %v%s\n\n", ColorRed, collectionName, err, ColorReset)
		return 0, len(documents)
	}

	fmt.Printf("%s✅ %s imported successfully (%d documents)%s\n\n",
		ColorGreen, collectionName, len(documents), ColorReset)

	return len(documents), 0
}

// Converts RFC3339 date strings → MongoDB Timestamp
func normalizeDatesToMongoTimestamp(m bson.M) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				m[k] = primitive.DateTime(t.UnixMilli())
			}
		case bson.M:
			normalizeDatesToMongoTimestamp(val)
		case map[string]interface{}:
			normalizeDatesToMongoTimestamp(val)
		case []interface{}:
			for _, it := range val {
				if mm, ok := it.(bson.M); ok {
					normalizeDatesToMongoTimestamp(mm)
				}
			}
		}
	}
}
