package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
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
		initData := r.Header.Get("X-Telegram-Data")
		if initData == "" {
			http.Error(w, "Отсутствуют данные Telegram", http.StatusUnauthorized)
			return
		}

		// Декодируем URL-encoded строку
		decodedData, err := url.QueryUnescape(initData)
		if err != nil {
			log.Printf("Ошибка при декодировании данных: %v", err)
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		// Парсим данные
		values, err := url.ParseQuery(decodedData)
		if err != nil {
			log.Printf("Ошибка при парсинге данных: %v", err)
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		// Получаем хеш из данных
		hash := values.Get("hash")
		if hash == "" {
			http.Error(w, "Отсутствует хеш", http.StatusBadRequest)
			return
		}

		// Создаем data-check-string
		pairs := make([]string, 0, len(values))
		for k := range values {
			if k != "hash" {
				pairs = append(pairs, k+"="+values.Get(k))
			}
		}
		sort.Strings(pairs)
		dataCheckString := strings.Join(pairs, "\n")
		// log.Printf("Data check string: %s", dataCheckString)

		// Получаем токен бота из переменных окружения
		botToken := os.Getenv("BOT_TOKEN")
		if botToken == "" {
			log.Printf("Отсутствует токен бота")
			http.Error(w, "Ошибка конфигурации сервера", http.StatusInternalServerError)
			return
		}

		// Создаем секретный ключ
		secretKey := generateSecretKey(botToken)

		// Проверяем подпись
		if !validateSignature(dataCheckString, hash, secretKey) {
			log.Printf("Неверная подпись. Hash: %s", hash)
			http.Error(w, "Неверная подпись", http.StatusUnauthorized)
			return
		}

		// Извлекаем данные пользователя
		userDataStr := values.Get("user")
		if userDataStr != "" {
			// Попробуем извлечь telegram_id из данных пользователя
			var telegramID int64
			var found bool

			// Если telegram_id есть в query параметрах
			telegramIDStr := r.URL.Query().Get("telegram_id")
			if telegramIDStr != "" {
				if id, err := strconv.ParseInt(telegramIDStr, 10, 64); err == nil {
					telegramID = id
					found = true
					log.Printf("Получен telegram_id из query: %d", telegramID)
				}
			}

			// Если telegram_id не найден в query, ищем в теле запроса для POST/PUT
			if !found && (r.Method == "POST" || r.Method == "PUT") {
				// Для POST и PUT, извлекаем telegram_id из userDataStr (JSON в данных от Telegram)
				// Примечание: в реальном приложении используйте более надежный парсинг JSON
				if strings.Contains(userDataStr, "id") {
					// Простой парсинг id из строки JSON
					idStartIndex := strings.Index(userDataStr, "\"id\":")
					if idStartIndex != -1 {
						idStartIndex += 5 // длина "id":
						commaIndex := strings.Index(userDataStr[idStartIndex:], ",")
						if commaIndex == -1 {
							commaIndex = strings.Index(userDataStr[idStartIndex:], "}")
						}
						if commaIndex != -1 {
							idStr := strings.TrimSpace(userDataStr[idStartIndex : idStartIndex+commaIndex])
							if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
								telegramID = id
								found = true
								log.Printf("Получен telegram_id из user data: %d", telegramID)
							}
						}
					}
				}
			}

			// Если нашли telegram_id, добавляем его в контекст
			if found {
				ctx := context.WithValue(r.Context(), "telegram_id", telegramID)
				// Используем новый контекст с запросом
				r = r.WithContext(ctx)
				log.Printf("Установлен telegram_id в контексте: %d", telegramID)
			}
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	}
}

// generateSecretKey генерирует секретный ключ из токена бота
func generateSecretKey(botToken string) []byte {
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(botToken))
	return h.Sum(nil)
}

// validateSignature проверяет подпись данных
func validateSignature(dataCheckString, hash string, secretKey []byte) bool {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	signature := hex.EncodeToString(h.Sum(nil))
	log.Printf("Generated signature: %s", signature)
	log.Printf("Received hash: %s", hash)
	return signature == hash
}
