package api

import (
	"errors"
	"fmt"
	"training/11-file-search/token"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func (server *Server) authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format	")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := errors.New("authorization type is not bearer")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessibleRoles, err := server.enforcer.GetRolesForUser(payload.Role)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		fmt.Printf("User role: %s, Accessible roles: %v\n", payload.Role, accessibleRoles)

		if !hasPermission(payload.Role, accessibleRoles) {
			err := fmt.Errorf("forbidden: role %s is not allowed to access this API", payload.Role)
			ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(err))
		}
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func hasPermission(userRole string, accessibleRoles []string) bool {
	for _, role := range accessibleRoles {
		if role == userRole {
			return true
		}
	}
	return false
}
