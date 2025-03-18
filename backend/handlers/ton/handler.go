package ton

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TonHandler структура для обработки TON-транзакций
type TonHandler struct {
	usersCollection    *mongo.Collection
	txCollection       *mongo.Collection
	settingsCollection *mongo.Collection
}

// NewHandler создает новый экземпляр TonHandler
func NewHandler(usersCollection, txCollection, settingsCollection *mongo.Collection) *TonHandler {
	return &TonHandler{
		usersCollection:    usersCollection,
		txCollection:       txCollection,
		settingsCollection: settingsCollection,
	}
}

// DepositRequest структура для запроса депозита TON
type DepositRequest struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	WillAmount    int     `json:"will_amount"`
	WalletAddress string  `json:"wallet_address"`
	TelegramID    int64   `json:"telegram_id"`
}

// UsdtDepositRequest структура для запроса депозита USDT
type UsdtDepositRequest struct {
	TransactionID     string  `json:"transaction_id"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	WillAmount        int     `json:"will_amount"`
	WalletAddress     string  `json:"wallet_address"`
	TelegramID        int64   `json:"telegram_id"`
	UsdtMasterAddress string  `json:"usdt_master_address"`
}

// TonTransaction структура для хранения информации о транзакциях
type TonTransaction struct {
	TransactionID    string    `bson:"transaction_id" json:"transaction_id"`
	Amount           float64   `bson:"amount" json:"amount"`     // Единое поле для суммы в любой валюте
	Currency         string    `bson:"currency" json:"currency"` // 'ton' или 'usdt'
	WillAmount       int       `bson:"will_amount" json:"will_amount"`
	WalletAddress    string    `bson:"wallet_address" json:"wallet_address"`
	TelegramID       int64     `bson:"telegram_id" json:"telegram_id"`
	Status           string    `bson:"status" json:"status"`                                             // pending, completed, failed
	PaymentType      string    `bson:"payment_type" json:"payment_type"`                                 // deposit, withdraw
	JettonMasterAddr string    `bson:"jetton_master_addr,omitempty" json:"jetton_master_addr,omitempty"` // для USDT
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

// HandleDeposit обрабатывает депозиты TON
func (h *TonHandler) HandleDeposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.TransactionID == "" || req.Amount <= 0 || req.WillAmount <= 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Устанавливаем currency = "ton" если не указана
	if req.Currency == "" {
		req.Currency = "ton"
	} else if req.Currency != "ton" {
		http.Error(w, "Invalid currency for TON deposit", http.StatusBadRequest)
		return
	}

	// Проверка, что транзакция с таким ID еще не существует
	var existingTx TonTransaction
	err := h.txCollection.FindOne(r.Context(), bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		// Транзакция уже существует
		http.Error(w, "Transaction with this ID already exists", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		// Произошла ошибка при поиске транзакции
		log.Printf("Error checking existing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создание записи о транзакции
	tx := TonTransaction{
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		WillAmount:    req.WillAmount,
		WalletAddress: req.WalletAddress,
		TelegramID:    req.TelegramID,
		Status:        "pending",
		PaymentType:   "deposit",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Сохранение транзакции в БД
	_, err = h.txCollection.InsertOne(r.Context(), tx)
	if err != nil {
		log.Printf("Error saving transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	response := map[string]interface{}{
		"success":        true,
		"transaction_id": req.TransactionID,
		"status":         "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Добавим вспомогательную функцию для конвертации адресов
func normalizeAddress(addr string) string {
	// Если адрес начинается с 0: (raw формат), преобразуем его в user-friendly (EQ)
	if strings.HasPrefix(addr, "0:") {
		if parsedAddr, err := address.ParseAddr(addr); err == nil {
			return parsedAddr.String() // Возвращает в формате EQ...
		}
	}

	// Если адрес начинается с EQ или UQ, оставляем как есть
	if strings.HasPrefix(addr, "EQ") || strings.HasPrefix(addr, "UQ") {
		return addr
	}

	// В остальных случаях возвращаем исходный адрес
	return addr
}

// HandleUsdtDeposit обрабатывает депозиты USDT (Jetton)
func (h *TonHandler) HandleUsdtDeposit(w http.ResponseWriter, r *http.Request) {
	var req UsdtDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.TransactionID == "" || req.Amount <= 0 || req.WillAmount <= 0 || req.UsdtMasterAddress == "" {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Устанавливаем currency = "usdt" если не указана
	if req.Currency == "" {
		req.Currency = "usdt"
	} else if req.Currency != "usdt" {
		http.Error(w, "Invalid currency for USDT deposit", http.StatusBadRequest)
		return
	}

	// Нормализуем адреса
	normalizedWalletAddress := normalizeAddress(req.WalletAddress)
	normalizedMasterAddress := normalizeAddress(req.UsdtMasterAddress)

	log.Printf("Обработка USDT-депозита: ID=%s, сумма=%f, адрес=%s, мастер=%s",
		req.TransactionID, req.Amount, normalizedWalletAddress, normalizedMasterAddress)

	// Проверка, что транзакция с таким ID еще не существует
	var existingTx TonTransaction
	err := h.txCollection.FindOne(r.Context(), bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		// Транзакция уже существует
		http.Error(w, "Transaction with this ID already exists", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		// Произошла ошибка при поиске транзакции
		log.Printf("Error checking existing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создание записи о транзакции с нормализованными адресами
	tx := TonTransaction{
		TransactionID:    req.TransactionID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		WillAmount:       req.WillAmount,
		WalletAddress:    normalizedWalletAddress,
		TelegramID:       req.TelegramID,
		Status:           "pending",
		PaymentType:      "deposit",
		JettonMasterAddr: normalizedMasterAddress,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Сохранение транзакции в БД
	_, err = h.txCollection.InsertOne(r.Context(), tx)
	if err != nil {
		log.Printf("Error saving transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	response := map[string]interface{}{
		"success":        true,
		"transaction_id": req.TransactionID,
		"status":         "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCheckTransaction проверяет статус TON-транзакции
func (h *TonHandler) HandleCheckTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID string `json:"transaction_id"`
		TelegramID    int64  `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.TransactionID == "" || req.TelegramID == 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Получаем информацию о транзакции из БД
	var tx TonTransaction
	err := h.txCollection.FindOne(r.Context(), bson.M{
		"transaction_id": req.TransactionID,
		"telegram_id":    req.TelegramID,
		"currency":       "ton",
	}).Decode(&tx)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Transaction not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching transaction: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"transaction_id": tx.TransactionID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"will_amount":    tx.WillAmount,
		"tx_status":      tx.Status,
		"payment_type":   tx.PaymentType,
		"created_at":     tx.CreatedAt,
		"updated_at":     tx.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCheckUsdtTransaction проверяет статус USDT-транзакции
func (h *TonHandler) HandleCheckUsdtTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID string `json:"transaction_id"`
		TelegramID    int64  `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.TransactionID == "" || req.TelegramID == 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Получаем информацию о транзакции из БД
	var tx TonTransaction
	err := h.txCollection.FindOne(r.Context(), bson.M{
		"transaction_id": req.TransactionID,
		"telegram_id":    req.TelegramID,
		"currency":       "usdt",
	}).Decode(&tx)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Transaction not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching transaction: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"transaction_id": tx.TransactionID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"will_amount":    tx.WillAmount,
		"tx_status":      tx.Status,
		"payment_type":   tx.PaymentType,
		"created_at":     tx.CreatedAt,
		"updated_at":     tx.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// checkTonTransaction проверяет TON-транзакцию
func checkTonTransaction(ctx context.Context, appWalletAddress string, tx TonTransaction) (bool, error) {
	log.Printf("Проверка транзакции %s на сумму %.2f TON от кошелька %s",
		tx.TransactionID, tx.Amount, tx.WalletAddress)
	log.Printf("Адрес кошелька приложения: %s", appWalletAddress)

	// Проверяем, нужно ли пропустить проверку отправителя
	skipSenderCheck := os.Getenv("SKIP_SENDER_CHECK") == "true"
	if skipSenderCheck {
		log.Println("Проверка адреса отправителя отключена")
	}

	// Проверяем, нужно ли пропустить проверку времени
	skipTimeCheck := os.Getenv("SKIP_TIME_CHECK") == "true"
	if skipTimeCheck {
		log.Println("Проверка времени транзакции отключена")
	}

	// Используем TON Center API для проверки транзакций
	tonCenterAPIKey := os.Getenv("TON_CENTER_API_KEY") // можете получить на toncenter.com
	baseURL := "https://toncenter.com/api/v2/getTransactions"
	if tonCenterAPIKey != "" {
		baseURL += "?api_key=" + tonCenterAPIKey
	}

	// Формируем URL с параметрами
	apiURL := fmt.Sprintf("%s&address=%s&limit=20", baseURL, appWalletAddress)
	log.Printf("Запрос к TON Center API: %s", apiURL)

	// Выполняем запрос к API
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("некорректный статус ответа: %d", resp.StatusCode)
	}

	// Парсим ответ
	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			InMsg struct {
				Source      string `json:"source"`
				Destination string `json:"destination"`
				Value       string `json:"value"`
				Message     string `json:"message"`
				CreatedLt   string `json:"created_lt"`
			} `json:"in_msg"`
			TransactionID struct {
				Hash string `json:"hash"`
			} `json:"transaction_id"`
			Utime int64 `json:"utime"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	log.Printf("Получен ответ от TON Center API: ok=%v, количество транзакций=%d", result.OK, len(result.Result))
	log.Printf("Транзакция %s", result.Result[0].TransactionID.Hash)

	if !result.OK || len(result.Result) == 0 {
		log.Printf("Транзакция %s не найдена в блокчейне", tx.TransactionID)
		return false, nil
	}

	// Ищем транзакцию от нужного отправителя на нужную сумму
	expectedValueNano := int64(tx.Amount * 1_000_000_000) // конвертируем в наноТОНы
	log.Printf("Ожидаемая сумма: %d наноТОНов", expectedValueNano)

	// Время создания транзакции в Unix формате
	txCreatedTime := tx.CreatedAt.Unix()
	log.Printf("Время создания транзакции: %d", txCreatedTime)

	// Ищем транзакцию с нашим идентификатором в сообщении
	log.Printf("Ищем транзакцию с идентификатором %s в сообщении", tx.TransactionID)

	for i, transaction := range result.Result {
		log.Printf("Проверка транзакции #%d:", i+1)
		log.Printf("  Отправитель: %s (ожидается: %s)", transaction.InMsg.Source, tx.WalletAddress)
		log.Printf("  Получатель: %s (ожидается: %s)", transaction.InMsg.Destination, appWalletAddress)
		log.Printf("  Сумма: %s наноТОНов", transaction.InMsg.Value)
		log.Printf("  Время: %d", transaction.Utime)
		log.Printf("  Сообщение: %s", transaction.InMsg.Message)

		// Проверяем, содержит ли сообщение наш идентификатор транзакции
		if strings.Contains(transaction.InMsg.Message, tx.TransactionID) {
			log.Printf("  Найден идентификатор транзакции в сообщении")

			// Все остальные проверки становятся дополнительными, так как
			// идентификатор транзакции уникален

			// Проверяем, что транзакция от нужного отправителя, если включена проверка
			if transaction.InMsg.Source != tx.WalletAddress && !skipSenderCheck {
				log.Printf("  Отправитель не совпадает, но проверка отключена, продолжаем")
			}

			// Проверяем сумму транзакции (с небольшой погрешностью из-за комиссии)
			valueNano, err := strconv.ParseInt(transaction.InMsg.Value, 10, 64)
			if err != nil {
				log.Printf("  Ошибка при парсинге суммы транзакции: %v", err)
				continue
			}
			log.Printf("  Разница в сумме: %d наноТОНов", expectedValueNano-valueNano)

			// Допускаем погрешность в 0.05 TON из-за комиссий
			if valueNano < expectedValueNano-50_000_000 {
				log.Printf("  Сумма меньше ожидаемой, пропускаем")
				continue
			}

			// Проверяем, что транзакция была создана после регистрации в нашей системе
			// с запасом в 5 минут для учета возможных расхождений во времени
			timeDiff := transaction.Utime - txCreatedTime
			log.Printf("  Разница во времени: %d секунд", timeDiff)
			if transaction.Utime < txCreatedTime-300 && !skipTimeCheck {
				log.Printf("  Транзакция слишком старая, пропускаем")
				continue
			}

			// Если дошли до этой точки, значит транзакция соответствует всем критериям
			log.Printf("Найдена подтвержденная транзакция для %s в блокчейне", tx.TransactionID)
			return true, nil
		}

		// Если идентификатор не найден в сообщении, проверяем стандартные параметры
		// Проверяем, что транзакция от нужного отправителя
		if transaction.InMsg.Source != tx.WalletAddress && !skipSenderCheck {
			log.Printf("  Отправитель не совпадает, пропускаем")
			continue
		}

		// Проверяем, что назначение транзакции - наш кошелек
		if transaction.InMsg.Destination != appWalletAddress {
			log.Printf("  Получатель не совпадает, пропускаем")
			continue
		}

		// Проверяем сумму транзакции (с небольшой погрешностью из-за комиссии)
		valueNano, err := strconv.ParseInt(transaction.InMsg.Value, 10, 64)
		if err != nil {
			log.Printf("  Ошибка при парсинге суммы транзакции: %v", err)
			continue
		}
		log.Printf("  Разница в сумме: %d наноТОНов", expectedValueNano-valueNano)

		// Допускаем погрешность в 0.05 TON из-за комиссий
		if valueNano < expectedValueNano-50_000_000 {
			log.Printf("  Сумма меньше ожидаемой, пропускаем")
			continue
		}

		// Проверяем, что транзакция была создана после регистрации в нашей системе
		// с запасом в 5 минут для учета возможных расхождений во времени
		timeDiff := transaction.Utime - txCreatedTime
		log.Printf("  Разница во времени: %d секунд", timeDiff)
		if transaction.Utime < txCreatedTime-300 && !skipTimeCheck {
			log.Printf("  Транзакция слишком старая, пропускаем")
			continue
		}

		// Если дошли до этой точки, значит транзакция соответствует всем критериям
		// даже без идентификатора в сообщении
		log.Printf("Найдена подходящая транзакция для %s в блокчейне (без идентификатора в сообщении)", tx.TransactionID)
		return true, nil
	}

	log.Printf("Транзакция %s не найдена в блокчейне или не соответствует ожиданиям", tx.TransactionID)
	return false, nil
}

