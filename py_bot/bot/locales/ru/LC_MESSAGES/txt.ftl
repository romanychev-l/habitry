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

# --- AI Анализ привычек ---
ai-user_not_found = ❌ Пользователь не найден. Сначала откройте приложение Habitry.
ai-collecting_data = 📊 Собираю данные о ваших привычках за последнюю неделю...
ai-no_habits = ❌ У вас пока нет привычек. Создайте первую привычку в приложении!
ai-insufficient_data = ❌ Недостаточно данных для анализа. Начните выполнять привычки!
ai-analyzing = 🤖 Анализирую ваши привычки с помощью DeepSeek AI...
ai-analysis_error = ❌ Произошла ошибка при анализе. Попробуйте позже.
ai-command_error = ❌ Произошла ошибка при анализе привычек. Попробуйте позже.
ai-default_username = Пользователь

# --- Дни недели для AI ---
ai-weekday_monday = Понедельник
ai-weekday_tuesday = Вторник
ai-weekday_wednesday = Среда
ai-weekday_thursday = Четверг
ai-weekday_friday = Пятница
ai-weekday_saturday = Суббота
ai-weekday_sunday = Воскресенье

# --- AI Оплата ---
ai-insufficient_balance = ❌ Недостаточно средств для анализа. Необходимо { $required } WILL, а у вас { $current } WILL.
ai-payment_charged = 💰 За анализ списано { $amount } WILL. Ваш баланс: { $balance } WILL.

# --- Команда Add Will ---
add-will_usage = Формат команды:
    /add_will <количество>
    Пример: /add_will 1000
add-will_invalid_amount = Количество должно быть положительным числом
add-will_invalid_format = Количество должно быть числом
add-will_admin_created = ✅ Создан профиль администратора
    💰 Баланс: { $amount } WILL
add-will_balance_updated = ✅ Баланс успешно обновлен!
    📈 Добавлено: { $amount } WILL
    💰 Текущий баланс: { $balance } WILL
add-will_error = Ошибка при добавлении токенов: { $error }