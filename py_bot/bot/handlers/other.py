import logging
from datetime import datetime, timedelta
from aiogram import Router, F, types, Bot
from aiogram.types import Message, ReplyKeyboardRemove, InputTextMessageContent, InlineQuery, InlineQueryResultArticle
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram_dialog import DialogManager
from fluentogram import TranslatorRunner
import uuid
from aiogram.utils.keyboard import InlineKeyboardBuilder


from bot.config_data.config import db, config_settings

other_router = Router()


class Actions(StatesGroup):
    base_state = State()


@other_router.message(Command("stats"))
async def cmd_stats(msg: Message):
    if msg.from_user.id != 248603604:  # Ваш ID
        logging.info(f"Unauthorized stats access attempt from user {msg.from_user.id}")
        return
        
    try:
        logging.info("Starting stats collection...")
        
        # Проверяем подключение к БД
        try:
            db['users'].database.client.server_info()
            logging.info("Successfully connected to MongoDB")
        except Exception as e:
            logging.error(f"MongoDB connection error: {e}")
            await msg.answer("Ошибка подключения к базе данных")
            return

        # Получаем всех пользователей
        total_users = db['users'].count_documents({})
        logging.info(f"Found {total_users} users in database")

        # Получаем все привычки
        habits = list(db['habits'].find({}))
        total_habits = len(habits)
        
        # Подготавливаем даты для фильтрации
        now = datetime.utcnow()
        today = now.date()
        yesterday = (now - timedelta(days=1)).date()
        
        # Получаем документы за сегодня и считаем сумму выполненных привычек
        today_docs = list(db['history'].find({"date": today.isoformat()}))
        completed_today = sum(len(doc.get('habits', [])) for doc in today_docs)
        
        # Получаем документы за вчера и считаем сумму выполненных привычек
        yesterday_docs = list(db['history'].find({"date": yesterday.isoformat()}))
        completed_yesterday = sum(len(doc.get('habits', [])) for doc in yesterday_docs)
        
        # Считаем общее количество выполнений из истории
        all_history_docs = list(db['history'].find({}))
        total_completions = sum(len(doc.get('habits', [])) for doc in all_history_docs)
        
        # Считаем общее количество связей между привычками
        total_links = sum(len(habit.get("followers", [])) for habit in habits)
        
        logging.info(f"Stats collected: users={total_users}, habits={total_habits}, "
                    f"completed_today={completed_today}, completed_yesterday={completed_yesterday}, "
                    f"total_completions={total_completions}, total_links={total_links}")
        
        # Формируем сообщение со статистикой
        stats_message = (
            f"📊 Статистика Habitry:\n\n"
            f"👥 Всего пользователей: {total_users}\n"
            f"📝 Всего привычек: {total_habits}\n"
            f"🔗 Всего связей между привычками: {total_links}\n"
            f"👍 Выполнено привычек сегодня: {completed_today}\n"
            f"✅ Выполнено привычек вчера: {completed_yesterday}\n"
            f"🏆 Общее количество выполнений: {total_completions}\n"
        )
        
        await msg.answer(stats_message)
        
    except Exception as e:
        logging.error(f"Error in stats command: {e}")
        await msg.answer(f"Ошибка при получении статистики: {str(e)}")

    
@other_router.message(Command("today_list_users"))
async def cmd_today_list_users(msg: Message):
    if msg.from_user.id != 248603604:  # Ваш ID
        logging.info(f"Unauthorized stats access attempt from user {msg.from_user.id}")
        return
        
    try:
        # Получаем текущую дату
        now = datetime.utcnow()
        today = now.date()
        
        # Получаем все документы за сегодня
        today_docs = list(db['history'].find({"date": today.isoformat()}))
        
        if not today_docs:
            await msg.answer("Сегодня пока никто не выполнил привычки")
            return
            
        # Создаем словарь для хранения статистики по пользователям
        user_stats = {}
        
        # Обрабатываем каждый документ
        for doc in today_docs:
            user = db['users'].find_one({"telegram_id": doc.get('telegram_id')})
            if user:
                username = user.get('username', 'unknown')
                user_stats[username] = len(doc.get('habits', []))
        
        # Формируем сообщение
        message = "📊 Статистика выполнения привычек сегодня:\n\n"
        
        for username, count in sorted(user_stats.items(), key=lambda x: x[1], reverse=True):
            profile_link = f"t.me/{config_settings.BOT_USERNAME}/app?startapp=profile_{username}"
            message += f"@{username}: {count} привычек\n"
            message += f"Профиль: {profile_link}\n\n"
            
        await msg.answer(message)
        
    except Exception as e:
        logging.error(f"Error in today_list_users command: {e}")
        await msg.answer(f"Ошибка при получении статистики: {str(e)}")


