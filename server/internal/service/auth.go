package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/dz-poker/server/internal/model"
	"github.com/dz-poker/server/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidAvatar      = errors.New("invalid avatar")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

const (
	JWTSecret     = "dz-poker-secret-key-change-in-production"
	JWTExpiration = 24 * 7 * time.Hour // 7 days
)

// AuthService handles authentication operations
type AuthService struct{}

// NewAuthService creates a new AuthService
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register creates a new user account
func (s *AuthService) Register(req *model.RegisterReq) (*model.RegisterResp, error) {
	// Validate avatar
	if !repository.ValidateAvatar(req.Avatar) {
		return nil, ErrInvalidAvatar
	}

	// Check if username exists
	existing, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := repository.CreateUser(req.Username, string(hashedPassword), req.Nickname, req.Avatar)
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.RegisterResp{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Chips:    user.Chips,
		Token:    token,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req *model.LoginReq) (*model.LoginResp, error) {
	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResp{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Chips:    user.Chips,
		Token:    token,
	}, nil
}

// ValidateToken validates a JWT token and returns the user
func (s *AuthService) ValidateToken(tokenString string) (*model.AuthResult, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	user, stats, err := repository.GetUserWithStats(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &model.AuthResult{
		User:  user,
		Stats: stats,
	}, nil
}

// GetUserInfo retrieves user information with stats
func (s *AuthService) GetUserInfo(userID string) (*model.UserInfoResp, error) {
	user, stats, err := repository.GetUserWithStats(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	resp := &model.UserInfoResp{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Chips:    user.Chips,
	}

	if stats != nil {
		winRate := "0%"
		if stats.TotalGames > 0 {
			winRate = fmt.Sprintf("%.1f%%", float64(stats.Wins)/float64(stats.TotalGames)*100)
		}
		resp.Stats = &model.StatsResp{
			TotalGames:  stats.TotalGames,
			Wins:        stats.Wins,
			WinRate:     winRate,
			TotalProfit: stats.TotalProfit,
		}
	}

	return resp, nil
}

// GetUserGameHistory retrieves user's game history
func (s *AuthService) GetUserGameHistory(userID string, limit int) (*model.GameHistoryResp, error) {
	histories, err := repository.GetUserGameHistory(userID, limit)
	if err != nil {
		return nil, err
	}

	total, err := repository.GetUserGameHistoryCount(userID)
	if err != nil {
		return nil, err
	}

	return &model.GameHistoryResp{
		Total: total,
		Games: histories,
	}, nil
}

// UpdateUserChips updates user's chips
func (s *AuthService) UpdateUserChips(userID string, chips int64) error {
	return repository.UpdateUserChips(userID, chips)
}

// GetAllUsers retrieves all users (admin only)
func (s *AuthService) GetAllUsers() ([]*model.User, error) {
	return repository.GetAllUsers()
}

// IsAdmin checks if user is admin (huahuo)
func (s *AuthService) IsAdmin(userID string) bool {
	user, err := repository.GetUserByID(userID)
	if err != nil || user == nil {
		return false
	}
	return user.Username == "huahuo"
}

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(JWTExpiration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}
