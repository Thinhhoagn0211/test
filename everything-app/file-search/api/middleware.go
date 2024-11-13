package api

import (
	"errors"
	"training/file-search/token"

	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func (server *Server) authMiddleware(tokenMaker token.Maker, enforcer *casbin.Enforcer) gin.HandlerFunc {
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

		var action string
		switch ctx.Request.Method {
		case http.MethodGet:
			action = "read"
		case http.MethodPost:
			action = "write"
		case http.MethodPut:
			action = "write"
		case http.MethodDelete:
			action = "write"
		}
		// Enforce the policy
		ok := enforcer.Enforce(server.role, ctx.FullPath(), action)
		if !ok {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
			ctx.Abort()
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