@other_router.message(Actions.base_state)
async def registration_start(msg: Message, state: FSMContext, i18n: TranslatorRunner, dialog_manager: DialogManager):
    # await state.update_data(status='employee')
    
    await msg.answer(i18n.message.ex_one(), reply_markup=ReplyKeyboardRemove())


# @other_router.message(Command("buy"))
# async def cmd_buy(msg: Message, bot: Bot):
#     await bot.send_invoice(
#         chat_id=msg.chat.id,
#         title="1 Telegram Stars",
#         description="Покупка Stars для поддержки канала",
#         payload="stars_1",
#         provider_token="",  # Пустой токен для Stars
#         currency="XTR",
#         prices=[
#             types.LabeledPrice(
#                 label="1 Stars",
#                 amount=1  # 100 Stars = 10000
#             )
#         ],
#         need_name=False,
#         need_phone_number=False,
#         need_email=False,
#         need_shipping_address=False,
#         is_flexible=False
#     )

# Обработчик pre_checkout_query
# @other_router.pre_checkout_query()
# async def process_pre_checkout_query(pre_checkout_query: types.PreCheckoutQuery, bot: Bot):
#     try:
#         await bot.answer_pre_checkout_query(
#             pre_checkout_query_id=pre_checkout_query.id,
#             ok=True
#         )
#         logging.info(f"Pre-checkout query processed: {pre_checkout_query.id}")
#     except Exception as e:
#         logging.error(f"Error in pre_checkout_query: {e}")

# Обработчик успешного платежа
# @other_router.message(F.successful_payment)
# async def successful_payment(message: types.Message, i18n: TranslatorRunner):
#     print(message)
#     try:
#         user_id = message.from_user.id
#         stars_amount = message.successful_payment.total_amount
#         will_tokens = stars_amount * 10  # 1 Stars = 10 WILL

#         # Получаем текущий баланс пользователя
#         user_record = db.users.find_one({"telegram_id": user_id}) # Исправлено users_collection на users
#         current_balance = user_record.get('balance', 0) if user_record else 0
#         new_balance = current_balance + will_tokens

#         result = db.users.update_one( # Исправлено users_collection на users
#             {"telegram_id": user_id},
#             {"$set": {"balance": new_balance}}
#         )
        
#         if result.modified_count > 0:
#             await message.answer(
#                 f"{i18n.message.payment_success()} {stars_amount} Stars\n"
#                 f"Начислено: {will_tokens} WILL"
#             )
#         else:
#             # Если пользователь не найден, создаем его
#             if not user_record:
#                  db.users.insert_one({
#                      "telegram_id": user_id,
#                      "username": message.from_user.username or f"user_{user_id}",
#                      "balance": will_tokens,
#                      "created_at": datetime.utcnow()
#                  })
#                  logging.info(f"New user created via payment: {user_id}")
#                  await message.answer(
#                      f"{i18n.message.payment_success()} {stars_amount} Stars\n"
#                      f"Начислено: {will_tokens} WILL"
#                  )
#             else:
#                  logging.error(f"Failed to update balance for user {user_id}")
#                  # await message.answer(i18n.message.error_payment()) # Можно раскомментировать, если нужно явное сообщение об ошибке

#     except Exception as e:
#         logging.error(f"Ошибка при обновлении БД: {e}")
#         await message.answer(i18n.message.error_payment())


