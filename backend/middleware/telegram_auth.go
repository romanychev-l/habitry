package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
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
		log.Printf("Data check string: %s", dataCheckString)

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
