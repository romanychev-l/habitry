package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
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

		// Декодируем URL-encoded строку
		decodedData, err := url.QueryUnescape(initDataStr)
		if err != nil {
			log.Printf("Ошибка при декодировании данных: %v", err)
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		log.Printf("Декодированные данные: %s", decodedData)

		// Парсим данные с помощью библиотеки
		data, err := initdata.Parse(decodedData)
		if err != nil {
			log.Printf("Ошибка при парсинге данных: %v", err)
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

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
		if err := initdata.Validate(decodedData, botToken, expIn); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			http.Error(w, "Неверная подпись или устаревшие данные", http.StatusUnauthorized)
			return
		}

		// Если есть данные пользователя, добавляем telegram_id в контекст
		if data.User.ID != 0 {
			telegramID := data.User.ID
			ctx := context.WithValue(r.Context(), "telegram_id", telegramID)
			r = r.WithContext(ctx)
			log.Printf("Установлен telegram_id в контексте: %d", telegramID)
		} else {
			// Пробуем получить telegram_id из query параметров
			telegramIDStr := r.URL.Query().Get("telegram_id")
			if telegramIDStr != "" {
				if id, err := strconv.ParseInt(telegramIDStr, 10, 64); err == nil {
					ctx := context.WithValue(r.Context(), "telegram_id", id)
					r = r.WithContext(ctx)
					log.Printf("Установлен telegram_id из query: %d", id)
				}
			}
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	}
}
