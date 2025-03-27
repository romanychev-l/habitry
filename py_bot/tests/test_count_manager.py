import pytest
from datetime import datetime, timedelta, timezone
from unittest.mock import Mock, patch
from bson import ObjectId
import sys

# Создаем мок для db с поддержкой доступа через квадратные скобки
class MockDB:
    def __init__(self):
        self.history = Mock()
        self.habits = Mock()
        self.users = Mock()
        self.settings = Mock()
    
    def __getitem__(self, key):
        return getattr(self, key)

mock_db_global = MockDB()

# Создаем фейковый модуль для config
class FakeConfigModule:
    db = mock_db_global

# Добавляем фейковый модуль в sys.modules
sys.modules['bot.config_data.config'] = FakeConfigModule()

# Теперь можно импортировать count_manager
from bot.services.count_manager import CountManager, is_habit_completed, get_unfulfilled_habits_with_stake

@pytest.fixture
def mock_bot():
    return Mock()

@pytest.fixture
def mock_db():
    # Очищаем моки перед каждым тестом
    mock_db_global.history.reset_mock()
    mock_db_global.habits.reset_mock()
    mock_db_global.users.reset_mock()
    mock_db_global.settings = Mock()  # Добавляем мок для settings
    return mock_db_global

@pytest.fixture
def test_habits():
    return [
        {
            "_id": ObjectId("507f1f77bcf86cd799439011"),
            "telegram_id": 123456789,
            "stake": 100,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439012"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439012"),
            "telegram_id": 987654321,
            "stake": 50,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439011"]
        }
    ]

@pytest.fixture
def test_users():
    return [
        {
            "telegram_id": 123456789,
            "balance": 1000,
            "timezone": "Europe/Moscow"  # Добавляем timezone
        },
        {
            "telegram_id": 987654321,
            "balance": 500,
            "timezone": "America/New_York"  # Добавляем timezone
        }
    ]

@pytest.fixture
def test_history():
    # Используем UTC для единообразия
    check_date = (datetime.now(timezone.utc) - timedelta(hours=36)).date()
    return {
        "telegram_id": 987654321,
        "date": check_date.strftime("%Y-%m-%d"),
        "habits": [
            {
                "habit_id": ObjectId("507f1f77bcf86cd799439012"),
                "done": True
            }
        ]
    }

def test_is_habit_completed(mock_db, test_history):
    # Используем UTC для единообразия
    check_date = (datetime.now(timezone.utc) - timedelta(hours=36)).date()
    mock_db.history.find_one.return_value = test_history
    
    result = is_habit_completed(
        check_date,
        987654321,
        ObjectId("507f1f77bcf86cd799439012")
    )
    
    assert result is True
    mock_db.history.find_one.assert_called_once()

def test_get_unfulfilled_habits_with_stake(mock_db, test_habits):
    # Используем UTC для единообразия
    check_date = (datetime.now(timezone.utc) - timedelta(hours=36)).date()
    mock_db.habits.find.return_value = test_habits
    # Устанавливаем, что первая привычка не выполнена, а вторая выполнена
    mock_db.history.find_one.side_effect = lambda query: None if query["telegram_id"] == 123456789 else test_history
    
    result = get_unfulfilled_habits_with_stake(check_date)
    
    assert len(result) == 1
    assert result[0]["_id"] == test_habits[0]["_id"]

@pytest.mark.asyncio
async def test_calculate_daily_rewards(mock_db, mock_bot, test_habits, test_users, test_history):
    # Используем UTC для единообразия
    check_date = (datetime.now(timezone.utc) - timedelta(hours=36)).date()
    
    # Настраиваем моки
    mock_db.habits.find.return_value = test_habits
    mock_db.habits.find_one.side_effect = lambda query: next((h for h in test_habits if h["_id"] == query["_id"]), None)
    mock_db.users.find_one.side_effect = lambda query: next((u for u in test_users if u["telegram_id"] == query["telegram_id"]), None)
    # Устанавливаем, что первая привычка не выполнена, а вторая выполнена
    mock_db.history.find_one.side_effect = lambda query: None if query["telegram_id"] == 123456789 else test_history
    
    count_manager = CountManager(mock_bot)
    await count_manager.calculate_daily_rewards()
    
    # Проверяем обновление балансов
    mock_db.users.update_one.assert_called()
    
    # Проверяем, что выигрыш был правильно распределен
    update_calls = mock_db.users.update_one.call_args_list
    assert len(update_calls) > 0
    
    # Проверяем, что баланс владельца привычки уменьшился
    owner_update = next(call for call in update_calls if call[0][0] == {"telegram_id": 123456789})
    assert owner_update[0][1]["$inc"]["balance"] < 0
    
    # Проверяем, что баланс последователя увеличился
    follower_update = next(call for call in update_calls if call[0][0] == {"telegram_id": 987654321})
    assert follower_update[0][1]["$inc"]["balance"] > 0

@pytest.fixture
def complex_test_users():
    return [
        {
            "telegram_id": 1001,
            "balance": 1000,
            "name": "User1",
            "timezone": "Europe/Moscow"
        },
        {
            "telegram_id": 1002,
            "balance": 500,
            "name": "User2",
            "timezone": "America/New_York"
        },
        {
            "telegram_id": 1003,
            "balance": 200,
            "name": "User3",
            "timezone": "Asia/Tokyo"
        },
        {
            "telegram_id": 1004,
            "balance": 100,
            "name": "User4",
            "timezone": "Europe/London"
        },
        {
            "telegram_id": 1005,
            "balance": 1000,
            "name": "User5",
            "timezone": "Australia/Sydney"
        },
        {
            "telegram_id": 1006,
            "balance": 100,
            "name": "User6",
            "timezone": "Pacific/Auckland"
        }
    ]

