package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type contextKey string

const (
	InitDataKey contextKey = "init-data"
	TimezoneKey contextKey = "timezone"
)

// Returns new context with specified init data.
func withInitData(ctx context.Context, initData initdata.InitData) context.Context {
	return context.WithValue(ctx, InitDataKey, initData)
}

// Returns the init data from the specified context.
func CtxInitData(ctx context.Context) (initdata.InitData, bool) {
	initData, ok := ctx.Value(InitDataKey).(initdata.InitData)
	return initData, ok
}

// Returns new context with specified timezone.
func withTimezone(ctx context.Context, timezone string) context.Context {
	return context.WithValue(ctx, TimezoneKey, timezone)
}

// Returns the timezone from the specified context.
func CtxTimezone(ctx context.Context) (string, bool) {
	timezone, ok := ctx.Value(TimezoneKey).(string)
	return timezone, ok
}

// AuthMiddleware проверяет данные, полученные от Telegram Mini App
func AuthMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Включаем CORS
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Timezone")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		// Получаем timezone из заголовка
		timezone := c.GetHeader("X-Timezone")
		if timezone == "" {
			timezone = "UTC" // дефолтное значение
		}

		// Проверяем валидность timezone
		_, err := time.LoadLocation(timezone)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": "invalid timezone",
			})
			return
		}

		// Добавляем timezone в контекст
		c.Request = c.Request.WithContext(
			withTimezone(c.Request.Context(), timezone),
		)

		// Получаем данные из заголовка Authorization
		authParts := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authParts) != 2 {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "Unauthorized",
			})
			return
		}

		authType := authParts[0]
		authData := authParts[1]

		switch authType {
		case "tma":
			// Проверяем валидность данных
			if err := initdata.Validate(authData, token, time.Hour); err != nil {
				c.AbortWithStatusJSON(401, gin.H{
					"message": err.Error(),
				})
				return
			}

			// Парсим данные
			initData, err := initdata.Parse(authData)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{
					"message": err.Error(),
				})
				return
			}

			// Сохраняем данные в контекст
			c.Request = c.Request.WithContext(
				withInitData(c.Request.Context(), initData),
			)
		default:
			c.AbortWithStatusJSON(401, gin.H{
				"message": "Invalid authorization type",
			})
			return
		}

		c.Next()
	}
}
