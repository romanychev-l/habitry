import asyncio
import logging

from aiogram import Bot, Dispatcher
from aiogram.client.default import DefaultBotProperties
from aiogram.enums import ParseMode
from aiogram_dialog import setup_dialogs
from fluentogram import TranslatorHub

from bot.config_data.config import config_settings
from bot.dialogs.start.dialogs import start_dialog
from bot.handlers.commands import commands_router
from bot.handlers.other import other_router
from bot.middlewares.i18n import TranslatorRunnerMiddleware
from bot.utils.i18n import create_translator_hub
from bot.services.notification_manager import NotificationManager
from bot.services.count_manager import CountManager
from bot.services.ping_manager import PingManager

logging.basicConfig(
    level=logging.INFO,
    format='[%(asctime)s] #%(levelname)-8s %(filename)s:'
           '%(lineno)d - %(name)s - %(message)s'
)

logger = logging.getLogger(__name__)


async def main() -> None:
    bot = Bot(
        token=config_settings.BOT_TOKEN.get_secret_value(),
        default=DefaultBotProperties(parse_mode=ParseMode.HTML)
    )
    # await bot.delete_webhook(drop_pending_updates=True)
    dp = Dispatcher()
    translator_hub: TranslatorHub = create_translator_hub()

    # Инициализируем и запускаем менеджер уведомлений
    notification_manager = NotificationManager(bot)
    notification_manager.start()
    # Загружаем существующие уведомления
    await notification_manager.load_existing_notifications()

    # Инициализируем и запускаем менеджер подсчета выигрышей
    count_manager = CountManager(bot, translator_hub)
    count_manager.start()
    
    # Инициализируем и запускаем менеджер пингов
    ping_manager = PingManager(bot)
    ping_manager.start()  # Теперь просто создаст асинхронную задачу

    dp.include_router(commands_router)
    dp.include_router(other_router)
    dp.include_router(start_dialog)

    dp.update.middleware(TranslatorRunnerMiddleware())

    setup_dialogs(dp)
    await dp.start_polling(bot, _translator_hub=translator_hub)


asyncio.run(main())