@other_router.message(Command("send"))
async def cmd_send_will(msg: Message, i18n: TranslatorRunner, bot: Bot):
    try:
        args = msg.text.split()
        logging.info(f"Received send command with args: {args}")
        if len(args) != 3:
            await msg.answer(i18n.send.usage())
            return

        _, recipient_username, amount_str = args
        
        # Убираем @ если есть
        recipient_username = recipient_username.lstrip('@')

        try:
            amount = int(amount_str)
            if amount <= 0:
                raise ValueError("Amount must be positive")
        except ValueError:
            await msg.answer(i18n.send.invalid_amount())
            return

        sender_id = msg.from_user.id
        sender = db.users.find_one({"telegram_id": sender_id})

        if not sender:
            logging.error(f"Sender user not found: {sender_id}")
            await msg.answer(i18n.send.sender_not_found())
            return

        sender_balance = sender.get('balance', 0)

        if sender_balance < amount:
            await msg.answer(i18n.send.insufficient_funds(balance=sender_balance))
            return

        recipient = db.users.find_one({"username": recipient_username})

        if not recipient:
            await msg.answer(i18n.send.recipient_not_found(username=recipient_username))
            return
            
        recipient_id = recipient.get("telegram_id")
        if sender_id == recipient_id:
            await msg.answer(i18n.send.send_to_self())
            return

        recipient_balance = recipient.get('balance', 0)

        # Обновляем балансы
        db.users.update_one({"telegram_id": sender_id}, {"$inc": {"balance": -amount}})
        db.users.update_one({"telegram_id": recipient_id}, {"$inc": {"balance": amount}})
        
        # Логируем транзакцию (опционально, можно создать отдельную коллекцию)
        db.transactions.insert_one({
            "sender_id": sender_id,
            "recipient_id": recipient_id,
            "recipient_username": recipient_username,
            "amount": amount,
            "timestamp": datetime.utcnow()
        })

        # Уведомление отправителю
        # await msg.answer(i18n.send.success_sender(amount=amount, username=recipient_username)) # Используем send-success-sender
        success_message = i18n.get("send-success-sender", amount=amount, username=recipient_username)
        await msg.answer(success_message)

        # Уведомление получателю
        try:
            sender_username = msg.from_user.username or f"user_{sender_id}"
            # recipient_text = i18n.send.success_recipient(amount=amount, username=sender_username) # Используем send-success-recipient
            recipient_text = i18n.get("send-success-recipient", amount=amount, username=sender_username)
            
            await bot.send_message(
                chat_id=recipient_id,
                text=recipient_text
            )
        except Exception as e:
            logging.error(f"Failed to notify recipient {recipient_id}: {e}")
            # Можно добавить логику возврата средств, если уведомление критично

        logging.info(f"WILL transfer: {sender_id} -> {recipient_id} ({recipient_username}), amount: {amount}")

    except Exception as e:
        logging.exception(f"Error in /send command: {e}")
        await msg.answer(i18n.error.generic())

@other_router.message(Command("add_quote"))
async def cmd_add_quote(msg: Message):
    # Проверка на администратора
    if msg.from_user.id != 248603604:  # Ваш ID
        logging.info(f"Unauthorized quote addition attempt from user {msg.from_user.id}")
        return
        
    try:
        # Парсим сообщение
        lines = msg.text.split('\n')
        if len(lines) < 2:
            await msg.answer(
                "Формат:\n"
                "/add_quote\n"
                "ru\n"
                "Текст цитаты на русском\n"
                "Автор (опционально)\n"
                "en\n"
                "Text in English\n"
                "Author (optional)"
            )
            return
            
        quote_doc = {}
        current_lang = None
        current_text = None
        current_autor = None
        
        # Пропускаем первую строку с командой
        for line in lines[1:]:
            line = line.strip()
            if not line:
                continue
                
            if line in ['ru', 'en']:
                # Если был предыдущий язык, сохраняем его данные
                if current_lang and current_text:
                    quote_doc[current_lang] = {
                        "text": current_text,
                        "autor": current_autor  # Будет обновлено, если найдем автора
                    }
                current_lang = line
                current_text = None
                current_autor = None
            elif current_lang and not current_text:
                current_text = line
            elif current_lang and current_text:
                current_autor = line
        
        # Сохраняем последний блок
        if current_lang and current_text:
            if current_lang not in quote_doc:
                quote_doc[current_lang] = {
                    "text": current_text,
                    "autor": current_autor
                }
        
        if not quote_doc:
            raise ValueError("No valid quotes found in message")
            
        # Добавляем в базу данных
        result = db.quotes.insert_one(quote_doc)
        
        # Формируем ответное сообщение
        response = "Цитата успешно добавлена!\n"
        
        await msg.answer(response)
        
    except Exception as e:
        logging.error(f"Error adding quote: {e}")
        await msg.answer("Ошибка при добавлении цитаты. Проверьте формат команды.")


