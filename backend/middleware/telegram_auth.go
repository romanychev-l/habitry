package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// TelegramAuthMiddleware проверяет данные, полученные от Telegram Mini App
func TelegramAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Включаем CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Data")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Пропускаем проверку аутентификации для публичного профиля
		if r.URL.Path == "/api/user/profile" && r.Method == "GET" {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем данные из заголовка
		initDataStr := r.Header.Get("X-Telegram-Data")
		if initDataStr == "" {
			http.Error(w, "Отсутствуют данные Telegram", http.StatusUnauthorized)
			return
		}

		log.Printf("Полученные данные: %s", initDataStr)

		// Парсим данные с помощью библиотеки
		// data, err := initdata.Parse(initDataStr)
		// if err != nil {
		// 	log.Printf("Ошибка при парсинге данных: %v\nПолученные данные: %s", err, initDataStr)
		// 	http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		// 	return
		// }

		// log.Printf("Успешно распарсены данные: %+v", data)

		// Проверяем подпись с помощью библиотеки
		botToken := os.Getenv("BOT_TOKEN")
		if botToken == "" {
			log.Printf("Отсутствует токен бота")
			http.Error(w, "Ошибка конфигурации сервера", http.StatusInternalServerError)
			return
		}

		// Устанавливаем срок действия данных (24 часа)
		expIn := 24 * time.Hour

		// Проверяем валидность данных
		if err := initdata.Validate(initDataStr, botToken, expIn); err != nil {
			log.Printf("Ошибка валидации данных: %v\nТокен бота (первые 10 символов): %s...\nСрок действия: %v",
				err, botToken[:10], expIn)
			http.Error(w, "Неверная подпись или устаревшие данные", http.StatusUnauthorized)
			return
		}

		log.Printf("Данные успешно провалидированы")

		// Если есть данные пользователя, добавляем telegram_id в контекст
		// if data.User.ID != 0 {
		// 	ctx := context.WithValue(r.Context(), "telegram_id", data.User.ID)
		// 	r = r.WithContext(ctx)
		// 	log.Printf("Установлен telegram_id в контексте: %d", data.User.ID)
		// } else {
		// 	// Пробуем получить telegram_id из query параметров
		// 	telegramIDStr := r.URL.Query().Get("telegram_id")
		// 	if telegramIDStr != "" {
		// 		if id, err := strconv.ParseInt(telegramIDStr, 10, 64); err == nil {
		// 			ctx := context.WithValue(r.Context(), "telegram_id", id)
		// 			r = r.WithContext(ctx)
		// 			log.Printf("Установлен telegram_id из query: %d", id)
		// 		}
		// 	}
		// }
		telegramIDStr := r.URL.Query().Get("telegram_id")
		if telegramIDStr != "" {
			if id, err := strconv.ParseInt(telegramIDStr, 10, 64); err == nil {
				ctx := context.WithValue(r.Context(), "telegram_id", id)
				r = r.WithContext(ctx)
			}
		}
		log.Printf("telegramIDStr: %s", telegramIDStr)
		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	}
}
