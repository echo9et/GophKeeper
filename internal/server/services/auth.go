package services

import (
	"errors"
	"net/http"
	"time"

	"log/slog"

	"GophKeeper.ru/internal/entities"
	"github.com/gin-gonic/gin"
)

func Auth(group *gin.RouterGroup, mngr entities.AuthManager) {
	auth(group.Group("/auth"), mngr)
}

// auth производит аутификацию пользователя
func auth(group *gin.RouterGroup, mngr entities.AuthManager) {
	group.POST("", func(ctx *gin.Context) {
		var user entities.User

		if err := ctx.BindJSON(&user); err != nil {
			slog.Error("Failed to parse request body", "error", err, "method", "auth::POST")
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if user.Login == "" || user.Password == "" {
			slog.Warn("Empty login or password", "method", "auth::POST")
			ctx.AbortWithError(http.StatusBadRequest, errors.New("неверный формат запроса"))
			return
		}

		// Получаем пользователя из БД с использованием контекста
		u, err := mngr.User(ctx.Request.Context(), user.Login)
		if err != nil {
			slog.Error("Database error while fetching user", "error", err, "method", "auth::POST")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if u == nil || !u.IsEcual(&user) {
			slog.Warn("Invalid credentials", "login", user.Login, "method", "auth::POST")
			ctx.AbortWithError(http.StatusUnauthorized, errors.New("неверная пара логин/пароль"))
			return
		}

		token, err := entities.GetToken(u)
		if err != nil {
			slog.Error("Failed to generate token", "error", err, "method", "auth::POST")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Устанавливаем токен в виде cookie
		ctx.SetCookie("token", token, int(time.Hour), "/", "", false, true)
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
