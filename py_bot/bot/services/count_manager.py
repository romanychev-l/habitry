import logging
from typing import Dict, List
from datetime import datetime, timedelta, timezone
from bson import ObjectId

from aiogram import Bot
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from bot.config_data.config import db

logger = logging.getLogger(__name__)

class CountManager:
    def __init__(self, bot: Bot):
        self.bot = bot
        # Явно указываем использование UTC для планировщика
        self.scheduler = AsyncIOScheduler(timezone='UTC')

    async def calculate_daily_rewards(self):
        """Подсчитывает и распределяет выигрыши за день"""
        try:
            # Используем UTC время и отступаем на 36 часов, чтобы гарантированно захватить "вчерашний" день
            # для всех часовых поясов (UTC+14 -> UTC-12 = 26 часов разницы, плюс запас)
            check_date = (datetime.now(timezone.utc) - timedelta(hours=36)).date()
            logger.info(f"Calculating rewards for date: {check_date} (UTC)")
            
            # Словарь для хранения информации о транзакциях каждого пользователя
            user_transactions = {}  # {telegram_id: {'sent': [], 'received': []}}
            
            unfulfilled_habits = get_unfulfilled_habits_with_stake(check_date)
            
            # Теперь можно работать с этими привычками дальше...
            logger.info(f"unfulfilled_habits: {unfulfilled_habits}")
            enriched_habits = []
            for habit in unfulfilled_habits:
                followers_data = []
                habit_name = habit.get('name', 'Без названия')
                for follower_id in habit['followers']:
                    follower_habit = db.habits.find_one({"_id": ObjectId(follower_id)})
                    logger.info(f"follower_habit: {follower_habit}")
                    logger.info(str(habit['_id']) in follower_habit['followers'])
                    logger.info(is_habit_completed(check_date, follower_habit['telegram_id'], follower_habit['_id']))
                    if (follower_habit and 
                        str(habit['_id']) in follower_habit['followers'] and 
                        is_habit_completed(check_date, follower_habit['telegram_id'], follower_habit['_id'])):
                        followers_data.append([
                            follower_habit['_id'],
                            follower_habit['telegram_id'],
                            follower_habit['stake'],
                            follower_habit.get('name', 'Без названия')
                        ])

                enriched_habit = habit
                enriched_habit['followers'] = followers_data
                enriched_habits.append(enriched_habit)

            logger.info(f"enriched_habits: {enriched_habits}")

            winnings = {}
            for habit in enriched_habits:
                sum_stakes_of_followers = sum(stake for _, _, stake, _ in habit['followers'])
                user_balance = db.users.find_one({"telegram_id": habit['telegram_id']})['balance']
                stake_of_owner = min(habit['stake'], user_balance)
                logger.info(f"stake_of_owner: {stake_of_owner}")
                logger.info(f"sum_stakes_of_followers: {sum_stakes_of_followers}")
                logger.info(f"user_balance: {user_balance}")

                # Если у владельца нет денег для ставки, пропускаем
                if stake_of_owner <= 0:
                    continue

                # Инициализируем запись для владельца привычки
                if habit['telegram_id'] not in user_transactions:
                    user_transactions[habit['telegram_id']] = {'sent': [], 'received': []}

                # Записываем информацию о списании у владельца
                user_transactions[habit['telegram_id']]['sent'].append({
                    'amount': stake_of_owner,
                    'habit_name': habit.get('name', 'Без названия')
                })

                db.users.update_one(
                    {"telegram_id": habit['telegram_id']}, 
                    {"$inc": {"balance": -stake_of_owner}}
                )

                if sum_stakes_of_followers <= 0:
                    db.settings.update_one(
                        {"_id": "system_settings"},
                        {"$inc": {"balance": stake_of_owner}},
                        upsert=True
                    )
                    continue

                total_distributed = 0
                for follower in habit['followers']:
                    follower_id, follower_telegram_id, follower_stake, follower_habit_name = follower
                    if follower_stake <= 0:
                        continue
                    
                    win_amount = int((stake_of_owner / sum_stakes_of_followers) * follower_stake)
                    total_distributed += win_amount
                    
                    # Инициализируем запись для получателя
                    if follower_telegram_id not in user_transactions:
                        user_transactions[follower_telegram_id] = {'sent': [], 'received': []}
                    
                    # Записываем информацию о получении
                    user_transactions[follower_telegram_id]['received'].append({
                        'amount': win_amount,
                        'from_habit': habit.get('name', 'Без названия'),
                        'for_habit': follower_habit_name
                    })
                    
                    if follower_telegram_id not in winnings:
                        winnings[follower_telegram_id] = win_amount
                    else:
                        winnings[follower_telegram_id] += win_amount

                # Добавляем остаток от округления в settings.balance
                remaining = stake_of_owner - total_distributed
                if remaining > 0:
                    db.settings.update_one(
                        {"_id": "system_settings"},
                        {"$inc": {"balance": remaining}},
                        upsert=True
                    )

            # Обновляем балансы и отправляем уведомления
            for telegram_id, transactions in user_transactions.items():
                if telegram_id in winnings:
                    db.users.update_one(
                        {"telegram_id": telegram_id}, 
                        {"$inc": {"balance": winnings[telegram_id]}}
                    )
                
                # Формируем сообщение для пользователя
                message_parts = []
                
                if transactions['sent']:
                    sent_text = "📤 Списания:\n"
                    for tx in transactions['sent']:
                        sent_text += f"- {tx['amount']} токенов за привычку '{tx['habit_name']}'\n"
                    message_parts.append(sent_text)
                
                if transactions['received']:
                    received_text = "📥 Получено:\n"
                    for tx in transactions['received']:
                        received_text += f"- {tx['amount']} токенов от '{tx['from_habit']}' за выполнение '{tx['for_habit']}'\n"
                    message_parts.append(received_text)
                
                if message_parts:
                    total_sent = sum(tx['amount'] for tx in transactions['sent'])
                    total_received = sum(tx['amount'] for tx in transactions['received'])
                    summary = f"💰 Итого: -{total_sent} / +{total_received} токенов"
                    message_parts.append(summary)
                    
                    full_message = "\n\n".join(message_parts)
                    try:
                        await self.bot.send_message(
                            telegram_id,
                            f"Отчёт о движении токенов за {check_date}:\n\n{full_message}"
                        )
                    except Exception as e:
                        logger.error(f"Failed to send message to user {telegram_id}: {e}")

            logger.info("Daily rewards calculated and distributed successfully")
        except Exception as e:
            logger.error(f"Failed to calculate daily rewards: {e}")

    def start(self):
        """Запускает планировщик для ежедневного подсчета"""
        # Устанавливаем задачу на выполнение каждый день в 10:05 UTC
        self.scheduler.add_job(
            self.calculate_daily_rewards,
            CronTrigger(hour=10, minute=5),
            id='calculate_daily_rewards'
        )
        self.scheduler.start()
        logger.info("Count manager scheduler started - will run daily at 10:05 UTC")

def is_habit_completed(date, telegram_id, habit_id):
    """
    Проверяет, была ли привычка выполнена в указанную дату
    """
    date_str = date.strftime("%Y-%m-%d")
    logger.info(f"Checking habit completion for date {date_str}, user {telegram_id}, habit {habit_id}")
    
    history = db['history'].find_one({
        "telegram_id": telegram_id,
        "date": date_str,
        "habits": {
            "$elemMatch": {
                "habit_id": habit_id,
                "done": True
            }
        }
    })
    logger.info(f"history: {history}")
    
    return history is not None

def get_unfulfilled_habits_with_stake(check_date):
    """
    Получает все привычки со ставкой, которые не были выполнены в указанную дату
    """
    check_weekday = check_date.isoweekday() - 1
    logger.info(f"Checking for unfulfilled habits on weekday: {check_weekday}, date: {check_date}")
    habits = list(db.habits.find({
        "stake": {"$gt": 0},
        "days": check_weekday
    }))
    
    # Фильтруем привычки, оставляя только невыполненные
    return [
        habit for habit in habits 
        if not is_habit_completed(check_date, habit['telegram_id'], habit['_id'])
    ]
    