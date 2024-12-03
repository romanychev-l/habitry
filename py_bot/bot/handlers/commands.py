import logging

from aiogram import Router, types
from aiogram.filters import CommandStart, Command, CommandObject
from aiogram.types import Message
from aiogram_dialog import DialogManager, StartMode
from aiogram.fsm.context import FSMContext
from fluentogram import TranslatorRunner

from bot.states.start import StartSG

commands_router = Router()


@commands_router.message(CommandStart())
async def process_start_command(
    msg: Message,
    state: FSMContext, 
    i18n: TranslatorRunner,
    dialog_manager: DialogManager
) -> None:
    # await dialog_manager.start(state=StartSG.start, mode=StartMode.RESET_STACK)
    await msg.answer("👇 Привет! Нажимай на кнопку Tracker")
    # await start_logic(msg, state, i18n)