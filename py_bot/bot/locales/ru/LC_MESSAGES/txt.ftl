button-ex_one = Пример кнопки

message-start = Привет! 👋
    
    Я твой персональный помощник на пути к здоровой и счастливой жизни. 

    Вместе мы сможем:
    • Создать полезные привычки
    • Достичь поставленных целей
    • Улучшить качество жизни

    Готов начать новую главу? Нажми кнопку "Открыть" и давай начнем! 💪

message-error_payment = Произошла ошибка при обработке платежа. Пожалуйста, обратитесь в поддержку.
message-payment_success = Оплачено
message-open = Открыть 🚀

# --- Отчет о распределении токенов (Обновлено) ---
report-title = Отчёт о движении WILL за { $date }:
report-section-sent = 📤 Списания:
report-sent-item = - { $amount } WILL за невыполнение '{ $habitTitle }' (распределено подписчикам)
report-section-received = 📥 Получено:
report-received-item = - { $amount } WILL от { $senderInfo } (привычка '{ $fromHabit }') за выполнение вами '{ $forHabit }'
report-summary = 💰 Итого за { $date }: -{ $totalSent } / +{ $totalReceived } WILL

# --- Запасные варианты для отчета ---
report-fallback-unknown-habit = Неизвестная привычка
report-fallback-unknown-user = Неизвестный пользователь
report-fallback-unknown-habit-placeholder = ???

# --- Ping Manager ---
ping-message = 🔔 Время отметить привычки!

# --- Notification Manager ---
notification-reminder = ⏰ Напоминание! Не забудьте про привычку '{ $habitTitle }'! { $link }

# --- Команда Send ---
send-usage = Использование: /send @username <количество>
send-invalid-amount = Пожалуйста, введите корректное положительное количество.
send-sender-not-found = Ошибка: Не удалось найти ваш аккаунт.
send-insufficient-funds = Недостаточно средств. Ваш баланс: { $balance } WILL.
send-recipient-not-found = Ошибка: Пользователь @{ $username } не найден.
send-send-to-self = Вы не можете отправить WILL самому себе.
send-success-sender = Успешно отправлено { $amount } WILL пользователю @{ $username }
send-success-recipient = Вы получили { $amount } WILL от @{ $username }

# --- Общие ошибки ---
error-generic = Произошла непредвиденная ошибка. Пожалуйста, попробуйте позже.