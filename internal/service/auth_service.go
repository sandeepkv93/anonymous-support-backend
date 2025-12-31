package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/dto"
	"github.com/yourorg/anonymous-support/internal/pkg/encryption"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"github.com/yourorg/anonymous-support/internal/pkg/metrics"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
	"github.com/yourorg/anonymous-support/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtManager  *jwt.Manager
	encManager  *encryption.Manager
	auditRepo   repository.AuditRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	jwtManager *jwt.Manager,
	encManager *encryption.Manager,
	auditRepo repository.AuditRepository,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
		encManager:  encManager,
		auditRepo:   auditRepo,
	}
}

func (s *AuthService) RegisterAnonymous(ctx context.Context, username string) (*dto.AuthResponse, error) {
	if err := validator.ValidateUsername(username); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	user := &domain.User{
		ID:          uuid.New(),
		Username:    username,
		AvatarID:    1, // Default avatar
		IsAnonymous: true,
		IsBanned:    false,
		IsPremium:   false,
		Role:        domain.RoleUser,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, err
	}

	// Emit metrics
	metrics.UsersRegisteredTotal.WithLabelValues("anonymous").Inc()
	metrics.AuthAttemptsTotal.WithLabelValues("anonymous_register", "success").Inc()

	return &dto.AuthResponse{
		User:         dto.NewUserDTO(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RegisterWithEmail(ctx context.Context, req *dto.RegisterWithEmailRequest) (*dto.AuthResponse, error) {
	if err := validator.ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validator.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	encryptedEmail, err := s.encManager.Encrypt(req.Email)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        &encryptedEmail,
		PasswordHash: string(hashedPassword),
		AvatarID:     1, // Default avatar
		IsAnonymous:  false,
		IsBanned:     false,
		IsPremium:    false,
		Role:         domain.RoleUser,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, err
	}

	userDTO := dto.NewUserDTO(user)
	userDTO.Email = req.Email // Override with plaintext email

	return &dto.AuthResponse{
		User:         userDTO,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Login uses email, so we need to find user by email
	// For now, check if email field contains @ (email) or not (username fallback)
	var user *domain.User
	var err error

	if len(req.Email) > 0 && req.Email[0] != '@' {
		// Attempt to find by email
		user, err = s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			// Fallback: try username
			user, err = s.userRepo.GetByUsername(ctx, req.Email)
		}
	} else {
		user, err = s.userRepo.GetByUsername(ctx, req.Email)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.IsAnonymous {
		return nil, fmt.Errorf("anonymous users cannot login with password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, err
	}

	// Decrypt email if exists
	userDTO := dto.NewUserDTO(user)
	if user.Email != nil {
		if email, err := s.encManager.Decrypt(*user.Email); err == nil {
			userDTO.Email = email
		}
	}

	return &dto.AuthResponse{
		User:         userDTO,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// 1. Validate the refresh token JWT signature and expiry
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Check if token is in the active token set (not revoked)
	isValid, err := s.sessionRepo.ValidateRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !isValid {
		// Token reuse detected! This could be a token theft attempt.
		// Revoke all refresh tokens for this user as a security measure.
		s.sessionRepo.RevokeAllRefreshTokens(ctx, userID)
		return nil, fmt.Errorf("token reuse detected, all sessions revoked for security")
	}

	// 3. Get user details
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Check if user is banned
	if user.IsBanned {
		return nil, fmt.Errorf("user is banned")
	}

	// 4. Generate new token pair (rotation)
	newAccessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	// 5. Revoke the old refresh token
	if err := s.sessionRepo.RevokeRefreshToken(ctx, userID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// 6. Store the new refresh token
	if err := s.sessionRepo.StoreRefreshToken(ctx, userID, newRefreshToken, 168*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store new token: %w", err)
	}

	// 7. Return the new token pair
	userDTO := dto.NewUserDTO(user)
	if user.Email != nil {
		if email, err := s.encManager.Decrypt(*user.Email); err == nil {
			userDTO.Email = email
		}
	}

	return &dto.AuthResponse{
		User:         userDTO,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteRefreshToken(ctx, userID.String())
}

func (s *AuthService) HandleOAuthLogin(ctx context.Context, provider, providerUserID, email, name string) (*dto.AuthResponse, error) {
	// Try to find existing user by email (OAuth accounts have verified emails)
	var user *domain.User
	var err error

	if email != "" {
		encryptedEmail, _ := s.encManager.Encrypt(email)
		user, err = s.userRepo.GetByEmail(ctx, encryptedEmail)
	}

	// If user doesn't exist, create a new account
	if err != nil || user == nil {
		user = &domain.User{
			ID:           uuid.New(),
			Username:     generateUsernameFromEmail(email),
			Email:        &email,
			IsAnonymous:  false,
			Role:         domain.RoleUser,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		}

		// Encrypt email before storing
		if email != "" {
			encryptedEmail, err := s.encManager.Encrypt(email)
			if err != nil {
				return nil, err
			}
			user.Email = &encryptedEmail
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		metrics.AuthAttemptsTotal.WithLabelValues("oauth_register", provider, "success").Inc()
	} else {
		metrics.AuthAttemptsTotal.WithLabelValues("oauth_login", provider, "success").Inc()
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*time.Hour); err != nil {
		return nil, err
	}

	userDTO := dto.NewUserDTO(user)
	userDTO.Email = email // Use plaintext email

	expiresIn := int64(3600) // 1 hour in seconds

	return &dto.AuthResponse{
		User:         userDTO,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func generateUsernameFromEmail(email string) string {
	if email == "" {
		return "user_" + uuid.New().String()[:8]
	}

	// Take part before @ and add random suffix
	parts := string(email)
	atIndex := 0
	for i, c := range parts {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex > 0 {
		return parts[:atIndex] + "_" + uuid.New().String()[:4]
	}

	return "user_" + uuid.New().String()[:8]
}
