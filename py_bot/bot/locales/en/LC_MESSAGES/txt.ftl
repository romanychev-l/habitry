button-ex_one = Example of a button

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
send-invalid-amount = Please enter a valid positive amount.
send-sender-not-found = Error: Could not find your user account.
send-insufficient-funds = Insufficient funds. Your balance is { $balance } WILL.
send-recipient-not-found = Error: User @{ $username } not found.
send-send-to-self = You cannot send WILL to yourself.
send-success-sender = Successfully sent { $amount } WILL to @{ $username }
send-success-recipient = You have received { $amount } WILL from @{ $username }

# --- Generic Errors ---
error-generic = An unexpected error occurred. Please try again later.