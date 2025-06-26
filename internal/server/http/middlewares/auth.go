package middlewares

import (
	"log/slog"
	"net/http"

	"GophKeeper.ru/internal/entities"
	"GophKeeper.ru/internal/utils"
	"github.com/gin-gonic/gin"
)

// WarpAuth обертка авторизации
func WarpAuth(mngr entities.AuthManager) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		Auth(ctx, mngr)
	}
}

// Auth проверяет наличие токена, если он есть — пускает дальше
func Auth(ctx *gin.Context, mngr entities.AuthManager) {
	if ctx.Request.URL.Path == "/api/auth" {
		ctx.Next()
		return
	}

	token, err := ctx.Cookie("token")
	if err != nil {
		slog.Error("Failed to get token cookie", "error", err, "method", "middlewares.Auth")
		ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	IDUser, err := utils.LoginFromToken(token, "secret_key")
	if err != nil {
		slog.Error("Failed to parse token", "error", err, "method", "middlewares.Auth")
		ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	user, err := mngr.UserFromID(ctx.Request.Context(), IDUser)
	if err != nil || user.IsDisable {
		slog.Error("User not found or disabled", "error", err, "method", "middlewares.Auth")
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	}

	ctx.Set("user_id", IDUser)
	ctx.Next()
}