@other_router.callback_query(F.data == "received")
async def callback_query_received(callback_query: types.CallbackQuery):
    logging.info(f"Callback query: {callback_query.data}")
    await callback_query.answer("Вы успешно получили WILL", show_alert=True)


@other_router.callback_query(F.data)
async def callback_query_send_will(callback_query: types.CallbackQuery):
    logging.info(f"Callback query: {callback_query.data}")
    user_id = callback_query.from_user.id
    
    data = callback_query.data.split()
    logging.info(f"Data: {data}")
    amount = data[1]
    recipient_username = data[2]
    sender_id = data[3]

    sender_balance = db.users.find_one({"telegram_id": sender_id}).get("balance", 0)

    if sender_balance < amount:
        await callback_query.answer("Недостаточно средств", show_alert=True)
    
    recipient = db.users.find_one({"username": recipient_username}).get("telegram_id", 0)

    if recipient == None:
        await callback_query.answer("Пользователь не найден", show_alert=True)
        
    db.users.update_one({"telegram_id": sender_id}, {"$inc": {"balance": -amount}})
    db.users.update_one({"telegram_id": recipient}, {"$inc": {"balance": amount}})

    await callback_query.answer()

    builder = InlineKeyboardBuilder()
    builder.add(types.InlineKeyboardButton(
        text="Получено",
        callback_data="received")
    )
    await callback_query.e.answer(
        f"Вы успешно отправили {amount} WILL пользователю @{recipient_username}", 
        reply_markup=builder.as_markup()
    )

@other_router.inline_query()
async def send_will(inline_query: InlineQuery):

    user_id = inline_query.from_user.id
    inline_query_text = inline_query.query
    inline_query_array = inline_query_text.split()

    user_balance = db.users.find_one({"telegram_id": user_id}).get("balance", 0)

    title = "Отправить "
    description = ""
    callback_data = f"send_{inline_query_text.replace(' ', '_')}_{user_id}"
    # callback_data = "received"
    logging.info(f"Callback data: {callback_data}")
    print(callback_data)

    builder = InlineKeyboardBuilder()
    builder.add(types.InlineKeyboardButton(
        text="Получить",
        callback_data=callback_data)
    )


    if len(inline_query_array) == 0:
        description = "Укажите количество WILL"
    elif len(inline_query_array) == 1:
        try:
            amount = int(inline_query_array[0])
            if amount <= 0:
                description = "Количество должно быть положительным числом"
            else:
                title = f"Отправить {amount} WILL"
                description = f"Вы хотите отправить {inline_query_array[0]} WILL\nУкажите юзернейм получателя"
        except Exception as e:
            description = "Количество должно быть числом"

    elif len(inline_query_array) == 2:
        try:
            amount = int(inline_query_array[0])
            logging.info(f"Amount: {amount}")
            if amount <= 0:
                description = "Количество должно быть положительным числом"
            else:
                user_balance = db.users.find_one({"telegram_id": user_id}).get("balance", 0)
                recipient = db.users.find_one({"username": inline_query_array[1][1:]})

                if user_balance < amount:
                    description = f"У вас недостаточно средств для отправки {amount} WILL"
                elif recipient == None:
                    description = f"Пользователь @{inline_query_array[1]} не найден"
                else:
                    description = f"Вы хотите отправить {amount} WILL\nПолучатель: {inline_query_array[1]}"
                    title = f"Отправить {amount} WILL"
        except Exception as e:
            print(e)
            description = "Количество должно быть числом"

    else:
        description = f"Некорректный ввод"



    result = [InlineQueryResultArticle(
        id=str(uuid.uuid4()),
        title=title,
        description=description,
        input_message_content=InputTextMessageContent(
            message_text=f"Вы хотите отправить {user_balance} WILL"
        ),
        reply_markup=builder.as_markup()
    )]
    
    await inline_query.answer(result, is_personal=True)
