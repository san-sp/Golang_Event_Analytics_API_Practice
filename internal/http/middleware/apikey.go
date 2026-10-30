package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"pulsetrack/internal/apikey"
)

const tenantIDContextKey = "tenantID"

// APIKeyAuth validates API keys and stores the authenticated tenant ID
// in the Gin context for downstream handlers.
func APIKeyAuth(service *apikey.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.Fields(header)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization header",
			})
			return
		}

		tenantID, err := service.Validate(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Set(tenantIDContextKey, tenantID)
		c.Next()
	}
}

// TenantIDFromContext retrieves the authenticated tenant ID.
func TenantIDFromContext(c *gin.Context) (bson.ObjectID, bool) {
	value, exists := c.Get(tenantIDContextKey)
	if !exists {
		return bson.NilObjectID, false
	}

	id, ok := value.(bson.ObjectID)
	if !ok || id.IsZero() {
		return bson.NilObjectID, false
	}

	return id, true
}
