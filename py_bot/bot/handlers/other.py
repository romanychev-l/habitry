import logging

from aiogram import Router, F, types, Bot
from aiogram.types import Message, ReplyKeyboardRemove, ReplyKeyboardMarkup, KeyboardButton
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram_dialog import DialogManager, StartMode
from fluentogram import TranslatorRunner

from bot.config_data.config import db
from bot.states.start import StartSG

other_router = Router()


class Actions(StatesGroup):
    base_state = State()


@other_router.message(Actions.base_state)
async def registration_start(msg: Message, state: FSMContext, i18n: TranslatorRunner, dialog_manager: DialogManager):
    # await state.update_data(status='employee')
    
    await msg.answer(i18n.message.ex_one(), reply_markup=ReplyKeyboardRemove())


@other_router.message(Command("buy"))
async def cmd_buy(msg: Message):
    await msg.send_invoice(
        chat_id=msg.chat.id,
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
@other_router.pre_checkout_query()
async def process_pre_checkout_query(pre_checkout_query: types.PreCheckoutQuery, bot: Bot):
    try:
        await bot.answer_pre_checkout_query(
            pre_checkout_query_id=pre_checkout_query.id,
            ok=True
        )
        logging.info(f"Pre-checkout query processed: {pre_checkout_query.id}")
    except Exception as e:
        logging.error(f"Error in pre_checkout_query: {e}")

# Обработчик успешного платежа
@other_router.message(F.successful_payment)
async def successful_payment(message: types.Message):
    print(message)
    try:
        user_id = message.from_user.id
        result = await db.users_collection.update_one(
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
