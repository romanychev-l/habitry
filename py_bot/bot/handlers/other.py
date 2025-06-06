import logging
from datetime import datetime
from aiogram import Router, types, Bot
from aiogram.types import Message
from aiogram.filters import Command
from fluentogram import TranslatorRunner

from bot.config_data.config import db


other_router = Router()


# @other_router.message(Command("buy"))
# async def cmd_buy(msg: Message, bot: Bot):
#     await bot.send_invoice(
#         chat_id=msg.chat.id,
#         title="1 Telegram Stars",
#         description="Покупка Stars для поддержки канала",
#         payload="stars_1",
#         provider_token="",  # Пустой токен для Stars
#         currency="XTR",
#         prices=[
#             types.LabeledPrice(
#                 label="1 Stars",
#                 amount=1  # 100 Stars = 10000
#             )
#         ],
#         need_name=False,
#         need_phone_number=False,
#         need_email=False,
#         need_shipping_address=False,
#         is_flexible=False
#     )

# Обработчик pre_checkout_query
# @other_router.pre_checkout_query()
# async def process_pre_checkout_query(pre_checkout_query: types.PreCheckoutQuery, bot: Bot):
#     try:
#         await bot.answer_pre_checkout_query(
#             pre_checkout_query_id=pre_checkout_query.id,
#             ok=True
#         )
#         logging.info(f"Pre-checkout query processed: {pre_checkout_query.id}")
#     except Exception as e:
#         logging.error(f"Error in pre_checkout_query: {e}")

# Обработчик успешного платежа
# @other_router.message(F.successful_payment)
# async def successful_payment(message: types.Message, i18n: TranslatorRunner):
#     print(message)
#     try:
#         user_id = message.from_user.id
#         stars_amount = message.successful_payment.total_amount
#         will_tokens = stars_amount * 10  # 1 Stars = 10 WILL

#         # Получаем текущий баланс пользователя
#         user_record = db.users.find_one({"telegram_id": user_id}) # Исправлено users_collection на users
#         current_balance = user_record.get('balance', 0) if user_record else 0
#         new_balance = current_balance + will_tokens

#         result = db.users.update_one( # Исправлено users_collection на users
#             {"telegram_id": user_id},
#             {"$set": {"balance": new_balance}}
#         )
        
#         if result.modified_count > 0:
#             await message.answer(
#                 f"{i18n.message.payment_success()} {stars_amount} Stars\n"
#                 f"Начислено: {will_tokens} WILL"
#             )
#         else:
#             # Если пользователь не найден, создаем его
#             if not user_record:
#                  db.users.insert_one({
#                      "telegram_id": user_id,
#                      "username": message.from_user.username or f"user_{user_id}",
#                      "balance": will_tokens,
#                      "created_at": datetime.utcnow()
#                  })
#                  logging.info(f"New user created via payment: {user_id}")
#                  await message.answer(
#                      f"{i18n.message.payment_success()} {stars_amount} Stars\n"
#                      f"Начислено: {will_tokens} WILL"
#                  )
#             else:
#                  logging.error(f"Failed to update balance for user {user_id}")
#                  # await message.answer(i18n.message.error_payment()) # Можно раскомментировать, если нужно явное сообщение об ошибке

#     except Exception as e:
#         logging.error(f"Ошибка при обновлении БД: {e}")
#         await message.answer(i18n.message.error_payment())
