import logging
from datetime import datetime
from aiogram import Router, F, types, Bot
from aiogram.types import Message, InputTextMessageContent, InlineQuery, InlineQueryResultArticle
from aiogram.filters import Command
from fluentogram import TranslatorRunner
import uuid
from aiogram.utils.keyboard import InlineKeyboardBuilder

from bot.config_data.config import db

f2f_router = Router()


@f2f_router.message(Command("send"))
async def cmd_send_will(msg: Message, i18n: TranslatorRunner, bot: Bot):
    try:
        args = msg.text.split()
        logging.info(f"Received send command with args: {args}")
        if len(args) != 3:
            await msg.answer(i18n.send.usage())
            return

        _, recipient_username, amount_str = args
        
        # Убираем @ если есть
        recipient_username = recipient_username.lstrip('@')

        try:
            amount = int(amount_str)
            if amount <= 0:
                raise ValueError("Amount must be positive")
        except ValueError:
            await msg.answer(i18n.send.invalid_amount())
            return

        sender_id = msg.from_user.id
        sender = db.users.find_one({"telegram_id": sender_id})

        if not sender:
            logging.error(f"Sender user not found: {sender_id}")
            await msg.answer(i18n.send.sender_not_found())
            return

        sender_balance = sender.get('balance', 0)

        if sender_balance < amount:
            await msg.answer(i18n.send.insufficient_funds(balance=sender_balance))
            return

        recipient = db.users.find_one({"username": recipient_username})

        if not recipient:
            await msg.answer(i18n.send.user_not_found(username=recipient_username))
            return
            
        recipient_id = recipient.get("telegram_id")
        if sender_id == recipient_id:
            await msg.answer(i18n.send.send_to_self())
            return

        recipient_balance = recipient.get('balance', 0)

        # Обновляем балансы
        db.users.update_one({"telegram_id": sender_id}, {"$inc": {"balance": -amount}})
        db.users.update_one({"telegram_id": recipient_id}, {"$inc": {"balance": amount}})
        
        # Логируем транзакцию (опционально, можно создать отдельную коллекцию)
        db.transactions.insert_one({
            "sender_id": sender_id,
            "recipient_id": recipient_id,
            "recipient_username": recipient_username,
            "amount": amount,
            "timestamp": datetime.utcnow()
        })

        # Уведомление отправителю
        success_message = i18n.send.success_transfer(amount=amount, username=recipient_username)
        await msg.answer(success_message)

        # Уведомление получателю
        try:
            sender_username = msg.from_user.username or f"user_{sender_id}"
            recipient_text = i18n.send.success_received(amount=amount, username=sender_username)
            
            await bot.send_message(
                chat_id=recipient_id,
                text=recipient_text
            )
        except Exception as e:
            logging.error(f"Failed to notify recipient {recipient_id}: {e}")
            # Можно добавить логику возврата средств, если уведомление критично

        logging.info(f"WILL transfer: {sender_id} -> {recipient_id} ({recipient_username}), amount: {amount}")

    except Exception as e:
        logging.exception(f"Error in /send command: {e}")
        await msg.answer(i18n.error.generic())


@f2f_router.callback_query(F.data == "received")
async def callback_query_received(callback_query: types.CallbackQuery, i18n: TranslatorRunner):
    logging.info(f"Callback query: {callback_query.data}")
    await callback_query.answer(i18n.send.button_received(), show_alert=True)


@f2f_router.callback_query(F.data)
async def callback_query_send_will(callback_query: types.CallbackQuery, i18n: TranslatorRunner):
    logging.info(f"Callback query: {callback_query.data}")
    user_id = callback_query.from_user.id
    
    data = callback_query.data.split()
    logging.info(f"Data: {data}")
    amount = data[1]
    recipient_username = data[2]
    sender_id = data[3]

    sender_balance = db.users.find_one({"telegram_id": sender_id}).get("balance", 0)

    if sender_balance < amount:
        await callback_query.answer(i18n.send.insufficient_funds_short(), show_alert=True)
    
    recipient = db.users.find_one({"username": recipient_username}).get("telegram_id", 0)

    if recipient == None:
        await callback_query.answer(i18n.send.user_not_found(username=recipient_username), show_alert=True)
        
    db.users.update_one({"telegram_id": sender_id}, {"$inc": {"balance": -amount}})
    db.users.update_one({"telegram_id": recipient}, {"$inc": {"balance": amount}})

    await callback_query.answer()

    builder = InlineKeyboardBuilder()
    builder.add(types.InlineKeyboardButton(
        text=i18n.send.button_received(),
        callback_data="received")
    )
    await callback_query.message.answer(
        i18n.send.success_transfer(amount=amount, username=recipient_username), 
        reply_markup=builder.as_markup()
    )


@f2f_router.inline_query()
async def send_will(inline_query: InlineQuery, i18n: TranslatorRunner):

    user_id = inline_query.from_user.id
    inline_query_text = inline_query.query
    inline_query_array = inline_query_text.split()

    user_balance = db.users.find_one({"telegram_id": user_id}).get("balance", 0)

    title = i18n.send.inline_title_default()
    description = ""
    callback_data = f"send_{inline_query_text.replace(' ', '_')}_{user_id}"
    logging.info(f"Callback data: {callback_data}")

    builder = InlineKeyboardBuilder()
    builder.add(types.InlineKeyboardButton(
        text=i18n.send.inline_button_receive(),
        callback_data=callback_data)
    )

    if len(inline_query_array) == 0:
        description = i18n.send.inline_specify_amount()
    elif len(inline_query_array) == 1:
        try:
            amount = int(inline_query_array[0])
            if amount <= 0:
                description = i18n.send.invalid_amount()
            else:
                title = i18n.send.inline_title_send(amount=amount)
                description = i18n.send.inline_specify_recipient(amount=amount)
        except Exception as e:
            description = i18n.send.invalid_number()

    elif len(inline_query_array) == 2:
        try:
            amount = int(inline_query_array[0])
            logging.info(f"Amount: {amount}")
            if amount <= 0:
                description = i18n.send.invalid_amount()
            else:
                user_balance = db.users.find_one({"telegram_id": user_id}).get("balance", 0)
                recipient = db.users.find_one({"username": inline_query_array[1][1:]})

                if user_balance < amount:
                    description = i18n.send.insufficient_funds_short()
                elif recipient == None:
                    description = i18n.send.user_not_found(username=inline_query_array[1])
                else:
                    description = i18n.send.inline_send_confirmation(amount=amount, username=inline_query_array[1])
                    title = i18n.send.inline_title_send(amount=amount)
        except Exception as e:
            print(e)
            description = i18n.send.invalid_number()

    else:
        description = i18n.send.inline_invalid_input()

    result = [InlineQueryResultArticle(
        id=str(uuid.uuid4()),
        title=title,
        description=description,
        input_message_content=InputTextMessageContent(
            message_text=i18n.send.inline_message_template(balance=user_balance)
        ),
        reply_markup=builder.as_markup()
    )]
    
    await inline_query.answer(result, is_personal=True)
