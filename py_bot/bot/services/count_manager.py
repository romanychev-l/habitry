import logging
from typing import Dict, List
from datetime import datetime, timedelta, timezone
from bson import ObjectId

from aiogram import Bot
from fluentogram import TranslatorHub
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from bot.config_data.config import db

logger = logging.getLogger(__name__)

class CountManager:
    def __init__(self, bot: Bot, translator_hub: TranslatorHub):
        self.bot = bot
        self.translator_hub = translator_hub
        # Явно указываем использование UTC для планировщика
        self.scheduler = AsyncIOScheduler(timezone='UTC')

    async def calculate_daily_rewards(self):
        """Подсчитывает и распределяет выигрыши за день"""
        logger.info("Starting daily reward calculation process.")
        try:
            # Используем UTC время и отступаем на 36 часов, чтобы гарантированно захватить "вчерашний" день
            # для всех часовых поясов (UTC+14 -> UTC-12 = 26 часов разницы, плюс запас)
            check_date = (datetime.now(timezone.utc) - timedelta(days=1)).date()
            logger.info(f"Calculating rewards for date: {check_date} (UTC)")
            
            # Словарь для хранения информации о транзакциях каждого пользователя
            user_transactions = {}  # {telegram_id: {'sent': [], 'received': []}}
            
            unfulfilled_habits = get_unfulfilled_habits_with_stake(check_date)
            logger.info(f"Found {len(unfulfilled_habits)} unfulfilled habits with stake for {check_date}.")
            
            # Теперь можно работать с этими привычками дальше...
            # logger.info(f"unfulfilled_habits: {unfulfilled_habits}") # Слишком многословно
            enriched_habits = []
            for habit in unfulfilled_habits:
                try: # Обработка ошибок для обогащения одной привычки
                    followers_data = []
                    habit_title = habit.get('title', 'Без названия')
                    habit_id = habit['_id']
                    logger.debug(f"Enriching habit '{habit_title}' ({habit_id})")
                    for follower_id_str in habit.get('followers', []):
                        try: # Обработка ошибок для одного подписчика
                            follower_id = ObjectId(follower_id_str)
                            follower_habit = db.habits.find_one({"_id": follower_id})
                            # logger.info(f"follower_habit: {follower_habit}") # Слишком многословно
                            # Проверяем, что подписчик действительно следует за этой привычкой (взаимность)
                            # и что подписчик выполнил свою привычку в check_date
                            if (follower_habit and
                                # Используем 'followers' для проверки взаимной подписки, как описано пользователем
                                str(habit['_id']) in follower_habit.get('followers', []) and
                                is_habit_completed(check_date, follower_habit['telegram_id'], follower_habit['_id'])):
                                followers_data.append([
                                    follower_habit['_id'],
                                    follower_habit['telegram_id'],
                                    follower_habit['stake'],
                                    follower_habit.get('title', 'Без названия')
                                ])
                        except Exception as e_follower:
                            logger.error(f"Error processing follower {follower_id_str} for habit {habit_id}: {e_follower}")
                            continue # Пропускаем этого подписчика

                    enriched_habit = habit
                    enriched_habit['followers'] = followers_data
                    enriched_habits.append(enriched_habit)
                except Exception as e_enrich:
                     logger.error(f"Error enriching habit {habit.get('_id', 'N/A')}: {e_enrich}")
                     continue # Пропускаем эту привычку

            logger.info(f"Successfully enriched {len(enriched_habits)} habits.")
            # logger.info(f"enriched_habits: {enriched_habits}") # Слишком многословно

            winnings = {} # {telegram_id: total_win_amount}
            total_system_profit = 0 # Собираем системную прибыль здесь

            for habit in enriched_habits:
                habit_id = habit['_id']
                habit_owner_id = habit['telegram_id']
                habit_title = habit.get('title', 'Без названия')
                logger.info(f"Processing habit '{habit_title}' ({habit_id}) owned by {habit_owner_id}")
                try: # Обработка ошибок для одной привычки
                    sum_stakes_of_followers = sum(stake for _, _, stake, _ in habit['followers'])
                    user = db.users.find_one({"telegram_id": habit_owner_id})
                    if not user:
                        logger.warning(f"Owner {habit_owner_id} not found for habit {habit_id}. Skipping.")
                        continue
                    user_balance = user.get('balance', 0)
                    # Получаем username владельца для сообщений
                    owner_username = user.get('username')
                    owner_display_name = f"@{owner_username}" if owner_username else f"ID: {habit_owner_id}"

                    stake_of_owner = min(habit['stake'], user_balance)
                    logger.info(f"Habit '{habit_title}' ({habit_id}): Owner {owner_display_name}, Stake: {stake_of_owner}, Followers total stake: {sum_stakes_of_followers}, Owner balance: {user_balance}")

                    # Если у владельца нет денег для ставки или ставка нулевая, пропускаем
                    if stake_of_owner <= 0:
                        logger.info(f"Owner {owner_display_name} has insufficient balance or zero stake for habit '{habit_title}' ({habit_id}). Skipping.")
                        continue

                    # Инициализируем запись для владельца привычки, если еще нет
                    if habit_owner_id not in user_transactions:
                        user_transactions[habit_owner_id] = {'sent': [], 'received': []}

                    # Записываем информацию о списании у владельца
                    # Пока не добавляем информацию о получателях, чтобы не усложнять
                    user_transactions[habit_owner_id]['sent'].append({
                        'amount': stake_of_owner,
                        'habit_title': habit_title
                    })
                    logger.debug(f"Recorded deduction of {stake_of_owner} from owner {owner_display_name} for habit '{habit_title}' ({habit_id}).")

                    # Списываем средства с владельца
                    db.users.update_one(
                        {"telegram_id": habit_owner_id},
                        {"$inc": {"balance": -stake_of_owner}}
                    )
                    logger.info(f"Decreased balance of owner {owner_display_name} by {stake_of_owner}.")

                    # Если нет подписчиков или их суммарная ставка 0, вся ставка уходит системе
                    if sum_stakes_of_followers <= 0:
                        logger.info(f"No active followers with stake for habit '{habit_title}' ({habit_id}). Full stake {stake_of_owner} goes to system.")
                        total_system_profit += stake_of_owner
                        continue # Переходим к следующей привычке

                    # Распределяем выигрыш между подписчиками
                    total_distributed_for_habit = 0
                    for follower in habit['followers']:
                        try: # Обработка ошибок для одного подписчика
                            follower_id, follower_telegram_id, follower_stake, follower_habit_title = follower
                            if follower_stake <= 0:
                                continue # Пропускаем подписчиков с нулевой ставкой

                            # Рассчитываем выигрыш пропорционально ставке подписчика
                            win_amount = int((stake_of_owner / sum_stakes_of_followers) * follower_stake)
                            if win_amount <= 0: # Пропускаем нулевые выигрыши
                                continue

                            total_distributed_for_habit += win_amount
                            logger.debug(f"Calculated win of {win_amount} for follower {follower_telegram_id} (habit '{follower_habit_title}') from owner {owner_display_name} (habit '{habit_title}').")

                            # Инициализируем запись для получателя, если еще нет
                            if follower_telegram_id not in user_transactions:
                                user_transactions[follower_telegram_id] = {'sent': [], 'received': []}

                            # Записываем информацию о получении
                            user_transactions[follower_telegram_id]['received'].append({
                                'amount': win_amount,
                                'from_user_display': owner_display_name, # Добавляем имя отправителя
                                'from_habit': habit_title, # Привычка владельца, с которой пришел выигрыш
                                'for_habit': follower_habit_title # Привычка подписчика, за которую выигрыш
                            })

                            # Обновляем общий выигрыш для этого подписчика
                            winnings[follower_telegram_id] = winnings.get(follower_telegram_id, 0) + win_amount

                        except Exception as e_follower_dist:
                            logger.error(f"Error distributing winnings to follower {follower_telegram_id} from habit {habit_id}: {e_follower_dist}")
                            continue # Пропускаем этого подписчика

                    # Добавляем остаток от округления в системную прибыль
                    remaining = stake_of_owner - total_distributed_for_habit
                    if remaining > 0:
                        logger.info(f"Adding remaining {remaining} from habit '{habit_title}' ({habit_id}) distribution to system profit.")
                        total_system_profit += remaining

                except Exception as e_habit:
                    logger.error(f"Error processing habit '{habit_title}' ({habit_id}) for owner {owner_display_name}: {e_habit}")
                    # Не прерываем цикл, просто переходим к следующей привычке
                    continue

            # Обновляем системный баланс один раз в конце
            if total_system_profit > 0:
                db.settings.update_one(
                    {"_id": "system_settings"},
                    {"$inc": {"balance": total_system_profit}},
                    upsert=True
                )
                logger.info(f"Updated system balance by +{total_system_profit}.")

            # Обновляем балансы победителей и отправляем уведомления
            logger.info(f"Starting balance updates and notifications for {len(user_transactions)} users.")
            for telegram_id, transactions in user_transactions.items():
                try: # Обработка ошибок для одного пользователя (обновление баланса + отправка сообщения)
                    # Получаем язык пользователя
                    user_data = db.users.find_one({"telegram_id": telegram_id}, {"language_code": 1})
                    # Определяем локаль, используя язык пользователя или 'en' по умолчанию
                    user_lang = user_data.get('language_code') if user_data else None
                    locale = 'ru' if user_lang == 'ru' else 'en' # По умолчанию 'en'
                    
                    # Получаем переводчик (TranslatorRunner) из хаба для нужной локали
                    translator = self.translator_hub.get_translator_by_locale(locale=locale)
                    
                    # Получаем переведенные fallback строки
                    unknown_habit_fallback = translator.get('report-fallback-unknown-habit')
                    unknown_user_fallback = translator.get('report-fallback-unknown-user')
                    unknown_habit_placeholder = translator.get('report-fallback-unknown-habit-placeholder')

                    win_amount = winnings.get(telegram_id, 0)
                    if win_amount > 0:
                        db.users.update_one(
                            {"telegram_id": telegram_id},
                            {"$inc": {"balance": win_amount}}
                        )
                        logger.info(f"Increased balance of user {telegram_id} by {win_amount}.")

                    # Формируем сообщение для пользователя с использованием переводов
                    message_parts = []
                    total_sent = sum(tx['amount'] for tx in transactions.get('sent', []))
                    total_received = sum(tx['amount'] for tx in transactions.get('received', []))

                    if not transactions.get('sent') and not transactions.get('received'):
                         logger.warning(f"No transactions recorded for user {telegram_id}, skipping notification.")
                         continue # Нет смысла отправлять пустое сообщение

                    if transactions.get('sent'):
                        # Используем ключ report-section-sent
                        sent_text = translator.get('report-section-sent') + "\n"
                        for tx in transactions['sent']:
                            # Используем ключ report-sent-item
                            habit_title = tx.get('habit_title', unknown_habit_fallback)
                            sent_text += translator.get(
                                'report-sent-item',
                                amount=tx['amount'],
                                habitTitle=habit_title
                            ) + "\n"
                        message_parts.append(sent_text.strip())

                    if transactions.get('received'):
                         # Используем ключ report-section-received
                        received_text = translator.get('report-section-received') + "\n"
                        for tx in transactions['received']:
                            # Используем ключ report-received-item
                            sender_info = tx.get('from_user_display', unknown_user_fallback)
                            from_habit = tx.get('from_habit', unknown_habit_placeholder)
                            for_habit = tx.get('for_habit', unknown_habit_placeholder)
                            received_text += translator.get(
                                'report-received-item',
                                amount=tx['amount'],
                                senderInfo=sender_info,
                                fromHabit=from_habit,
                                forHabit=for_habit
                            ) + "\n"
                        message_parts.append(received_text.strip())

                    # Используем ключ report-summary
                    summary = translator.get(
                        'report-summary',
                        date=check_date.strftime("%Y-%m-%d"), # Оставляем формат YYYY-MM-DD
                        totalSent=total_sent,
                        totalReceived=total_received
                    )
                    message_parts.append(summary)

                    # Используем ключ report-title для заголовка
                    message_title = translator.get('report-title', date=check_date.strftime("%Y-%m-%d"))

                    full_message = "\n\n".join(filter(None, message_parts))

                    if full_message:
                        try:
                            logger.info(f"Sending notification to user {telegram_id} in locale '{locale}'.")
                            await self.bot.send_message(
                                telegram_id,
                                f"{message_title}\n\n{full_message}" # Собираем заголовок и тело
                            )
                            logger.debug(f"Successfully sent notification to user {telegram_id}.")
                        except Exception as e_send:
                            logger.error(f"Failed to send message to user {telegram_id}: {e_send}")
                            # Продолжаем цикл, даже если сообщение не отправилось
                    else:
                         logger.warning(f"Generated empty message for user {telegram_id}, notification not sent.")

                except Exception as e_user_update:
                    logger.error(f"Error processing update/notification for user {telegram_id}: {e_user_update}")
                    continue # Переходим к следующему пользователю

            logger.info("Daily rewards calculation and distribution process finished successfully.")
        except Exception as e:
            logger.error(f"CRITICAL: Failed to calculate daily rewards due to an unexpected error: {e}", exc_info=True)

    def start(self):
        """Запускает планировщик для ежедневного подсчета"""
        # Устанавливаем задачу на выполнение каждый день в 10:05 UTC
        self.scheduler.add_job(
            self.calculate_daily_rewards,
            CronTrigger(hour=13, minute=5),
            id='calculate_daily_rewards'
        )
        self.scheduler.start()
        logger.info("Count manager scheduler started - will run daily at 10:05 UTC")

def is_habit_completed(date, telegram_id, habit_id):
    """
    Проверяет, была ли привычка выполнена в указанную дату
    """
    date_str = date.strftime("%Y-%m-%d")
    logger.info(f"Checking habit completion for date {date_str}, user {telegram_id}, habit {habit_id}")
    
    history = db['history'].find_one({
        "telegram_id": telegram_id,
        "date": date_str,
        "habits": {
            "$elemMatch": {
                "habit_id": habit_id,
                "done": True
            }
        }
    })
    logger.info(f"history: {history}")
    
    return history is not None

def get_unfulfilled_habits_with_stake(check_date):
    """
    Получает все привычки со ставкой, которые не были выполнены в указанную дату
    """
    check_weekday = check_date.isoweekday() - 1
    logger.info(f"Checking for unfulfilled habits on weekday: {check_weekday}, date: {check_date}")
    habits = list(db.habits.find({
        "stake": {"$gt": 0},
        "days": check_weekday
    }))
    
    # Фильтруем привычки, оставляя только невыполненные
    return [
        habit for habit in habits 
        if not is_habit_completed(check_date, habit['telegram_id'], habit['_id'])
    ]
    