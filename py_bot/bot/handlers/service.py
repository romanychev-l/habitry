import logging
from datetime import datetime, timedelta
from aiogram import Router
from aiogram.types import Message
from aiogram.filters import Command
from fluentogram import TranslatorRunner

from bot.config_data.config import db, config_settings

service_router = Router()

@service_router.message(Command("stats"))
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

    
@service_router.message(Command("today_list_users"))
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


@service_router.message(Command("add_will"))
async def cmd_add_will(msg: Message, i18n: TranslatorRunner):
    # Проверка на администратора
    if msg.from_user.id != 248603604:  # Ваш ID
        logging.info(f"Unauthorized add_will attempt from user {msg.from_user.id}")
        return
        
    try:
        args = msg.text.split()
        if len(args) != 2:
            await msg.answer(i18n.add.will_usage())
            return
            
        try:
            amount = int(args[1])
            if amount <= 0:
                await msg.answer(i18n.add.will_invalid_amount())
                return
        except ValueError:
            await msg.answer(i18n.add.will_invalid_format())
            return
            
        admin_id = msg.from_user.id
        
        # Получаем текущий баланс администратора
        admin_user = db.users.find_one({"telegram_id": admin_id})
        
        if not admin_user:
            # Создаем запись администратора, если её нет
            db.users.insert_one({
                "telegram_id": admin_id,
                "username": msg.from_user.username or f"admin_{admin_id}",
                "balance": amount,
                "created_at": datetime.utcnow()
            })
            await msg.answer(i18n.add.will_admin_created(amount=amount))
            logging.info(f"Created admin profile with {amount} WILL")
        else:
            # Обновляем баланс
            current_balance = admin_user.get('balance', 0)
            new_balance = current_balance + amount
            
            db.users.update_one(
                {"telegram_id": admin_id}, 
                {"$inc": {"balance": amount}}
            )
            
            await msg.answer(i18n.add.will_balance_updated(amount=amount, balance=new_balance))
            logging.info(f"Added {amount} WILL to admin balance. New balance: {new_balance}")
            
    except Exception as e:
        logging.error(f"Error in add_will command: {e}")
        await msg.answer(i18n.add.will_error(error=str(e)))