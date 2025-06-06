import pymongo 
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import SecretStr

class Settings(BaseSettings):
    BOT_TOKEN: SecretStr
    MONGO_HOST: str
    MONGO_PORT: int
    MONGO_DB_NAME: str
    BOT_USERNAME: str
    DEEPSEEK_API_KEY: SecretStr

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

config_settings = Settings()

mongo_client = pymongo.MongoClient(config_settings.MONGO_HOST, config_settings.MONGO_PORT) 
db = mongo_client[config_settings.MONGO_DB_NAME] 