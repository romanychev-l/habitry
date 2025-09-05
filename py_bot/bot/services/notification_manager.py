import logging
from typing import Optional

from aiogram import Bot, types
from aiogram.types import WebAppInfo
from aiogram.utils.keyboard import InlineKeyboardBuilder
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from bot.config_data.config import db

logger = logging.getLogger(__name__)

class NotificationManager:
    def __init__(self, bot: Bot):
        self.bot = bot
        self.scheduler = AsyncIOScheduler()
        self.jobs = {}  # Хранение джобов по user_id

    async def get_random_quote(self, language: str) -> str:
        try:
            # Получаем случайную цитату для указанного языка
            quotes = list(db.quotes.aggregate([
                {"$match": {language: {"$exists": True}}},
                {"$sample": {"size": 1}}
            ]))
            if quotes:
                quote_obj = quotes[0][language]
                quote_text = quote_obj['text']
                quote_author = quote_obj['autor']
                
                # Формируем строку с цитатой и автором
                if quote_author and quote_author != "null":
                    return f"{quote_text}\n\n@{quote_author}"
                return quote_text
                
            raise Exception("No quotes found")
        except Exception as e:
            logger.error(f"Failed to get random quote: {e}")
            # Возвращаем резервную цитату в случае ошибки
            if language == 'ru':
                return "🌟 Каждый день - это новая возможность стать лучше!"
            return "🌟 Every day is a new opportunity to become better!"

    async def send_notification(self, user_id: int, language: str):
        try:
            builder = InlineKeyboardBuilder()
            builder.row(types.InlineKeyboardButton(
                    text="Open",
                    web_app=WebAppInfo(url="https://lenichev.space/ht/")
                    # url="https://lenichev.space/ht/"
                )
            )

            quote = await self.get_random_quote(language)
            await self.bot.send_message(user_id, quote, reply_markup=builder.as_markup())
            logger.info(f"Notification sent to user {user_id}")
        except Exception as e:
            logger.error(f"Failed to send notification to user {user_id}: {e}")

    async def add_user_notification(
        self, 
        user_id: int, 
        time: str, 
        timezone: str, 
        language: str
    ) -> bool:
        """
        Добавляет или обновляет расписание уведомлений для пользователя
        
        :param user_id: ID пользователя в Telegram
        :param time: Время в формате "HH:MM"
        :param timezone: Часовой пояс пользователя
        :param language: Код языка пользователя
        :return: True если успешно, False если произошла ошибка
        """
        try:
            # Удаляем старое расписание если было
            self.remove_user_notification(user_id)
            
            # Создаем новое расписание
            hour, minute = map(int, time.split(':'))
            job = self.scheduler.add_job(
                self.send_notification,
                CronTrigger(hour=hour, minute=minute, timezone=timezone),
                args=[user_id, language],
                id=str(user_id),  # Используем user_id как идентификатор задачи
                replace_existing=True
            )
            self.jobs[user_id] = job
            logger.info(f"Added notification schedule for user {user_id} at {time} {timezone}")
            return True
        except Exception as e:
            logger.error(f"Failed to add notification for user {user_id}: {e}")
            return False

    def remove_user_notification(self, user_id: int) -> bool:
        """
        Удаляет расписание уведомлений для пользователя
        
        :param user_id: ID пользователя в Telegram
        :return: True если успешно, False если пользователь не найден
        """
        try:
            if user_id in self.jobs:
                self.jobs[user_id].remove()
                del self.jobs[user_id]
                logger.info(f"Removed notification schedule for user {user_id}")
                return True
            return False
        except Exception as e:
            logger.error(f"Failed to remove notification for user {user_id}: {e}")
            return False

    async def load_existing_notifications(self):
        """Загружает существующие настройки уведомлений из базы данных"""
        try:
            # Получаем всех пользователей с включенными уведомлениями
            users = list(db.users.find({
                "notifications_enabled": True,
                "notification_time": {"$exists": True, "$ne": ""}
            }))
            
            for user in users:
                await self.add_user_notification(
                    user["telegram_id"],
                    user["notification_time"],
                    user.get("timezone", "UTC"),
                    user.get("language_code", "en")
                )
            
            logger.info("Existing notifications loaded successfully")
        except Exception as e:
            logger.error(f"Failed to load existing notifications: {e}")

    def start(self):
        """Запускает планировщик и добавляет задачу на периодическую проверку обновлений"""
        # Добавляем задачу на периодическую проверку обновлений каждые 5 минут
        self.scheduler.add_job(
            self.load_existing_notifications,
            'interval',
            minutes=5,
            id='check_notifications_updates'
        )
        self.scheduler.start()
        logger.info("Notification scheduler started") 