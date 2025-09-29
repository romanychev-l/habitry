import logging
import asyncio
from datetime import datetime
from aiogram import Bot
from bson.objectid import ObjectId
from apscheduler.schedulers.asyncio import AsyncIOScheduler

from bot.config_data.config import config_settings, db

logger = logging.getLogger(__name__)

class PingManager:
    def __init__(self, bot: Bot):
        self.bot = bot
        self.db = db
        self.scheduler = AsyncIOScheduler()

    def start(self):
        """Запускает планировщик для обработки пингов."""
        self.scheduler.add_job(
            self._process_pending_pings,
            'interval',
            minutes=1,
            id='process_pending_pings'
        )
        self.scheduler.start()
        logger.info("Менеджер пингов запущен")

    def stop(self):
        """Останавливает планировщик пингов."""
        if self.scheduler.running:
            self.scheduler.shutdown()
            logger.info("Менеджер пингов остановлен")

    async def _process_pending_pings(self):
        """Обрабатывает ожидающие отправки пинги."""
        logger.info("Начинаем обработку пингов")
        try:
            # Получаем все ожидающие пинги
            pending_pings = list(self.db.pings.find({"status": "pending"}))
            logger.info(f"Найдено {len(pending_pings)} ожидающих пингов")
            
            # Обрабатываем каждый пинг
            for ping in pending_pings:
                await self._send_ping(ping)
        except Exception as e:
            logger.error(f"Ошибка при обработке пингов: {e}")

    async def _send_ping(self, ping):
        """Отправляет один пинг."""
        try:
            # Готовим сообщение для отправки
            sender_habit_id = ping.get("habit_id")
            sender_username = ping.get("sender_username", "пользователь")
            follower_id = ping.get("follower_id")
            
            if not follower_id or not sender_habit_id:
                logger.error(f"В пинге отсутствуют необходимые поля: {ping}")
                await self._update_ping_status(ping, "error")
                return
            
            # Находим привычку получателя, которая подписана на привычку отправителя
            follower_habit = self.db.habits.find_one({
                "telegram_id": follower_id,
                "followers": sender_habit_id
            })
            
            if not follower_habit:
                logger.error(f"Не найдена привычка получателя {follower_id}, подписанная на привычку отправителя {sender_habit_id}")
                await self._update_ping_status(ping, "error")
                return
            
            # Получаем название привычки получателя
            follower_habit_title = follower_habit.get("title", "привычка")
            
            # Формируем текст сообщения
            message_text = (
                f"💡 <b>Напоминание о привычке</b>\n\n"
                f"Пользователь @{sender_username} напоминает вам о необходимости выполнить привычку:\n"
                f"<b>{follower_habit_title}</b>\n\n"
                f"Не забудьте отметить выполнение сегодня!"
            )
            
            # Отправляем сообщение
            await self.bot.send_message(
                chat_id=follower_id,
                text=message_text,
                parse_mode="HTML"
            )
            
            # Обновляем статус пинга
            await self._update_ping_status(ping, "sent")
            logger.info(f"Пинг успешно отправлен пользователю {follower_id} для привычки '{follower_habit_title}'")
            
        except Exception as e:
            logger.error(f"Ошибка при отправке пинга {ping.get('_id')}: {e}")
            await self._update_ping_status(ping, "error")
    
    async def _update_ping_status(self, ping, status):
        """Обновляет статус пинга в базе данных."""
        try:
            if status == "sent":
                # Если пинг отправлен успешно, удаляем его из базы
                self.db.pings.delete_one({"_id": ping.get("_id")})
                logger.info(f"Пинг {ping.get('_id')} удален после успешной отправки")
            else:
                # В случае ошибки обновляем статус
                self.db.pings.update_one(
                    {"_id": ping.get("_id")},
                    {"$set": {"status": status}}
                )
                logger.info(f"Статус пинга {ping.get('_id')} обновлен на {status}")
        except Exception as e:
            logger.error(f"Ошибка при обновлении статуса пинга {ping.get('_id')}: {e}") 