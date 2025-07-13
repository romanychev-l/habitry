message-start = Hello! 👋

    I'm your personal assistant on the path to a healthy and happy life.

    Together we can:
    • Build healthy habits
    • Achieve your goals
    • Improve your quality of life

    Ready to start a new chapter? Click "Open" and let's begin! 💪

message-error_payment = An error occurred while processing the payment. Please contact support.
message-payment_success = Paid
message-open = Open 🚀

# --- Token Distribution Report (Updated) ---
report-title = WILL Movement Report for { $date }:
report-section-sent = 📤 Deductions:
report-sent-item = - { $amount } WILL for not completing '{ $habitTitle }' (distributed to followers)
report-section-received = 📥 Received:
report-received-item = - { $amount } WILL from { $senderInfo } (habit '{ $fromHabit }') for completing your '{ $forHabit }'
report-summary = 💰 Total for { $date }: -{ $totalSent } / +{ $totalReceived } WILL

# --- Report Fallbacks ---
report-fallback-unknown-habit = Unknown habit
report-fallback-unknown-user = Unknown user
report-fallback-unknown-habit-placeholder = ???

# --- Ping Manager ---
ping-message = 🔔 Time to check in on your habits!

# --- Notification Manager ---
notification-reminder = ⏰ Reminder! Don't forget about your habit '{ $habitTitle }'! { $link }

# --- Send Command ---
send-usage = Usage: /send @username <amount>
send-invalid_amount = Please enter a valid positive amount.
send-invalid_number = Amount must be a number
send-sender_not_found = Error: Could not find your user account.
send-insufficient_funds = Insufficient funds. Your balance is { $balance } WILL.
send-insufficient_funds_short = Insufficient funds
send-user_not_found = User @{ $username } not found
send-send_to_self = You cannot send WILL to yourself.
send-success_transfer = Successfully sent { $amount } WILL to @{ $username }
send-success_received = You have received { $amount } WILL from @{ $username }
send-button_received = Received
send-inline_specify_amount = Please specify WILL amount
send-inline_specify_recipient = You want to send { $amount } WILL
    Please specify recipient username
send-inline_send_confirmation = You want to send { $amount } WILL
    Recipient: { $username }
send-inline_invalid_input = Invalid input
send-inline_title_send = Send { $amount } WILL
send-inline_title_default = Send 
send-inline_button_receive = Receive
send-inline_message_template = You want to send { $balance } WILL

# --- Generic Errors ---
error-generic = An unexpected error occurred. Please try again later.

# --- AI Habit Analysis ---
ai-user_not_found = ❌ User not found. Please open the Habitry app first.
ai-collecting_data = 📊 Collecting data about your habits for the past week...
ai-no_habits = ❌ You don't have any habits yet. Create your first habit in the app!
ai-insufficient_data = ❌ Insufficient data for analysis. Start completing your habits!
ai-analyzing = 🤖 Analyzing your habits with DeepSeek AI...
ai-analysis_error = ❌ An error occurred during analysis. Please try again later.
ai-command_error = ❌ An error occurred while analyzing habits. Please try again later.
ai-default_username = User

# --- Weekdays for AI ---
ai-weekday_monday = Monday
ai-weekday_tuesday = Tuesday
ai-weekday_wednesday = Wednesday
ai-weekday_thursday = Thursday
ai-weekday_friday = Friday
ai-weekday_saturday = Saturday
ai-weekday_sunday = Sunday

# --- AI Payment ---
ai-insufficient_balance = ❌ Insufficient funds for analysis. Required { $required } WILL, but you have { $current } WILL.
ai-payment_charged = 💰 { $amount } WILL charged for analysis. Your balance: { $balance } WILL.

# --- Add Will Command ---
add-will_usage = Command format:
    /add_will <amount>
    Example: /add_will 1000
add-will_invalid_amount = Amount must be a positive number
add-will_invalid_format = Amount must be a number
add-will_admin_created = ✅ Admin profile created
    💰 Balance: { $amount } WILL
add-will_balance_updated = ✅ Balance successfully updated!
    📈 Added: { $amount } WILL
    💰 Current balance: { $balance } WILL
add-will_error = Error adding tokens: { $error }