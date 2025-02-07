import logging
from typing import Dict, List
from datetime import datetime, timedelta
from bson import ObjectId

from aiogram import Bot
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from bot.config_data.config import db

logger = logging.getLogger(__name__)

class CountManager:
    def __init__(self, bot: Bot):
        self.bot = bot
        self.scheduler = AsyncIOScheduler()

    async def calculate_daily_rewards(self):
        """Подсчитывает и распределяет выигрыши за день"""
        try:
            yesterday_date = (datetime.now() - timedelta(days=1)).date()
            logger.info(f"yesterday_date: {yesterday_date}")
            
            unfulfilled_habits = get_unfulfilled_habits_with_stake(yesterday_date)
            
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
                    logger.info(is_habit_completed(yesterday_date, follower_habit['telegram_id'], follower_habit['_id']))
                    if (follower_habit and 
                        str(habit['_id']) in follower_habit['followers'] and 
                        is_habit_completed(yesterday_date, follower_habit['telegram_id'], follower_habit['_id'])):
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

                for follower in habit['followers']:
                    follower_id, follower_telegram_id, follower_stake = follower  # распаковываем список
                    win_amount = (stake_of_owner / sum_stakes_of_followers) * follower_stake
                    if follower_telegram_id not in winnings:
                        winnings[follower_telegram_id] = win_amount
                    else:
                        winnings[follower_telegram_id] += win_amount
                
                db.users.update_one({"telegram_id": habit['telegram_id']}, 
                                    {"$inc": {"balance": -stake_of_owner}})


            logger.info(f"winnings: {winnings}")

            for key, value in winnings.items():
                db.users.update_one({"telegram_id": key}, 
                                    {"$inc": {"balance": value}})

            logger.info("Daily rewards calculated and distributed successfully")
        except Exception as e:
            logger.error(f"Failed to calculate daily rewards: {e}")

    def start(self):
        """Запускает планировщик для ежедневного подсчета в начале дня"""
        # Устанавливаем задачу на выполнение каждый день в 00:05 UTC
        self.scheduler.add_job(
            self.calculate_daily_rewards,
            CronTrigger(hour=0, minute=5),
            id='calculate_daily_rewards'
        )
        self.scheduler.start()
        logger.info("Count manager scheduler started")

def is_habit_completed(date, telegram_id, habit_id):
    """
    Проверяет, была ли привычка выполнена в указанную дату
    """
    date_str = date.strftime("%Y-%m-%d")
    logger.info(habit_id)
    logger.info(type(habit_id))
    # habit_id = ObjectId(habit_id)
    # logger.info(habit_id)
    
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

def get_unfulfilled_habits_with_stake(yesterday_date):
    """
    Получает все привычки со ставкой, которые не были выполнены вчера
    """
    yesterday_weekday = yesterday_date.isoweekday() - 1
    logger.info(f"yesterday_weekday: {yesterday_weekday}")
    habits = list(db.habits.find({
        "stake": {"$gt": 0},
        "days": yesterday_weekday
    }))
    
    # Фильтруем привычки, оставляя только невыполненные
    return [
        habit for habit in habits 
        if not is_habit_completed(yesterday_date, habit['telegram_id'], habit['_id'])
    ]
    