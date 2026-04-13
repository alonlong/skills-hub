package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret string) JWTIssuer {
	return JWTIssuer{
		secret: []byte(secret),
		ttl:    8 * time.Hour,
	}
}

func (i JWTIssuer) Issue(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.UserID,
		"username": user.Username,
		"roles":    user.PlatformRoles,
		"exp":      time.Now().Add(i.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
}

func (i JWTIssuer) Parse(token string) (Actor, error) {
	parsedToken, err := jwt.Parse(token, func(current *jwt.Token) (any, error) {
		if current.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", current.Method.Alg())
		}
		return i.secret, nil
	})
	if err != nil {
		return Actor{}, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return Actor{}, fmt.Errorf("invalid token claims")
	}

	return Actor{
		UserID:        stringClaim(claims["sub"]),
		Username:      stringClaim(claims["username"]),
		PlatformRoles: stringSliceClaim(claims["roles"]),
	}, nil
}

func stringClaim(value any) string {
	if current, ok := value.(string); ok {
		return current
	}
	return ""
}

func stringSliceClaim(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return []string{}
	}

	roles := make([]string, 0, len(items))
	for _, item := range items {
		if role, ok := item.(string); ok {
			roles = append(roles, role)
		}
	}
	return roles
}
