package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/encryption"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
	"github.com/yourorg/anonymous-support/internal/repository/redis"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *postgres.UserRepository
	sessionRepo *redis.SessionRepository
	jwtManager  *jwt.Manager
	encManager  *encryption.Manager
}

func NewAuthService(
	userRepo *postgres.UserRepository,
	sessionRepo *redis.SessionRepository,
	jwtManager *jwt.Manager,
	encManager *encryption.Manager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
		encManager:  encManager,
	}
}

func (s *AuthService) RegisterAnonymous(ctx context.Context, username string, avatarID int) (*domain.User, string, string, error) {
	if err := validator.ValidateUsername(username); err != nil {
		return nil, "", "", err
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", fmt.Errorf("username already exists")
	}

	user := &domain.User{
		ID:          uuid.New(),
		Username:    username,
		AvatarID:    avatarID,
		IsAnonymous: true,
		IsBanned:    false,
		IsPremium:   false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, "", "", err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RegisterWithEmail(ctx context.Context, username, email, password string, avatarID int) (*domain.User, string, string, error) {
	if err := validator.ValidateUsername(username); err != nil {
		return nil, "", "", err
	}
	if err := validator.ValidateEmail(email); err != nil {
		return nil, "", "", err
	}
	if err := validator.ValidatePassword(password); err != nil {
		return nil, "", "", err
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", fmt.Errorf("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	encryptedEmail, err := s.encManager.Encrypt(email)
	if err != nil {
		return nil, "", "", err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        &encryptedEmail,
		PasswordHash: string(hashedPassword),
		AvatarID:     avatarID,
		IsAnonymous:  false,
		IsBanned:     false,
		IsPremium:    false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, "", "", err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*domain.User, string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	if user.IsAnonymous {
		return nil, "", "", fmt.Errorf("anonymous users cannot login with password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, "", "", err
	}

	if err := s.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, 168*3600*1000000000); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	storedToken, err := s.sessionRepo.GetRefreshToken(ctx, userID)
	if err != nil || storedToken != refreshToken {
		return "", fmt.Errorf("invalid refresh token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	return s.sessionRepo.DeleteRefreshToken(ctx, userID)
}
