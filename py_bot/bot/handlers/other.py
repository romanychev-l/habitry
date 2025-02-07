import logging
from datetime import datetime, timedelta
from aiogram import Router, F, types, Bot
from aiogram.types import Message, ReplyKeyboardRemove, ReplyKeyboardMarkup, KeyboardButton
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram_dialog import DialogManager, StartMode
from fluentogram import TranslatorRunner
from zoneinfo import ZoneInfo

from bot.config_data.config import db
from bot.states.start import StartSG

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
            f"✅ Выполнено привычек сегодня: {completed_today}\n"
            f"✅ Выполнено привычек вчера: {completed_yesterday}\n"
            f"🏆 Общее количество выполнений: {total_completions}\n"
        )
        
        await msg.answer(stats_message)
        
    except Exception as e:
        logging.error(f"Error in stats command: {e}")
        await msg.answer(f"Ошибка при получении статистики: {str(e)}")


@other_router.message(Actions.base_state)
async def registration_start(msg: Message, state: FSMContext, i18n: TranslatorRunner, dialog_manager: DialogManager):
    # await state.update_data(status='employee')
    
    await msg.answer(i18n.message.ex_one(), reply_markup=ReplyKeyboardRemove())


@other_router.message(Command("buy"))
async def cmd_buy(msg: Message, bot: Bot):
    await bot.send_invoice(
        chat_id=msg.chat.id,
        title="1 Telegram Stars",
        description="Покупка Stars для поддержки канала",
        payload="stars_1",
        provider_token="",  # Пустой токен для Stars
        currency="XTR",
        prices=[
            types.LabeledPrice(
                label="1 Stars",
                amount=1  # 100 Stars = 10000
            )
        ],
        need_name=False,
        need_phone_number=False,
        need_email=False,
        need_shipping_address=False,
        is_flexible=False
    )

# Обработчик pre_checkout_query
@other_router.pre_checkout_query()
async def process_pre_checkout_query(pre_checkout_query: types.PreCheckoutQuery, bot: Bot):
    try:
        await bot.answer_pre_checkout_query(
            pre_checkout_query_id=pre_checkout_query.id,
            ok=True
        )
        logging.info(f"Pre-checkout query processed: {pre_checkout_query.id}")
    except Exception as e:
        logging.error(f"Error in pre_checkout_query: {e}")

# Обработчик успешного платежа
@other_router.message(F.successful_payment)
async def successful_payment(message: types.Message, i18n: TranslatorRunner):
    print(message)
    try:
        user_id = message.from_user.id
        stars_amount = message.successful_payment.total_amount
        will_tokens = stars_amount * 10  # 1 Stars = 10 WILL

        # Получаем текущий баланс пользователя
        user = db.users_collection.find_one({"telegram_id": user_id})
        current_balance = user.get('balance', 0) if user else 0
        new_balance = current_balance + will_tokens

        result = db.users_collection.update_one(
            {"telegram_id": user_id},
            {"$set": {"balance": new_balance}}
        )
        
        if result.modified_count > 0:
            await message.answer(
                f"{i18n.message.payment_success()} {stars_amount} Stars\n"
                f"Начислено: {will_tokens} WILL"
            )
        else:
            logging.error(f"Пользователь не найден: {user_id}")
            # await message.answer(i18n.message.error_payment())
            
    except Exception as e:
        logging.error(f"Ошибка при обновлении БД: {e}")
        await message.answer(i18n.message.error_payment())

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
