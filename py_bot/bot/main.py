from aiogram import Bot, Dispatcher, types, F
from aiogram.filters import Command
import asyncio
import logging
from motor.motor_asyncio import AsyncIOMotorClient

# Подключение к MongoDB
MONGO_URI = "mongodb://localhost:27017"
client = AsyncIOMotorClient(MONGO_URI)
db = client.ht_db  # Используем ту же базу данных, что и в бэкенде
users_collection = db.users

# Инициализация бота
bot = Bot(token="1310848694:AAG9QRn_dO_WX6w2Hn7Y8W5Y7Wir4hOMsAU")
dp = Dispatcher()

# Обработчик команды /start
@dp.message(Command("start"))
async def cmd_start(message: types.Message):
    await message.answer("👇 Привет! Нажимай на кнопку Tracker")

# Обработчик команды /buy
@dp.message(Command("buy"))
async def cmd_buy(message: types.Message):
    await bot.send_invoice(
        chat_id=message.chat.id,
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
@dp.pre_checkout_query()
async def process_pre_checkout_query(pre_checkout_query: types.PreCheckoutQuery):
    try:
        await bot.answer_pre_checkout_query(
            pre_checkout_query_id=pre_checkout_query.id,
            ok=True
        )
        logging.info(f"Pre-checkout query processed: {pre_checkout_query.id}")
    except Exception as e:
        logging.error(f"Error in pre_checkout_query: {e}")

# Обработчик успешного платежа
@dp.message(F.successful_payment)
async def successful_payment(message: types.Message):
    print(message)
    try:
        user_id = message.from_user.id
        result = await users_collection.update_one(
            {"telegram_id": user_id},
            {"$set": {"credit": 0}}
        )
        
        if result.modified_count > 0:
            await message.answer(
                f"Спасибо за покупку! Оплачено: {message.successful_payment.total_amount} Stars"
            )
        else:
            logging.error(f"Пользователь не найден: {user_id}")
            await message.answer("Произошла ошибка при обработке платежа. Пожалуйста, обратитесь в поддержку.")
            
    except Exception as e:
        logging.error(f"Ошибка при обновлении БД: {e}")
        await message.answer("Произошла ошибка при обработке платежа. Пожалуйста, обратитесь в поддержку.")

async def main():
    logging.basicConfig(level=logging.INFO)
    logging.info("Bot starting...")
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(main())
