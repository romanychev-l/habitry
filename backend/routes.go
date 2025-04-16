package main

import (
	"backend/handlers/follower"
	"backend/handlers/habit"
	"backend/handlers/invoice"
	"backend/handlers/ping"
	"backend/handlers/ton"
	"backend/handlers/user"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func setupGinRouter(
	userHandler *user.Handler,
	habitHandler *habit.Handler,
	invoiceHandler *invoice.Handler,
	followerHandler *follower.Handler,
	tonHandler *ton.TonHandler,
	pingHandler *ping.Handler,
	botToken string,
) *gin.Engine {
	// Создаем роутер без middleware
	r := gin.New()

	// Добавляем middleware вручную
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Применяем middleware аутентификации
	r.Use(middleware.AuthMiddleware(botToken))

	// Группа API
	api := r.Group("/api")
	{
		// Маршруты пользователя
		userGroup := api.Group("/user")
		{
			userGroup.POST("", userHandler.HandleUser)
			userGroup.PUT("/last_visit", userHandler.HandleUpdateLastVisit)
			userGroup.GET("/settings", userHandler.HandleSettings)
			userGroup.PUT("/settings", userHandler.HandleSettings)
			userGroup.GET("/profile", userHandler.HandleUserProfile)
		}

		// Маршруты привычек
		habitGroup := api.Group("/habit")
		{
			habitGroup.POST("/create", habitHandler.HandleCreate)
			habitGroup.PUT("/click", habitHandler.HandleUpdate)
			habitGroup.PUT("/edit", habitHandler.HandleEdit)
			habitGroup.DELETE("/delete", habitHandler.HandleDelete)
			habitGroup.PUT("/undo", habitHandler.HandleUndo)
			habitGroup.POST("/join", habitHandler.HandleJoin)
			habitGroup.GET("/followers", habitHandler.HandleGetFollowers)
			habitGroup.GET("/progress", followerHandler.HandleHabitProgress)
			habitGroup.GET("/activity", habitHandler.HandleGetActivity)
			habitGroup.POST("/unfollow", followerHandler.HandleUnfollow)
		}

		// Маршруты TON
		tonGroup := api.Group("/ton")
		{
			// tonGroup.POST("/deposit", tonHandler.HandleDeposit)
			// tonGroup.GET("/transaction", tonHandler.HandleCheckTransaction)
			tonGroup.POST("/usdt-deposit", tonHandler.HandleUsdtDeposit)
			tonGroup.POST("/check-usdt-transaction", tonHandler.HandleCheckUsdtTransaction)
			tonGroup.POST("/withdraw", tonHandler.HandleWithdraw)
		}

		// Маршруты пингов
		pingGroup := api.Group("/pings")
		{
			pingGroup.POST("/create", pingHandler.HandleCreatePing)
		}

		// Маршруты инвойсов
		invoiceGroup := api.Group("/invoice")
		{
			invoiceGroup.GET("", invoiceHandler.HandleCreateInvoice)
		}
	}

	return r
}
