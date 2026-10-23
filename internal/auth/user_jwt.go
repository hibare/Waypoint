package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hibare/Waypoint/internal/config"
	"github.com/hibare/Waypoint/internal/constants"
	"github.com/hibare/Waypoint/internal/db/users"
)

const (
	// JWT cookie name.
	JWTCookieName = "x-waypoint-auth"
)

var (
	// ErrFailedJWTParsing is returned when JWT parsing fails.
	ErrFailedJWTParsing = errors.New("failed to parse JWT token")

	// ErrInvalidJWTToken is returned when the JWT token is invalid.
	ErrInvalidJWTToken = errors.New("invalid JWT token")

	// ErrInvalidJWTClaims is returned when the JWT claims are invalid.
	ErrInvalidJWTClaims = errors.New("invalid JWT claims")
)

// UserJWTClaims represents the custom claims for our JWT tokens.
type UserJWTClaims struct {
	UserID     string   `json:"user_id"`
	UserEmail  string   `json:"user_email"`
	UserName   string   `json:"user_name"`
	UserGroups []string `json:"user_groups,omitempty"`
	jwt.RegisteredClaims
}

// CreateUserJWT creates a signed JWT token for the user with custom expiration.
func CreateUserJWT(user *users.User, expiry time.Time) (string, error) {
	claims := UserJWTClaims{
		UserID:     user.ID.String(),
		UserEmail:  user.Email,
		UserName:   strings.TrimSpace(user.FirstName + " " + user.LastName),
		UserGroups: user.Groups,
		ExpiresAt:  jwt.NewNumericDate(expiry),
		IssuedAt:   jwt.NewNumericDate(time.Now().UTC()),
		NotBefore:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:     constants.ProgramIdentifier,
		Subject:    user.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := []byte(config.Current.Core.SecretKey)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedJWTSigning, err)
	}

	return tokenString, nil
}

// VerifyUserJWT verifies and parses a JWT token.
func VerifyUserJWT(tokenString string) (*UserJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserJWTClaims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return []byte(config.Current.Core.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedJWTParsing, err)
	}

	if !token.Valid {
		return nil, ErrInvalidJWTToken
	}

	claims, ok := token.Claims.(*UserJWTClaims)
	if !ok {
		return nil, ErrInvalidJWTClaims
	}

	return claims, nil
}
