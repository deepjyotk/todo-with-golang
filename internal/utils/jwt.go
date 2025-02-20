// internal/utils/jwt.go
package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetUserIDFromContext extracts the user ID from the JWT claims stored in the Gin context.
// It returns the user ID as a uint or an error if it cannot be extracted.
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	// Retrieve the JWT claims from the context.
	jwtClaimsVal, exists := c.Get("jwt")
	if !exists {
		return 0, errors.New("user not authenticated")
	}

	// Assert that the claims are of type jwt.MapClaims.
	claims, ok := jwtClaimsVal.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid JWT claims")
	}

	// Extract the user_id from the claims.
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id not found in token")
	}

	return uint(userIDFloat), nil
}
