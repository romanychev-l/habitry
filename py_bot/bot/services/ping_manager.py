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
        # Добавляем задачу проверки пингов каждые 5 минут
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
            habit_title = ping.get("habit_title", "привычка")
            sender_username = ping.get("sender_username", "пользователь")
            follower_id = ping.get("follower_id")
            
            if not follower_id:
                logger.error(f"В пинге отсутствует follower_id: {ping}")
                await self._update_ping_status(ping, "error")
                return
            
            # Формируем текст сообщения
            message_text = (
                f"💡 <b>Напоминание о привычке</b>\n\n"
                f"Пользователь @{sender_username} напоминает вам о необходимости выполнить привычку:\n"
                f"<b>{habit_title}</b>\n\n"
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
            logger.info(f"Пинг успешно отправлен пользователю {follower_id}")
            
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