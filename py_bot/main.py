from aiogram import Bot, Dispatcher, types, F
from aiogram.filters import Command
import asyncio
import logging

# Инициализация бота
bot = Bot(token="1310848694:AAG9QRn_dO_WX6w2Hn7Y8W5Y7Wir4hOMsAU")
dp = Dispatcher()

# Обработчик команды /start
@dp.message(Command("start"))
async def cmd_start(message: types.Message):
    await message.answer("Привет! Используй /buy для покупки Stars")

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
    await message.answer(
        f"Спасибо за покупку! Оплачено: {message.successful_payment.total_amount} Stars"
    )

async def main():
    logging.basicConfig(level=logging.INFO)
    logging.info("Bot starting...")
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(main())
