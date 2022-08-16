package middleware

import (
	"log"
	"net/http"

	"book-keeper/helper"
	"book-keeper/service"

	"github.com/gin-gonic/gin"
)

//AuthorizeJWT validates the token user given, return 401 if not valid
func AuthorizeJWT(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response := helper.BuildErrorResponse("Failed to process request", "No token found", nil)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		metadata, _ := jwtService.ExtractTokenMetadata(c.Request)
		token, err := jwtService.ValidateToken(authHeader, metadata.TokenUuid)
		if token.Valid {
			c.Next()
		} else {
			log.Println(err)
			response := helper.BuildErrorResponse("Token is not valid", err.Error(), nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		}
	}
}
