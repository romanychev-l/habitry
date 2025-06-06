import logging
from datetime import datetime, timedelta
from aiogram import Router, F, types
from aiogram.types import Message
from aiogram.filters import Command
from openai import OpenAI
import json
from fluentogram import TranslatorRunner

from bot.config_data.config import db, config_settings

"""
AI API Functions:
- _call_deepseek_api(): запрос к DeepSeek API
- analyze_habits_with_ai(): анализ привычек с помощью DeepSeek

Команды:
- /analyze_habits - анализ привычек с DeepSeek
"""

ai_router = Router()

# Константа стоимости AI анализа
AI_ANALYSIS_COST = 10


async def check_and_charge_balance(user_id: int, amount: int, i18n) -> tuple[bool, int]:
    """
    Проверяет и списывает баланс пользователя
    
    Args:
        user_id: ID пользователя в Telegram
        amount: Сумма для списания
        i18n: Объект переводчика для локализации
    
    Returns:
        Tuple[bool, int]: (успешность операции, новый баланс)
    """
    try:
        # Получаем текущий баланс пользователя
        user = db.users.find_one({"telegram_id": user_id})
        if not user:
            return False, 0
            
        current_balance = user.get('balance', 0)
        
        # Проверяем достаточность средств
        if current_balance < amount:
            return False, current_balance
            
        # Списываем средства
        result = db.users.update_one(
            {"telegram_id": user_id},
            {"$inc": {"balance": -amount}}
        )
        
        if result.modified_count > 0:
            new_balance = current_balance - amount
            logging.info(f"Charged {amount} WILL from user {user_id}. New balance: {new_balance}")
            return True, new_balance
        else:
            logging.error(f"Failed to charge balance for user {user_id}")
            return False, current_balance
            
    except Exception as e:
        logging.error(f"Error charging balance for user {user_id}: {e}")
        return False, 0


async def _call_deepseek_api(prompt: str, model: str = "deepseek-chat", **kwargs) -> str:
    """
    Запрос к DeepSeek API
    """
    try:
        client = OpenAI(
            api_key=config_settings.DEEPSEEK_API_KEY.get_secret_value(),
            base_url="https://api.deepseek.com"
        )
        
        # Система сообщений для DeepSeek
        system_message = kwargs.get("system_message", "You are a helpful assistant specialized in habit analysis and personal development coaching.")
        
        response = client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": system_message},
                {"role": "user", "content": prompt},
            ],
            stream=False
        )
        
        if response and response.choices and len(response.choices) > 0:
            return response.choices[0].message.content
        else:
            logging.error("Empty response from DeepSeek")
            return None
            
    except Exception as e:
        logging.error(f"Error with DeepSeek API: {e}")
        return None


async def collect_week_habits_data(user_id: int, habits: list, i18n) -> dict:
    """
    Собирает данные о выполнении привычек за последние 7 дней
    """
    now = datetime.utcnow()
    week_data = {
        "user_id": user_id,
        "habits_info": [],
        "days_data": []
    }
    
    # Информация о привычках
    for habit in habits:
        week_data["habits_info"].append({
            "id": str(habit["_id"]),
            "title": habit["title"],
            "want_to_become": habit.get("want_to_become", ""),
            "days": habit.get("days", []),  # Дни недели когда должна выполняться
            "is_auto": habit.get("is_auto", False),
            "streak": habit.get("streak", 0),
            "score": habit.get("score", 0)
        })
    
    # Данные по дням за последнюю неделю
    for i in range(7):
        day_date = (now - timedelta(days=i)).date()
        day_str = day_date.strftime("%Y-%m-%d")
        weekday = day_date.weekday()  # 0 = понедельник, 6 = воскресенье
        
        # Получаем историю за этот день
        history = db.history.find_one({
            "telegram_id": user_id,
            "date": day_str
        })
        
        completed_habits = []
        if history and "habits" in history:
            completed_habits = [
                {
                    "habit_id": str(h["habit_id"]),
                    "title": h["title"],
                    "done": h.get("done", True)
                }
                for h in history["habits"]
                if h.get("done", True)
            ]
        
        # Определяем какие привычки должны были быть выполнены в этот день
        planned_habits = []
        for habit in habits:
            habit_days = habit.get("days", [])
            # Проверяем, запланирована ли привычка на этот день недели
            if weekday in habit_days:
                planned_habits.append({
                    "habit_id": str(habit["_id"]),
                    "title": habit["title"]
                })
        
        week_data["days_data"].append({
            "date": day_str,
            "weekday": weekday,
            "weekday_name": [
                i18n.ai.weekday_monday(),
                i18n.ai.weekday_tuesday(),
                i18n.ai.weekday_wednesday(),
                i18n.ai.weekday_thursday(),
                i18n.ai.weekday_friday(),
                i18n.ai.weekday_saturday(),
                i18n.ai.weekday_sunday()
            ][weekday],
            "planned_habits": planned_habits,
            "completed_habits": completed_habits,
            "completion_rate": len(completed_habits) / len(planned_habits) if planned_habits else 1.0
        })
    
    return week_data