// CheckUsdtTransaction проверяет USDT-транзакцию следуя примеру работы с Jetton
func (h *TonHandler) CheckUsdtTransaction(ctx context.Context) (bool, error) {
	// log.Printf("Начинаем проверку USDT транзакции: %s", tx.TransactionID)
	// log.Printf("Параметры транзакции: сумма=%f USDT, адрес кошелька=%s, мастер-контракт=%s",
	// 	tx.UsdtAmount, tx.WalletAddress, tx.JettonMasterAddr)

	// Нормализуем адреса перед использованием
	// normalizedAppWalletAddress := normalizeAddress(appWalletAddress)
	appWalletAddress := os.Getenv("TON_WALLET_ADDRESS")
	if appWalletAddress == "" {
		log.Fatal("TON_WALLET_ADDRESS environment variable not set")
	}
	usdtMasterAddress := os.Getenv("USDT_MASTER_ADDRESS")
	if usdtMasterAddress == "" {
		log.Fatal("USDT_MASTER_ADDRESS environment variable not set")
	}

	normalizedJettonMasterAddr := normalizeAddress(usdtMasterAddress)

	log.Printf("Нормализованные адреса: приложение=%s, мастер-контракт=%s",
		appWalletAddress, normalizedJettonMasterAddr)

	// Проверяем формат адреса мастер-контракта
	if (!strings.HasPrefix(normalizedJettonMasterAddr, "EQ") && !strings.HasPrefix(normalizedJettonMasterAddr, "UQ")) &&
		(!strings.HasPrefix(normalizedJettonMasterAddr, "0:")) {
		log.Printf("Некорректный формат адреса Jetton мастер-контракта после нормализации: %s", normalizedJettonMasterAddr)
		return false, fmt.Errorf("некорректный формат адреса Jetton мастер-контракта: %s", normalizedJettonMasterAddr)
	}

	client := liteclient.NewConnectionPool()

	cfg, err := liteclient.GetConfigFromUrl(ctx, "https://ton.org/global.config.json")
	if err != nil {
		log.Printf("Ошибка получения конфигурации: %v", err)
		return false, fmt.Errorf("ошибка получения конфигурации: %v", err)
	}
	log.Println("Конфигурация TON получена успешно")

	// Подключаемся к lite серверам
	err = client.AddConnectionsFromConfig(ctx, cfg)
	if err != nil {
		log.Printf("Ошибка подключения к серверам: %v", err)
		return false, fmt.Errorf("ошибка подключения: %v", err)
	}
	log.Println("Подключение к lite серверам установлено")

	// Инициализируем API клиент
	api := ton.NewAPIClient(client, ton.ProofCheckPolicySecure).WithRetry()
	master, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		log.Printf("Ошибка получения информации о мастерчейне: %v", err)
		return false, fmt.Errorf("ошибка получения информации о мастерчейне: %v", err)
	}
	log.Printf("Получена информация о мастерчейне: seqno=%d", master.SeqNo)

	// Устанавливаем trusted block для улучшения безопасности
	api.SetTrustedBlock(master)
	log.Printf("Установлен trusted block: seqno=%d", master.SeqNo)

	// Парсим адрес приложения
	treasuryAddress, err := address.ParseAddr(appWalletAddress)
	if err != nil {
		log.Printf("Ошибка парсинга адреса приложения: %v", err)
		return false, fmt.Errorf("ошибка парсинга адреса приложения: %v", err)
	}
	log.Printf("Адрес кошелька приложения: %s", treasuryAddress.String())

	// Парсим адрес Jetton мастер-контракта
	jettonMasterAddr, err := address.ParseAddr(normalizedJettonMasterAddr)
	if err != nil {
		log.Printf("Ошибка парсинга адреса Jetton мастер-контракта: %v", err)
		return false, fmt.Errorf("ошибка парсинга адреса Jetton мастер-контракта: %v", err)
	}
	log.Printf("Адрес Jetton мастер-контракта: %s", jettonMasterAddr.String())

	// Получаем аккаунт приложения
	acc, err := api.GetAccount(ctx, master, treasuryAddress)
	if err != nil {
		log.Printf("Ошибка получения аккаунта: %v", err)
		return false, fmt.Errorf("ошибка получения аккаунта: %v", err)
	}
	log.Printf("Получена информация об аккаунте. LastTxLT: %d, LastTxHash: %x", acc.LastTxLT, acc.LastTxHash)

	// Пытаемся получить сохраненный lastProcessedLT из базы данных
	var settings struct {
		Key   string `bson:"key"`
		Value uint64 `bson:"value"`
	}
	err = h.settingsCollection.FindOne(ctx, bson.M{"key": "usdt_last_tx_lt"}).Decode(&settings)

	var lastProcessedLT uint64
	if err == nil {
		// Используем сохраненное значение
		lastProcessedLT = settings.Value
		log.Printf("Найдено сохраненное значение lastProcessedLT: %d", lastProcessedLT)
	} else if err == mongo.ErrNoDocuments {
		// Если записи нет, используем текущий LT
		lastProcessedLT = acc.LastTxLT
		log.Printf("Не найдено сохраненное значение lastProcessedLT, используем текущий: %d", lastProcessedLT)
	} else {
		// Если возникла ошибка при поиске
		log.Printf("Ошибка получения lastProcessedLT из БД: %v, используем текущий: %d", err, acc.LastTxLT)
		lastProcessedLT = acc.LastTxLT
	}

	// channel with new transactions
	transactions := make(chan *tlb.Transaction)

	// it is a blocking call, so we start it asynchronously
	go api.SubscribeOnTransactions(context.Background(), treasuryAddress, lastProcessedLT, transactions)

	log.Println("waiting for transfers...")

	// Проверяем транзакции
	for tx := range transactions {
		// Обновляем последний обработанный LT сразу после получения транзакции
		_, err = h.settingsCollection.UpdateOne(
			ctx,
			bson.M{"key": "usdt_last_tx_lt"},
			bson.M{"$set": bson.M{"value": tx.LT}},
			options.Update().SetUpsert(true), // Создаем запись, если она не существует
		)
		if err != nil {
			log.Printf("Ошибка при обновлении lastProcessedLT в БД: %v", err)
		} else {
			log.Printf("Обновлен lastProcessedLT: %d", tx.LT)
		}

		if tx.IO.In != nil && tx.IO.In.MsgType == tlb.MsgTypeInternal {
			ti := tx.IO.In.AsInternal()
			src := ti.SrcAddr

			if dsc, ok := tx.Description.(tlb.TransactionDescriptionOrdinary); ok && dsc.BouncePhase != nil {
				if _, ok = dsc.BouncePhase.Phase.(tlb.BouncePhaseOk); ok {
					// transaction was bounced, and coins were returned to sender
					// this can happen mostly on custom contracts
					continue
				}
			}

			if !ti.ExtraCurrencies.IsEmpty() {
				kv, err := ti.ExtraCurrencies.LoadAll()
				if err != nil {
					log.Fatalln("load extra currencies err: ", err.Error())
					return false, fmt.Errorf("load extra currencies err: %v", err)
				}

				for _, dictKV := range kv {
					currencyId := dictKV.Key.MustLoadUInt(32)
					amount := dictKV.Value.MustLoadVarUInt(32)

					log.Println("received", amount.String(), "ExtraCurrency with id", currencyId, "from", src.String())
				}
			}

			// verify that event sender is our jetton wallet
			treasuryJettonWallet, err := address.ParseAddr("EQDRE5lNhaH2M1nnCxTXFmUlQveGVoO10Tl5j495R2B9ZfMi")
			if ti.SrcAddr.Equals(treasuryJettonWallet) {
				var transfer jetton.TransferNotification
				if err = tlb.LoadFromCell(&transfer, ti.Body.BeginParse()); err == nil {
					// convert decimals to 6 for USDT (it can be fetched from jetton details too), default is 9
					amt := tlb.MustFromNano(transfer.Amount.Nano(), 6)
					payload := transfer.ForwardPayload.BeginParse()

					payloadOp := payload.MustLoadUInt(32)
					if payloadOp != 0 {
						log.Println("no text comment in transfer_notification")
						continue
					}

					comment := payload.MustLoadStringSnake()
					log.Println("comment", comment)

					// Ищем транзакцию в базе данных по комментарию
					var transaction TonTransaction
					sr := h.txCollection.FindOne(context.Background(), bson.M{"transaction_id": comment})
					err := sr.Decode(&transaction)
					if err != nil {
						log.Printf("Ошибка при поиске транзакции по комментарию: %v", err)
						continue
					}

					// Обновляем статус транзакции
					updateResult, err := h.txCollection.UpdateOne(
						context.Background(),
						bson.M{"transaction_id": transaction.TransactionID},
						bson.M{
							"$set": bson.M{
								"status":     "completed",
								"updated_at": time.Now(),
							},
						},
					)
					if err != nil {
						log.Printf("Ошибка при обновлении статуса транзакции: %v", err)
						continue
					}
					log.Printf("Статус транзакции обновлен: %v", updateResult)

					// Обновляем баланс пользователя
					updateResult, err = h.usersCollection.UpdateOne(
						context.Background(),
						bson.M{"telegram_id": transaction.TelegramID},
						bson.M{"$inc": bson.M{"balance": transaction.WillAmount}},
					)
					if err != nil {
						log.Printf("Ошибка при обновлении баланса пользователя: %v", err)
						continue
					}
					log.Printf("Баланс пользователя обновлен: %v", updateResult)
					// reassign sender to real jetton sender instead of its jetton wallet contract
					src = transfer.Sender
					log.Println("received", amt.String(), "USDT from", src.String())
				}
			}

			if ti.Amount.Nano().Sign() > 0 {
				// show received ton amount
				log.Println("received", ti.Amount.String(), "TON from", src.String())
			}
		}
	}

	return false, nil
}

