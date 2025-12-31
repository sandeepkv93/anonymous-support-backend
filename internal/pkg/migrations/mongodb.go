package migrations

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Migration represents a single MongoDB migration
type Migration struct {
	Version     int
	Description string
	Up          func(context.Context, *mongo.Database) error
	Down        func(context.Context, *mongo.Database) error
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version     int       `bson:"version"`
	Description string    `bson:"description"`
	AppliedAt   time.Time `bson:"applied_at"`
}

// MongoMigrator manages MongoDB migrations
type MongoMigrator struct {
	db         *mongo.Database
	logger     *zap.Logger
	migrations []Migration
}

// NewMongoMigrator creates a new MongoDB migrator
func NewMongoMigrator(db *mongo.Database, logger *zap.Logger) *MongoMigrator {
	return &MongoMigrator{
		db:         db,
		logger:     logger,
		migrations: []Migration{},
	}
}

// Register registers a migration
func (m *MongoMigrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// ensureMigrationsCollection ensures the migrations collection exists
func (m *MongoMigrator) ensureMigrationsCollection(ctx context.Context) error {
	collection := m.db.Collection("schema_migrations")

	// Create unique index on version
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// getAppliedVersions returns a set of applied migration versions
func (m *MongoMigrator) getAppliedVersions(ctx context.Context) (map[int]bool, error) {
	collection := m.db.Collection("schema_migrations")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	applied := make(map[int]bool)
	for cursor.Next(ctx) {
		var record MigrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		applied[record.Version] = true
	}

	return applied, nil
}

// recordMigration records a migration as applied
func (m *MongoMigrator) recordMigration(ctx context.Context, migration Migration) error {
	collection := m.db.Collection("schema_migrations")
	record := MigrationRecord{
		Version:     migration.Version,
		Description: migration.Description,
		AppliedAt:   time.Now(),
	}
	_, err := collection.InsertOne(ctx, record)
	return err
}

// removeMigrationRecord removes a migration record
func (m *MongoMigrator) removeMigrationRecord(ctx context.Context, version int) error {
	collection := m.db.Collection("schema_migrations")
	_, err := collection.DeleteOne(ctx, bson.M{"version": version})
	return err
}

// Up applies all pending migrations
func (m *MongoMigrator) Up(ctx context.Context) error {
	m.logger.Info("Starting MongoDB migrations")

	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations collection: %w", err)
	}

	// Get applied versions
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Apply pending migrations
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			m.logger.Info("Skipping already applied migration",
				zap.Int("version", migration.Version),
				zap.String("description", migration.Description))
			continue
		}

		m.logger.Info("Applying migration",
			zap.Int("version", migration.Version),
			zap.String("description", migration.Description))

		// Apply the migration
		if err := migration.Up(ctx, m.db); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Record the migration
		if err := m.recordMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		m.logger.Info("Successfully applied migration",
			zap.Int("version", migration.Version))
	}

	m.logger.Info("All MongoDB migrations completed successfully")
	return nil
}

// Down rolls back the last applied migration
func (m *MongoMigrator) Down(ctx context.Context) error {
	m.logger.Info("Rolling back last MongoDB migration")

	// Get applied versions
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to roll back")
		return nil
	}

	// Find the last applied migration
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if !applied[migration.Version] {
			continue
		}

		m.logger.Info("Rolling back migration",
			zap.Int("version", migration.Version),
			zap.String("description", migration.Description))

		// Roll back the migration
		if err := migration.Down(ctx, m.db); err != nil {
			return fmt.Errorf("failed to roll back migration %d: %w", migration.Version, err)
		}

		// Remove the migration record
		if err := m.removeMigrationRecord(ctx, migration.Version); err != nil {
			return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
		}

		m.logger.Info("Successfully rolled back migration",
			zap.Int("version", migration.Version))
		return nil
	}

	return nil
}

// Status returns the current migration status
func (m *MongoMigrator) Status(ctx context.Context) error {
	// Get applied versions
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	m.logger.Info("MongoDB migration status")
	for _, migration := range m.migrations {
		status := "pending"
		if applied[migration.Version] {
			status = "applied"
		}
		m.logger.Info("Migration",
			zap.Int("version", migration.Version),
			zap.String("description", migration.Description),
			zap.String("status", status))
	}

	return nil
}
