package services

import (
	"log/slog"
	"net/http"

	"GophKeeper.ru/internal/entities"
	"GophKeeper.ru/internal/server/storage"
	"github.com/gin-gonic/gin"
)

func AccessData(group *gin.RouterGroup, db *storage.Database) {
	accessData(group.Group("/data"), db)
}

func accessData(group *gin.RouterGroup, mngr entities.DataManager) {
	group.POST("", func(ctx *gin.Context) {
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			slog.Error("user_id not found", "method", "accessData::POST")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		record := entities.Record{}
		err := ctx.BindJSON(&record)
		if err != nil {
			slog.Error("unmarshal to record: "+err.Error(), "method", "accessData::POST")
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Передаем контекст из Gin в UpdateRecord
		err = mngr.UpdateRecord(ctx.Request.Context(), userID, record)
		if err != nil {
			slog.Error("database error: "+err.Error(), "method", "accessData::POST")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	group.GET("", func(ctx *gin.Context) {
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			slog.Error("user_id not found", "method", "accessData::GET")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		countUpdate := ctx.GetInt("update")

		// Получаем update_id с контекстом
		currentUpdate, err := mngr.GetCountUpdate(ctx.Request.Context(), userID)
		if err != nil {
			slog.Error("GetCountUpdate error: "+err.Error(), "method", "accessData::GET")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if countUpdate == currentUpdate {
			ctx.Status(http.StatusNoContent)
			return
		}

		// Получаем данные пользователя с контекстом
		data, err := mngr.GetData(ctx.Request.Context(), userID)
		if err != nil {
			slog.Error("GetData error: "+err.Error(), "method", "accessData::GET")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		data.Value = currentUpdate // Обновляем значение update

		ctx.JSON(http.StatusOK, data)
	})

	group.DELETE("", func(ctx *gin.Context) {
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			slog.Error("user_id not found", "method", "accessData::DELETE")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		record := entities.Record{}
		err := ctx.BindJSON(&record)
		if err != nil {
			slog.Error("unmarshal to record: "+err.Error(), "method", "accessData::DELETE")
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Удаляем запись с использованием контекста
		err = mngr.RemoveRecord(ctx.Request.Context(), userID, record.Key)
		if err != nil {
			slog.Error("RemoveRecord error: "+err.Error(), "method", "accessData::DELETE")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusOK)
	})
}