@pytest.fixture
def complex_test_habits():
    return [
        {
            "_id": ObjectId("507f1f77bcf86cd799439001"),
            "telegram_id": 1001,
            "stake": 100,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439002", "507f1f77bcf86cd799439003", "507f1f77bcf86cd799439005", "507f1f77bcf86cd799439006"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439002"),
            "telegram_id": 1002,
            "stake": 50,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439001", "507f1f77bcf86cd799439003"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439003"),
            "telegram_id": 1003,
            "stake": 0,  # Нулевой стейк
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439001", "507f1f77bcf86cd799439002"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439004"),
            "telegram_id": 1004,
            "stake": 200,  # Стейк больше баланса
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439001"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439005"),
            "telegram_id": 1005,
            "stake": 100,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439001", "507f1f77bcf86cd799439006"]
        },
        {
            "_id": ObjectId("507f1f77bcf86cd799439006"),
            "telegram_id": 1006,
            "stake": 10,
            "days": [0, 1, 2, 3, 4, 5, 6],
            "followers": ["507f1f77bcf86cd799439001", "507f1f77bcf86cd799439005"]
        }
    ]

@pytest.fixture
def complex_test_history():
    yesterday = (datetime.now() - timedelta(days=1)).date()
    return {
        "telegram_id": 1002,  # User2 выполнил привычку
        "date": yesterday.strftime("%Y-%m-%d"),
        "habits": [
            {
                "habit_id": ObjectId("507f1f77bcf86cd799439002"),
                "done": True
            }
        ]
    }

@pytest.mark.asyncio
async def test_complex_daily_rewards(mock_db, mock_bot, complex_test_habits, complex_test_users, complex_test_history):
    yesterday = (datetime.now() - timedelta(days=1)).date()
    
    # Настраиваем моки
    mock_db.habits.find.return_value = complex_test_habits
    mock_db.habits.find_one.side_effect = lambda query: next((h for h in complex_test_habits if h["_id"] == query["_id"]), None)
    mock_db.users.find_one.side_effect = lambda query: next((u for u in complex_test_users if u["telegram_id"] == query["telegram_id"]), None)
    
    # Настраиваем проверку выполнения привычек
    def check_habit_completion(query):
        # User1 не выполнил привычку
        if query["telegram_id"] == 1001:
            return None
        # User2 выполнил привычку
        if query["telegram_id"] == 1002:
            return complex_test_history
        # User3 не выполнил привычку
        if query["telegram_id"] == 1003:
            return None
        # User4 не выполнил привычку
        if query["telegram_id"] == 1004:
            return None
        # User5 выполнил привычку
        if query["telegram_id"] == 1005:
            return {
                "telegram_id": 1005,
                "date": yesterday.strftime("%Y-%m-%d"),
                "habits": [
                    {
                        "habit_id": ObjectId("507f1f77bcf86cd799439005"),
                        "done": True
                    }
                ]
            }
        # User6 выполнил привычку
        if query["telegram_id"] == 1006:
            return {
                "telegram_id": 1006,
                "date": yesterday.strftime("%Y-%m-%d"),
                "habits": [
                    {
                        "habit_id": ObjectId("507f1f77bcf86cd799439006"),
                        "done": True
                    }
                ]
            }
        return None
    
    mock_db.history.find_one.side_effect = check_habit_completion
    
    # Запускаем расчет выигрышей
    count_manager = CountManager(mock_bot)
    await count_manager.calculate_daily_rewards()
    
    # Проверяем обновление балансов
    update_calls = mock_db.users.update_one.call_args_list
    
    # Собираем все обновления балансов
    balance_updates = {}
    for call in update_calls:
        query = call[0][0]
        update = call[0][1]
        user_id = query["telegram_id"]
        balance_updates[user_id] = update["$inc"]["balance"]

    print(balance_updates)
    
    # Проверяем обновления для каждого пользователя
    assert balance_updates[1001] < 0  # User1 потерял деньги (не выполнил привычку)
    assert balance_updates[1002] == 31  # User2 получил 31 токен (выполнил привычку)
    assert 1003 not in balance_updates  # User3 не участвует (нулевой стейк и не выполнил привычку)
    assert balance_updates[1004] < 0  # User4 потерял деньги (не выполнил привычку)
    assert balance_updates[1005] == 62  # User5 получил 62 токена (выполнил привычку)
    assert balance_updates[1006] == 6  # User6 получил 6 токенов (выполнил привычку)
    
    # Проверяем, что нераспределенные ставки пошли в settings.balance
    settings_updates = mock_db.settings.update_one.call_args_list
    print("Settings updates:")
    for update in settings_updates:
        print(f"Update: {update[0][1]['$inc']['balance']}")
    settings_balance_change = sum(update[0][1]["$inc"]["balance"] for update in settings_updates)
    print(f"Total settings balance change: {settings_balance_change}")
    assert settings_balance_change > 0  # Проверяем, что сумма положительная
    
    # Проверяем, что сумма всех изменений баланса равна 0 (деньги не создаются и не исчезают)
    total_balance_change = sum(balance_updates.values()) + settings_balance_change
    print(f"Total balance change: {total_balance_change}")
    assert abs(total_balance_change) < 0.01  # Учитываем возможные ошибки округления
    
    # Проверяем, что никто не может потерять больше, чем у него есть
    for user in complex_test_users:
        if user["telegram_id"] in balance_updates:
            assert user["balance"] + balance_updates[user["telegram_id"]] >= 0 