// ProcessWithdrawals обрабатывает запросы на вывод средств
func (h *TonHandler) ProcessWithdrawals(ctx context.Context) error {
	log.Println("Начинаем обработку запросов на вывод WILL")

	// Находим все транзакции вывода со статусом "pending"
	filter := bson.M{
		"status":       "pending",
		"payment_type": "withdraw",
		"currency":     "usdt", // Сейчас обрабатываем только USDT
	}

	cursor, err := h.txCollection.Find(ctx, filter)
	if err != nil {
		log.Printf("Ошибка при поиске транзакций вывода: %v", err)
		return fmt.Errorf("ошибка при поиске транзакций вывода: %v", err)
	}
	defer cursor.Close(ctx)

	// Получаем адрес мастер-контракта USDT из переменных окружения
	usdtMasterAddr := os.Getenv("USDT_MASTER_ADDRESS")
	if usdtMasterAddr == "" {
		log.Println("USDT_MASTER_ADDRESS не установлен, используем адрес по умолчанию")
		return fmt.Errorf("USDT_MASTER_ADDRESS не установлен")
	}

	appWalletAddressUserFriendly := os.Getenv("TON_WALLET_ADDRESS")
	if appWalletAddressUserFriendly == "" {
		log.Fatal("TON_WALLET_ADDRESS environment variable not set")
	}

	appWalletAddress, err := address.ParseAddr(appWalletAddressUserFriendly)
	if err != nil {
		log.Printf("Ошибка при парсинге адреса приложения: %v", err)
		return fmt.Errorf("ошибка при парсинге адреса приложения: %v", err)
	}

	// Seed-фраза для кошелька приложения (в реальном приложении должна храниться в безопасном месте)
	seedPhrase := os.Getenv("WALLET_SEED_PHRASE")
	if seedPhrase == "" {
		log.Printf("WALLET_SEED_PHRASE не установлен, вывод средств невозможен")
		return fmt.Errorf("WALLET_SEED_PHRASE не установлен")
	}

	// Установка соединения с TON
	client := liteclient.NewConnectionPool()
	err = client.AddConnectionsFromConfigUrl(ctx, "https://ton.org/global.config.json")
	if err != nil {
		log.Printf("Ошибка при подключении к TON: %v", err)
		return fmt.Errorf("ошибка при подключении к TON: %v", err)
	}

	// Инициализация API клиента
	api := ton.NewAPIClient(client)

	// Создаем кошелек из seed-фразы
	words := strings.Split(seedPhrase, " ")
	log.Println("words", words)
	w, err := wallet.FromSeed(api, words, wallet.ConfigV5R1Final{NetworkGlobalID: -239, Workchain: 0})
	if err != nil {
		log.Printf("Ошибка при создании кошелька: %v", err)
		return fmt.Errorf("ошибка при создании кошелька: %v", err)
	}

	log.Printf("Кошелек приложения инициализирован: %s", w.WalletAddress().String())

	// Инициализация Jetton мастер-клиента
	jettonMasterAddr, err := address.ParseAddr(usdtMasterAddr)
	if err != nil {
		log.Printf("Ошибка при парсинге адреса Jetton мастер-контракта: %v", err)
		return fmt.Errorf("ошибка при парсинге адреса Jetton мастер-контракта: %v", err)
	}

	token := jetton.NewJettonMasterClient(api, jettonMasterAddr)

	// Находим наш Jetton-кошелек
	tokenWallet, err := token.GetJettonWallet(ctx, w.WalletAddress())
	if err != nil {
		log.Printf("Ошибка при получении Jetton-кошелька: %v", err)
		return fmt.Errorf("ошибка при получении Jetton-кошелька: %v", err)
	}

	// Получаем баланс Jetton-кошелька
	tokenBalance, err := tokenWallet.GetBalance(ctx)
	if err != nil {
		log.Printf("Ошибка при получении баланса Jetton-кошелька: %v", err)
		return fmt.Errorf("ошибка при получении баланса Jetton-кошелька: %v", err)
	}

	log.Printf("Баланс USDT кошелька приложения: %s", tokenBalance.String())

	// Обрабатываем каждую транзакцию
	for cursor.Next(ctx) {
		var tx TonTransaction
		if err := cursor.Decode(&tx); err != nil {
			log.Printf("Ошибка при декодировании транзакции: %v", err)
			continue
		}

		log.Printf("Обработка транзакции вывода %s: %f USDT на адрес %s",
			tx.TransactionID, tx.Amount, tx.WalletAddress)

		// Рассчитываем комиссию 1%
		originalAmount := tx.Amount
		fee := originalAmount * 0.01
		finalAmount := originalAmount - fee

		log.Printf("Расчет комиссии: Исходная сумма: %f USDT, Комиссия (1%%): %f USDT, Итоговая сумма: %f USDT",
			originalAmount, fee, finalAmount)

		// Проверяем, что у нас достаточно токенов для вывода
		withdrawAmountNano := tlb.MustFromDecimal(fmt.Sprintf("%f", finalAmount), 6)
		if tokenBalance.Cmp(withdrawAmountNano.Nano()) < 0 {
			log.Printf("Недостаточно USDT на балансе кошелька приложения для вывода. Требуется: %s, доступно: %s",
				withdrawAmountNano.String(), tokenBalance.String())

			// Устанавливаем статус "failed" для транзакции
			_, err = h.txCollection.UpdateOne(
				ctx,
				bson.M{"transaction_id": tx.TransactionID},
				bson.M{
					"$set": bson.M{
						"status":     "failed",
						"updated_at": time.Now(),
					},
				},
			)
			if err != nil {
				log.Printf("Ошибка при обновлении статуса транзакции на failed: %v", err)
			}

			continue
		}

		// Парсим адрес получателя
		log.Println("tx.WalletAddress", tx.WalletAddress)
		log.Println("Normalized", normalizeAddress(tx.WalletAddress))
		recipientAddr, err := address.ParseRawAddr(tx.WalletAddress)
		log.Println("recipientAddr", recipientAddr)
		if err != nil {
			log.Printf("Ошибка при парсинге адреса получателя: %v", err)

			// Устанавливаем статус "failed" для транзакции
			_, err = h.txCollection.UpdateOne(
				ctx,
				bson.M{"transaction_id": tx.TransactionID},
				bson.M{
					"$set": bson.M{
						"status":     "failed",
						"updated_at": time.Now(),
					},
				},
			)
			if err != nil {
				log.Printf("Ошибка при обновлении статуса транзакции на failed: %v", err)
			}

			continue
		}

		// Создаем комментарий транзакции
		comment, err := wallet.CreateCommentCell(tx.TransactionID)
		if err != nil {
			log.Printf("Ошибка при создании комментария: %v", err)
			continue
		}

		// Создаем payload для перевода Jetton
		transferPayload, err := tokenWallet.BuildTransferPayloadV2(
			recipientAddr,                  // адрес получателя
			appWalletAddress,               // адрес для ответа
			withdrawAmountNano,             // сумма перевода
			tlb.MustFromTON("0.000000001"), // тоны для оплаты комиссии форвард-сообщения (0)
			comment,                        // комментарий
			nil,                            // дополнительный payload
		)
		if err != nil {
			log.Printf("Ошибка при создании payload для перевода: %v", err)
			continue
		}

		// Создаем сообщение для перевода (0.05 TON для оплаты комиссий)
		msg := wallet.SimpleMessage(tokenWallet.Address(), tlb.MustFromTON("0.05"), transferPayload)

		// Устанавливаем статус "processing" для транзакции
		_, err = h.txCollection.UpdateOne(
			ctx,
			bson.M{"transaction_id": tx.TransactionID},
			bson.M{
				"$set": bson.M{
					"status":     "processing",
					"updated_at": time.Now(),
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка при обновлении статуса транзакции на processing: %v", err)
			continue
		}

		log.Printf("Отправка транзакции перевода USDT...")

		// Отправляем транзакцию и ждем подтверждения
		txResult, _, err := w.SendWaitTransaction(ctx, msg)
		if err != nil {
			log.Printf("Ошибка при отправке транзакции: %v", err)

			// Устанавливаем статус "failed" для транзакции
			_, err = h.txCollection.UpdateOne(
				ctx,
				bson.M{"transaction_id": tx.TransactionID},
				bson.M{
					"$set": bson.M{
						"status":     "failed",
						"updated_at": time.Now(),
					},
				},
			)
			if err != nil {
				log.Printf("Ошибка при обновлении статуса транзакции на failed: %v", err)
			}

			continue
		}

		// Транзакция подтверждена, обновляем статус
		_, err = h.txCollection.UpdateOne(
			ctx,
			bson.M{"transaction_id": tx.TransactionID},
			bson.M{
				"$set": bson.M{
					"status":     "completed",
					"updated_at": time.Now(),
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка при обновлении статуса транзакции на completed: %v", err)
			continue
		}

		log.Printf("Транзакция вывода %s успешно выполнена, хэш: %x",
			tx.TransactionID, txResult.Hash)

		// Обновляем баланс Jetton-кошелька после транзакции
		tokenBalance, err = tokenWallet.GetBalance(ctx)
		if err != nil {
			log.Printf("Ошибка при получении обновленного баланса Jetton-кошелька: %v", err)
		} else {
			log.Printf("Новый баланс USDT кошелька приложения: %s", tokenBalance.String())
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Ошибка курсора при обработке транзакций: %v", err)
		return fmt.Errorf("ошибка курсора при обработке транзакций: %v", err)
	}

	log.Println("Обработка запросов на вывод WILL завершена")
	return nil
}

// WithdrawRequest структура для запроса вывода WILL токенов
type WithdrawRequest struct {
	TransactionID string  `json:"transaction_id"`
	WillAmount    int     `json:"will_amount"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	WalletAddress string  `json:"wallet_address"`
	TelegramID    int64   `json:"telegram_id"`
}

// HandleWithdraw обрабатывает запросы на вывод WILL токенов
func (h *TonHandler) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Валидация параметров запроса
	if req.TransactionID == "" || req.WillAmount <= 0 || req.WalletAddress == "" || req.TelegramID == 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Устанавливаем currency = "usdt" если не указана
	if req.Currency == "" {
		req.Currency = "usdt"
	} else if req.Currency != "usdt" && req.Currency != "ton" {
		http.Error(w, "Invalid currency for withdrawal", http.StatusBadRequest)
		return
	}

	// Проверка существования транзакции с таким ID
	var existingTx TonTransaction
	err := h.txCollection.FindOne(r.Context(), bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		// Транзакция уже существует
		http.Error(w, "Transaction with this ID already exists", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		// Произошла ошибка при поиске транзакции
		log.Printf("Error checking existing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Получаем информацию о пользователе
	var user struct {
		TelegramID int64 `bson:"telegram_id"`
		Balance    int   `bson:"balance"`
	}
	err = h.usersCollection.FindOne(r.Context(), bson.M{"telegram_id": req.TelegramID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем, достаточно ли средств у пользователя
	if user.Balance < req.WillAmount {
		http.Error(w, fmt.Sprintf("Insufficient funds. Current balance: %d WILL", user.Balance), http.StatusBadRequest)
		return
	}

	// Нормализуем адрес кошелька
	normalizedWalletAddress := normalizeAddress(req.WalletAddress)

	// Создаем запись о транзакции вывода
	tx := TonTransaction{
		TransactionID:    req.TransactionID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		WillAmount:       -req.WillAmount, // Отрицательное значение, так как это вывод
		WalletAddress:    normalizedWalletAddress,
		TelegramID:       req.TelegramID,
		Status:           "pending",
		PaymentType:      "withdraw",
		JettonMasterAddr: os.Getenv("USDT_MASTER_ADDRESS"), // Используем адрес из конфигурации
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Сохраняем транзакцию в БД
	_, err = h.txCollection.InsertOne(r.Context(), tx)
	if err != nil {
		log.Printf("Error saving withdraw transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Уменьшаем баланс пользователя
	_, err = h.usersCollection.UpdateOne(
		r.Context(),
		bson.M{"telegram_id": req.TelegramID},
		bson.M{"$inc": bson.M{"balance": -req.WillAmount}}, // Уменьшаем баланс
	)
	if err != nil {
		log.Printf("Error updating user balance: %v", err)

		// Откатываем изменения - удаляем транзакцию
		_, deleteErr := h.txCollection.DeleteOne(r.Context(), bson.M{"transaction_id": req.TransactionID})
		if deleteErr != nil {
			log.Printf("Error rolling back transaction: %v", deleteErr)
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"success":        true,
		"transaction_id": req.TransactionID,
		"status":         "pending",
		"message":        "Withdraw request accepted and will be processed within 24 hours",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
