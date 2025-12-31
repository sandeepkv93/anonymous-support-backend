package validator

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "john_doe", false},
		{"valid with numbers", "user123", false},
		{"too short", "ab", true},
		{"too long", "this_is_a_very_long_username_that_exceeds_the_limit_of_fifty_characters", true},
		{"invalid characters", "user@name", true},
		{"valid with hyphen", "user-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"invalid no @", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"invalid no tld", "user@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "password123", false},
		{"too short", "pass", true},
		{"exactly 8 chars", "12345678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePostContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"valid content", "This is a valid post", false},
		{"empty content", "", true},
		{"whitespace only", "   ", true},
		{"very long content", string(make([]byte, 5001)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePostContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
