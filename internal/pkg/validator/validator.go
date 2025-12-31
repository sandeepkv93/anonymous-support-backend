package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}
	return nil
}

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

func ValidatePostContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("post content cannot be empty")
	}
	if len(content) > 5000 {
		return fmt.Errorf("post content cannot exceed 5000 characters")
	}
	return nil
}

func ValidateResponseContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("response content cannot be empty")
	}
	if len(content) > 2000 {
		return fmt.Errorf("response content cannot exceed 2000 characters")
	}
	return nil
}
