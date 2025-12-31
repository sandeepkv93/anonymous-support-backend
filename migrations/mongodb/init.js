// Posts collection
db.createCollection("posts", {
    validator: {
        $jsonSchema: {
            required: ["user_id", "username", "type", "content", "created_at"],
            properties: {
                urgency_level: { minimum: 1, maximum: 10 }
            }
        }
    }
});

db.posts.createIndex({ user_id: 1, created_at: -1 });
db.posts.createIndex({ type: 1, created_at: -1 });
db.posts.createIndex({ categories: 1, created_at: -1 });
db.posts.createIndex({ created_at: -1 });
db.posts.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });

// Support responses collection
db.createCollection("support_responses");
db.support_responses.createIndex({ post_id: 1, created_at: -1 });
db.support_responses.createIndex({ user_id: 1, created_at: -1 });

// User trackers collection
db.createCollection("user_trackers");
db.user_trackers.createIndex({ user_id: 1 }, { unique: true });

print("MongoDB collections and indexes created successfully");
