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
            
            unfulfilled_habits = get_unfulfilled_habits_with_stake(check_date)
            
            # Теперь можно работать с этими привычками дальше...
            logger.info(f"unfulfilled_habits: {unfulfilled_habits}")
            enriched_habits = []
            for habit in unfulfilled_habits:
                followers_data = []
                logger.info(f"habit: {habit}")
                for follower_id in habit['followers']:
                    follower_habit = db.habits.find_one({"_id": ObjectId(follower_id)})
                    logger.info(f"follower_habit: {follower_habit}")
                    logger.info(str(habit['_id']) in follower_habit['followers'])
                    logger.info(is_habit_completed(check_date, follower_habit['telegram_id'], follower_habit['_id']))
                    if (follower_habit and 
                        str(habit['_id']) in follower_habit['followers'] and 
                        is_habit_completed(check_date, follower_habit['telegram_id'], follower_habit['_id'])):
                        followers_data.append([follower_habit['_id'], follower_habit['telegram_id'], follower_habit['stake']])

                enriched_habit = habit
                enriched_habit['followers'] = followers_data
                enriched_habits.append(enriched_habit)

            logger.info(f"enriched_habits: {enriched_habits}")

            winnings = {}
            for habit in enriched_habits:
                sum_stakes_of_followers = sum(stake for _, _, stake in habit['followers'])
                user_balance = db.users.find_one({"telegram_id": habit['telegram_id']})['balance']
                stake_of_owner = min(habit['stake'], user_balance)
                logger.info(f"stake_of_owner: {stake_of_owner}")
                logger.info(f"sum_stakes_of_followers: {sum_stakes_of_followers}")
                logger.info(f"user_balance: {user_balance}")

                # Если у владельца нет денег для ставки, пропускаем
                if stake_of_owner <= 0:
                    continue

                # Обновляем баланс владельца
                db.users.update_one({"telegram_id": habit['telegram_id']}, 
                                    {"$inc": {"balance": -stake_of_owner}})

                # Если нет последователей с ненулевой ставкой, добавляем ставку в settings.balance
                if sum_stakes_of_followers <= 0:
                    db.settings.update_one(
                        {"_id": "system_settings"},
                        {"$inc": {"balance": stake_of_owner}},
                        upsert=True
                    )
                    continue

                total_distributed = 0
                for follower in habit['followers']:
                    follower_id, follower_telegram_id, follower_stake = follower  # распаковываем список
                    # Пропускаем последователей с нулевой ставкой
                    if follower_stake <= 0:
                        continue
                    win_amount = int((stake_of_owner / sum_stakes_of_followers) * follower_stake)  # Округляем вниз
                    total_distributed += win_amount
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

            logger.info(f"winnings: {winnings}")

            for key, value in winnings.items():
                db.users.update_one({"telegram_id": key}, 
                                    {"$inc": {"balance": value}})

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
    