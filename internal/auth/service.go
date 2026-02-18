package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	jwt         *JWTManager
	redisClient *redis.Client
}

func NewService(jwt *JWTManager, redisClient *redis.Client) *Service {
	return &Service{
		jwt:         jwt,
		redisClient: redisClient,
	}
}

func (s *Service) GenerateTokens(userID, email string) (*TokenPair, error) {
	pair, tokenID, err := s.jwt.GenerateTokenPair(userID, email)
	if err != nil {
		return nil, err
	}

	// Store refresh token ID in Redis
	key := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	err = s.redisClient.Set(context.Background(), key, "1", s.jwt.RefreshExpiry()).Err()
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return pair, nil
}

func (s *Service) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh token exists in Redis
	key := fmt.Sprintf("refresh:%s:%s", claims.UserID, claims.TokenID)
	exists, err := s.redisClient.Exists(context.Background(), key).Result()
	if err != nil {
		return nil, fmt.Errorf("checking refresh token: %w", err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("refresh token revoked")
	}

	// Revoke old refresh token
	s.redisClient.Del(context.Background(), key)

	// Generate new token pair
	// We need email from the original token - fetch from new generation
	pair, newTokenID, err := s.jwt.GenerateTokenPair(claims.UserID, "")
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newKey := fmt.Sprintf("refresh:%s:%s", claims.UserID, newTokenID)
	err = s.redisClient.Set(context.Background(), newKey, "1", s.jwt.RefreshExpiry()).Err()
	if err != nil {
		return nil, fmt.Errorf("storing new refresh token: %w", err)
	}

	return pair, nil
}

func (s *Service) Logout(userID string) error {
	// Delete all refresh tokens for this user
	pattern := fmt.Sprintf("refresh:%s:*", userID)
	iter := s.redisClient.Scan(context.Background(), 0, pattern, 100).Iterator()
	for iter.Next(context.Background()) {
		s.redisClient.Del(context.Background(), iter.Val())
	}
	return iter.Err()
}

func (s *Service) ValidateAccessToken(token string) (*AccessClaims, error) {
	return s.jwt.ValidateAccessToken(token)
}

// StoreRefreshTokenWithExpiry stores a refresh token with a specific TTL.
// Used by the handler when email is available.
func (s *Service) StoreRefreshToken(userID, tokenID string, expiry time.Duration) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	return s.redisClient.Set(context.Background(), key, "1", expiry).Err()
}

func (s *Service) JWT() *JWTManager {
	return s.jwt
}
