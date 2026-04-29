package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func getMainKeyboard() tgbotapi.ReplyKeyboardMarkup {

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Уведомления дня"),
			tgbotapi.NewKeyboardButton("Чек-лист смены"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Актуальные лимиты"),
			tgbotapi.NewKeyboardButton("Скрипты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Посчитать балансы"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func notificationKeyboard() tgbotapi.InlineKeyboardMarkup {

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Отправить", "send"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Переписать", "rewrite"),
		),
	)
}
func getAdminKeyboard() tgbotapi.ReplyKeyboardMarkup {

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Новое уведомление команды"),
			tgbotapi.NewKeyboardButton("Уведомления дня"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Чек-лист смены"),
			tgbotapi.NewKeyboardButton("Актуальные лимиты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Скрипты"),
			tgbotapi.NewKeyboardButton("Посчитать балансы"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}
func checklistMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать чеклист смены", "edit_checklist_menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

func checklistEditKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Отправить", "save_checklist"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Переписать", "rewrite_checklist"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}
func limitsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Intent", "show_intent"),
			tgbotapi.NewInlineKeyboardButtonData("🤝 P2P", "show_p2p"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Обновить лимиты", "edit_limits"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

func limitsEditKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Intent", "copy_intent"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤝 P2P", "copy_p2p"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}
func scriptsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Intent", "scripts_intent"),
			tgbotapi.NewInlineKeyboardButtonData("🤝 P2P", "scripts_p2p"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

func intentScriptsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎰 1win", "script_1win"),
			tgbotapi.NewInlineKeyboardButtonData("⚽ 4ra", "script_4ra"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "scripts_back"),
		),
	)
}
