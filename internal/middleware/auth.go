package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Define a context key for storing the user ID
type contextKey string

var UserNameKey = contextKey("user_name")
var UserIDKey = contextKey("user_id")

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the query parameter instead of the Authorization header
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token query parameter missing"})
			c.Abort()
			return
		}

		secretKey := os.Getenv("SECRET_KEY")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Store the user ID in the request context without modifying the key
			userName := claims["sub"].(string)
			userId := claims["id"].(string)
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, UserNameKey, userName)
			ctx = context.WithValue(ctx, UserIDKey, userId)
			c.Request = c.Request.WithContext(ctx)

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
		}
	}
}
