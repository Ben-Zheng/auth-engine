package common

import (
	"context"
	"net/http"
	"strings"

	"github.com/auth-engine/internal/pkg/models"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

const (
	AuthResultKey = "authorization"
)

func VerifyAuthorization() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			hlog.Errorf("Authorization header is missing.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Unauthorized: No authorization header provided."})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			hlog.Errorf("Invalid authorization token format.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.H{"error": "Unauthorized: No authorization header provided."})
			return
		}
		token := parts[1]

		// 解析JWT
		parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.H{"error": "Failed to parse JWT"})
			return
		}
		// 从token.Claims中获取sub和preferred_username字段
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.H{"error": "Failed to get token claims"})
			return
		}
		sub, _ := claims["sub"].(string)
		preferredUsername, _ := claims["preferred_username"].(string)
		// 将结果存储在上下文中
		c.Set(AuthResultKey, models.AuthResult{Sub: sub, PreferredUsername: preferredUsername})

		md := metadata.New(map[string]string{})
		md.Set("authorization", authHeader)
		ctx = metadata.NewIncomingContext(ctx, md)
		type DCETokenKey string
		ctx = context.WithValue(ctx, DCETokenKey("DceToken"), token)
		c.Next(ctx)
	}
}
