package app

import (
	"fmt"
	"os"
	"strings"
	domainModel "tictactoe/internal/domain/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessToken  = "access"
	refreshToken = "refresh"

	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type tokenClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type jwtProvider struct {
	secret []byte
}

func NewJWTProvider() JWTProvider {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "school21_default_secret_if_getenv_fails"
	}
	return &jwtProvider{secret: []byte(secret)}
}

func (j *jwtProvider) GenerateAccessToken(user domainModel.User) (string, error) {
	return j.generateToken(user.ID, accessToken, accessTokenTTL)
}

func (j *jwtProvider) GenerateRefreshToken(user domainModel.User) (string, error) {
	return j.generateToken(user.ID, refreshToken, refreshTokenTTL)
}
func (j *jwtProvider) ValidateAccessToken(token string) error {
	claims, err := j.parseToken(token)
	if err != nil {
		return err
	}
	if claims.TokenType != accessToken {
		return ErrInvalidToken
	}
	return nil
}

func (j *jwtProvider) ValidateRefreshToken(token string) error {
	claims, err := j.parseToken(token)
	if err != nil {
		return err
	}
	if claims.TokenType != refreshToken {
		return ErrInvalidToken
	}
	return nil
}
func (j *jwtProvider) ExtractUUIDfromToken(token string) (uuid.UUID, error) {
	claims, err := j.parseToken(token)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: invalid uuid in sub", ErrInvalidToken)
	}
	return id, nil
}

func (j *jwtProvider) generateToken(userID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign %s token: %w", tokenType, err)
	}
	return signed, nil
}

func (j *jwtProvider) parseToken(tokenStr string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
