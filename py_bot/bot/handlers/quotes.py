import logging
from aiogram import Router
from aiogram.types import Message
from aiogram.filters import Command

from bot.config_data.config import db

quotes_router = Router()


@quotes_router.message(Command("add_quote"))
async def cmd_add_quote(msg: Message):
    # Проверка на администратора
    if msg.from_user.id != 248603604:  # Ваш ID
        logging.info(f"Unauthorized quote addition attempt from user {msg.from_user.id}")
        return
        
    try:
        # Парсим сообщение
        lines = msg.text.split('\n')
        if len(lines) < 2:
            await msg.answer(
                "Формат:\n"
                "/add_quote\n"
                "ru\n"
                "Текст цитаты на русском\n"
                "Автор (опционально)\n"
                "en\n"
                "Text in English\n"
                "Author (optional)"
            )
            return
            
        quote_doc = {}
        current_lang = None
        current_text = None
        current_autor = None
        
        # Пропускаем первую строку с командой
        for line in lines[1:]:
            line = line.strip()
            if not line:
                continue
                
            if line in ['ru', 'en']:
                # Если был предыдущий язык, сохраняем его данные
                if current_lang and current_text:
                    quote_doc[current_lang] = {
                        "text": current_text,
                        "autor": current_autor  # Будет обновлено, если найдем автора
                    }
                current_lang = line
                current_text = None
                current_autor = None
            elif current_lang and not current_text:
                current_text = line
            elif current_lang and current_text:
                current_autor = line
        
        # Сохраняем последний блок
        if current_lang and current_text:
            if current_lang not in quote_doc:
                quote_doc[current_lang] = {
                    "text": current_text,
                    "autor": current_autor
                }
        
        if not quote_doc:
            raise ValueError("No valid quotes found in message")
            
        # Добавляем в базу данных
        result = db.quotes.insert_one(quote_doc)
        
        # Формируем ответное сообщение
        response = "Цитата успешно добавлена!\n"
        
        await msg.answer(response)
        
    except Exception as e:
        logging.error(f"Error adding quote: {e}")
        await msg.answer("Ошибка при добавлении цитаты. Проверьте формат команды.")
