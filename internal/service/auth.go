package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"devflow/internal/auth"
	"devflow/internal/model"
	"devflow/internal/repository"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrUsernameTaken     = errors.New("username already exists")
	ErrInvalidCredential = errors.New("invalid username or password")
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
}

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type LoginInput struct {
	Username string
	Password string
}

type AuthResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func NewAuthService(users *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	nickname := strings.TrimSpace(input.Nickname)
	if username == "" || password == "" || nickname == "" {
		return nil, ErrInvalidInput
	}
	if len(username) < 3 || len(username) > 64 || len(password) < 6 || len(nickname) > 64 {
		return nil, ErrInvalidInput
	}

	if _, err := s.users.FindByUsername(ctx, username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Nickname:     nickname,
		Status:       1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}
	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredential
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Me(ctx context.Context, userID uint64) (*model.User, error) {
	return s.users.FindByID(ctx, userID)
}
