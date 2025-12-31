package domain

import "time"

// SoftDeletable interface for entities that support soft delete
type SoftDeletable interface {
	IsDeleted() bool
	GetDeletedAt() *time.Time
	MarkAsDeleted()
}

// SoftDeleteFields provides common soft delete fields
type SoftDeleteFields struct {
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the entity is soft deleted
func (s *SoftDeleteFields) IsDeleted() bool {
	return s.DeletedAt != nil
}

// GetDeletedAt returns the deletion timestamp
func (s *SoftDeleteFields) GetDeletedAt() *time.Time {
	return s.DeletedAt
}

// MarkAsDeleted marks the entity as deleted
func (s *SoftDeleteFields) MarkAsDeleted() {
	now := time.Now()
	s.DeletedAt = &now
}

// Restore removes the soft delete marker
func (s *SoftDeleteFields) Restore() {
	s.DeletedAt = nil
}
