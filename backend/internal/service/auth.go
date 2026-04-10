package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
)

// ErrInvalidCredentials is returned when email/password don't match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrEmailTaken is returned when email already exists.
var ErrEmailTaken = errors.New("email already in use")

type AuthService struct {
	users      *repository.UserRepository
	jwtSecret  string
	bcryptCost int
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, bcryptCost int) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, bcryptCost: bcryptCost}
}

type AuthResult struct {
	Token string
	User  *model.User
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*AuthResult, error) {
	// Normalise email
	email = strings.ToLower(strings.TrimSpace(email))

	// Check for duplicate email
	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Persist
	user, err := s.users.Create(ctx, name, email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Issue JWT
	token, err := GenerateToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		// Don't reveal whether the email exists
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := GenerateToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{Token: token, User: user}, nil
}