async def analyze_habits_with_ai(week_data: dict, user_name: str, i18n) -> str:
    """
    Анализ привычек с помощью DeepSeek AI
    
    Args:
        week_data: Данные о привычках за неделю
        user_name: Имя пользователя
        i18n: Объект переводчика для локализации
    """
    try:
        # Получаем локализованные промпты
        week_data_json = json.dumps(week_data, ensure_ascii=False, indent=2)
        
        prompt = i18n.ai.analysis_prompt(
            userName=user_name,
            weekData=week_data_json
        )
        
        
        system_message = i18n.ai.system_message()
        
        # Запрос к DeepSeek API
        logging.info(f"Prompt: {prompt}")
        logging.info(f"System message: {system_message}")
        analysis = await _call_deepseek_api(
            prompt=prompt,
            system_message=system_message
        )
        
        return analysis
            
    except Exception as e:
        logging.error(f"Error analyzing habits with DeepSeek: {e}")
        return None


@ai_router.message(Command("analyze_habits"))
async def cmd_analyze_habits(msg: Message, i18n: TranslatorRunner):
    """
    Анализирует недельную статистику привычек пользователя с помощью DeepSeek AI
    """
    try:
        user_id = msg.from_user.id
        
        # Проверяем, есть ли пользователь в базе
        user = db.users.find_one({"telegram_id": user_id})
        if not user:
            await msg.answer(i18n.ai.user_not_found())
            return
        
        # Проверяем и списываем баланс за AI анализ
        payment_success, new_balance = await check_and_charge_balance(user_id, AI_ANALYSIS_COST, i18n)
        if not payment_success:
            current_balance = user.get('balance', 0)
            await msg.answer(i18n.ai.insufficient_balance(
                required=AI_ANALYSIS_COST,
                current=current_balance
            ))
            return
            
        # Уведомляем о списании средств
        await msg.answer(i18n.ai.payment_charged(
            amount=AI_ANALYSIS_COST,
            balance=new_balance
        ))
            
        await msg.answer(i18n.ai.collecting_data())
        
        # Получаем привычки пользователя
        habits = list(db.habits.find({"telegram_id": user_id}))
        if not habits:
            await msg.answer(i18n.ai.no_habits())
            return
            
        # Собираем данные за последние 7 дней
        week_data = await collect_week_habits_data(user_id, habits, i18n)
        
        if not week_data["days_data"]:
            await msg.answer(i18n.ai.insufficient_data())
            return
            
        # Отправляем данные в DeepSeek AI для анализа
        await msg.answer(i18n.ai.analyzing())
        
        analysis = await analyze_habits_with_ai(week_data, user.get('first_name', i18n.ai.default_username()), i18n)
        
        logging.info(f"Analysis length: {len(analysis) if analysis else 0}")
        if analysis:
            # Ограничиваем длину сообщения для Telegram
            analysis = analysis[:4096]

        # Отправляем результат анализа
        if analysis:
            await msg.answer(analysis)
        else:
            await msg.answer(i18n.ai.analysis_error())
            
    except Exception as e:
        logging.error(f"Error in analyze_habits command: {e}")
        await msg.answer(i18n.ai.command_error())


