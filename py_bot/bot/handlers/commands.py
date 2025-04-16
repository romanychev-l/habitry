import logging

from aiogram import Router, types
from aiogram.filters import CommandStart, Command, CommandObject
from aiogram.types import Message
from aiogram_dialog import DialogManager, StartMode
from aiogram.fsm.context import FSMContext
from fluentogram import TranslatorRunner
from aiogram.utils.keyboard import InlineKeyboardBuilder
from aiogram.types import WebAppInfo, FSInputFile
from bot.states.start import StartSG
from bot.config_data.config import config_settings

commands_router = Router()


# @commands_router.message(CommandStart())
# async def process_start_command(
#     msg: Message,
#     state: FSMContext, 
#     i18n: TranslatorRunner,
#     dialog_manager: DialogManager
# ) -> None:
#     # await dialog_manager.start(state=StartSG.start, mode=StartMode.RESET_STACK)
#     await msg.answer(i18n.message.start())
#     # await start_logic(msg, state, i18n)

@commands_router.message(Command("start"))
async def start(msg: types.Message, i18n: TranslatorRunner):
    # await msg.answer('https://pmpu.site/gfit?tg_id=' + str(msg.from_user.id))
    builder = InlineKeyboardBuilder()
    builder.row(
        types.InlineKeyboardButton(
            text=i18n.message.open(),
            # url="https://lenichev.site/ht_front_dev/",
            url="https://t.me/" + config_settings.BOT_USERNAME + "/app",
            # web_app=WebAppInfo(url="https://lenichev.site/ht_front_dev/"),
        )
    )
    # kb = [[KeyboardButton(text=Открыть)]]
    # keyboard = ReplyKeyboardMarkup(keyboard=kb, resize_keyboard=True, one_time_keyboard=True)

    # filename = "./bot/src/gigi.png"
    # await msg.answer_photo(
    #     photo=FSInputFile(filename),
    #     caption="Привет",
    #     reply_markup=builder.as_markup(),
    # )
    await msg.answer(
        text=i18n.message.start(),
        reply_markup=builder.as_markup()
    )