package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/config"
	"github.com/yourorg/anonymous-support/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("Starting database seeding...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// TODO: Initialize database connections
	// db := initPostgres(cfg)
	// mongoDB := initMongo(cfg)
	// redis := initRedis(cfg)

	fmt.Println("Seeding users...")
	seedUsers(ctx)

	fmt.Println("Seeding circles...")
	seedCircles(ctx)

	fmt.Println("Seeding posts...")
	seedPosts(ctx)

	fmt.Println("Seeding complete!")
}

func seedUsers(ctx context.Context) {
	users := []struct {
		username string
		email    string
		password string
		role     domain.UserRole
	}{
		{"admin_user", "admin@example.com", "AdminPass123!", domain.RoleAdmin},
		{"moderator_user", "mod@example.com", "ModPass123!", domain.RoleModerator},
		{"test_user_1", "user1@example.com", "UserPass123!", domain.RoleUser},
		{"test_user_2", "user2@example.com", "UserPass123!", domain.RoleUser},
		{"anonymous_1", "", "", domain.RoleUser},
	}

	for _, u := range users {
		hashedPassword := ""
		if u.password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Failed to hash password for %s: %v", u.username, err)
				continue
			}
			hashedPassword = string(hash)
		}

		user := &domain.User{
			ID:             uuid.New(),
			Username:       u.username,
			Email:          u.email,
			PasswordHash:   hashedPassword,
			Role:           u.role,
			IsAnonymous:    u.email == "",
			StrengthPoints: 0,
			CreatedAt:      time.Now(),
			LastActiveAt:   time.Now(),
		}

		// TODO: Save to database
		// userRepo.Create(ctx, user)
		fmt.Printf("  - Created user: %s (%s)\n", user.Username, user.Role)
	}
}

func seedCircles(ctx context.Context) {
	circles := []struct {
		name        string
		description string
		category    string
		maxMembers  int32
	}{
		{"Daily Check-In", "Daily accountability and support", "general", 1000},
		{"Evening Warriors", "Support for evening triggers", "alcohol", 500},
		{"Early Recovery", "For those in their first 90 days", "general", 200},
		{"Long-term Sobriety", "For those with 1+ years sober", "milestone", 300},
	}

	for _, c := range circles {
		circle := &domain.Circle{
			ID:          uuid.New(),
			Name:        c.name,
			Description: c.description,
			Category:    c.category,
			MaxMembers:  c.maxMembers,
			MemberCount: 0,
			IsPrivate:   false,
			CreatedBy:   uuid.New(), // Should be actual user ID
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// TODO: Save to database
		// circleRepo.Create(ctx, circle)
		fmt.Printf("  - Created circle: %s\n", circle.Name)
	}
}

func seedPosts(ctx context.Context) {
	posts := []struct {
		postType     domain.PostType
		content      string
		categories   []string
		urgencyLevel int32
	}{
		{
			domain.PostTypeSOS,
			"Having strong cravings. Need support right now.",
			[]string{"alcohol", "cravings"},
			5,
		},
		{
			domain.PostTypeCheckIn,
			"Day 30! Feeling great and grateful for this community.",
			[]string{"milestone", "progress"},
			1,
		},
		{
			domain.PostTypeVictory,
			"Got through my first weekend sober in years!",
			[]string{"alcohol", "weekend"},
			1,
		},
		{
			domain.PostTypeQuestion,
			"How do you all handle social situations where everyone is drinking?",
			[]string{"social", "strategies"},
			2,
		},
	}

	for _, p := range posts {
		// TODO: Create posts in MongoDB
		// post := &domain.Post{...}
		// postRepo.Create(ctx, post)
		fmt.Printf("  - Created %s post: %s\n", p.postType, p.content[:30]+"...")
	}
}
