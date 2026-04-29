package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db          *sql.DB
	adminID     int64
	userState   = make(map[int64]string)
	tempMessage = make(map[int64]string)
	balanceStep = make(map[int64]int)
	balanceData = make(map[int64]map[string]float64)
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	botToken := os.Getenv("BOT_TOKEN")
	adminID = 8006127742

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Бот успешно запущен")

	db, err = sql.Open("sqlite3", "bot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables()
	go startScheduler(bot)
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update)
		}
		if update.CallbackQuery != nil {
			handleCallback(bot, update)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	userID := update.Message.From.ID
	username := update.Message.From.UserName
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	saveUser(userID, username)

	if text == "/start" {
		sendMainMenu(bot, chatID, userID)
		return
	}

	if userState[userID] == "waiting_notification" {

		tempMessage[userID] = text

		preview := "🔥 Новое уведомление от лида!\n\n" + text

		msg := tgbotapi.NewMessage(chatID, preview)
		msg.ReplyMarkup = notificationKeyboard()

		bot.Send(msg)

		userState[userID] = "preview"
		return
	}

	switch text {

	case "Новое уведомление команды":

		if !isAdmin(userID) {
			return
		}

		userState[userID] = "waiting_notification"

		bot.Send(tgbotapi.NewMessage(chatID, "Введи сообщение для команды"))

	case "Уведомления дня":
		sendTodayNotifications(bot, chatID)

	case "Чек-лист смены":
		if isAdmin(userID) {
			current := getChecklist()
			text := fmt.Sprintf("📋 **Актуальный чек-лист:**\n\n%s", current)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = checklistMenuKeyboard()
			bot.Send(msg)
		} else {
			sendChecklist(bot, chatID)
		}
		return

	case "Актуальные лимиты":
		msg := tgbotapi.NewMessage(chatID, "💰 **Выберите тип лимитов:**")
		msg.ReplyMarkup = limitsKeyboard()
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "Скрипты":
		msg := tgbotapi.NewMessage(chatID, "📜 **Выберите тип скриптов:**")
		msg.ReplyMarkup = scriptsKeyboard()
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "Посчитать балансы":
		balanceData[userID] = make(map[string]float64)
		balanceStep[userID] = 0
		bot.Send(tgbotapi.NewMessage(chatID, "💰 Лимиты провайдеров\n\nБади:"))

	case "edit_checklist":
		if !isAdmin(userID) {
			return
		}
		userState[userID] = "waiting_checklist"
		bot.Send(tgbotapi.NewMessage(chatID, "📝 Введи актуальную информацию для чек-листа"))

	}
	if userState[userID] == "waiting_checklist" {
		tempMessage[userID] = text
		now := time.Now().Format("02.01.2006")
		preview := fmt.Sprintf("**📋 Чек-лист смены**\n*Дата последнего обновления: %s*\n\n%s", now, text)

		msg := tgbotapi.NewMessage(chatID, preview)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = checklistEditKeyboard()
		bot.Send(msg)

		userState[userID] = "checklist_preview"
		return
	}
	if userState[userID] == "waiting_intent_update" {
		_, p2pText, _ := getLimits()
		err := saveLimits(text, p2pText)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка сохранения"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**✅ Intent лимиты обновлены:**\n\n`%s`", text))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		userState[userID] = ""

	} else if userState[userID] == "waiting_p2p_update" {
		intentText, _, _ := getLimits()
		err := saveLimits(intentText, text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка сохранения"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**✅ P2P лимиты обновлены:**\n\n`%s`", text))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		userState[userID] = ""
	}
	if _, hasBalance := balanceStep[userID]; hasBalance {
		step := balanceStep[userID]
		providers := []string{"Бади", "Ген", "Врр", "Рх", "Глобал", "Глидех"}

		// Пропускаем пустые сообщения/кнопки
		if text == "" || text == "Посчитать балансы" {
			return
		}

		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Только числа!"))
			return
		}

		provider := providers[step%6]
		if step < 6 {
			balanceData[userID][fmt.Sprintf("limit_%s", provider)] = value
		} else {
			balanceData[userID][fmt.Sprintf("poured_%s", provider)] = value
		}

		balanceStep[userID]++
		nextStep := balanceStep[userID]

		if nextStep == 6 {
			bot.Send(tgbotapi.NewMessage(chatID, "✅ Лимиты сохранены!\n\nСколько вылил:\nБади:"))
		} else if nextStep < 12 {
			nextProvider := providers[nextStep%6]
			nextType := "вылил"
			if nextStep < 6 {
				nextType = "лимит"
			}
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("%s (%s):", nextProvider, nextType)))
		} else {
			totalLimit := 0.0
			totalPoured := 0.0
			for _, p := range providers {
				totalLimit += balanceData[userID][fmt.Sprintf("limit_%s", p)]
				totalPoured += balanceData[userID][fmt.Sprintf("poured_%s", p)]
			}

			remain := totalLimit - totalPoured
			w1 := remain * 0.71
			ls := remain * 0.16
			topx := remain * 0.13

			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Добрый день! Коллеги, у вас осталось %.0f INR трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.", w1)))
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Добрый день! Коллеги, у вас осталось %.0f INR трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.", ls)))
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Добрый день! Коллеги, у вас осталось %.0f INR трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.", topx)))
			delete(balanceStep, userID)
			delete(balanceData, userID)
		}
		return
	}

}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	data := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID

	bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "OK"))

	switch data {

	case "send":
		text := tempMessage[userID]
		err := saveNotification(text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка БД"))
			return
		}
		finalText := "🔥 Новое уведомление от лида!\n\n" + text
		broadcastMessage(bot, finalText)

		bot.Send(tgbotapi.NewMessage(chatID, "✅ Успешно отправлено!"))
		userState[userID] = ""
		tempMessage[userID] = ""

	case "rewrite":

		userState[userID] = "waiting_notification"

		bot.Send(tgbotapi.NewMessage(chatID, "Введи сообщение заново"))
	case "edit_checklist_menu":
		userState[userID] = "waiting_checklist"
		bot.Send(tgbotapi.NewMessage(chatID, "📝 Введи актуальную информацию"))

	case "save_checklist":
		text := tempMessage[userID]
		if text == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Текст потерян"))
			return
		}
		err := updateChecklist(text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка сохранения"))
			return
		}
		now := time.Now().Format("02.01.2006")
		broadcastText := fmt.Sprintf("**📋 Чек-лист смены**\n*Дата последнего обновления: %s*\n\n%s", now, text)
		broadcastMessage(bot, broadcastText)

		bot.Send(tgbotapi.NewMessage(chatID, "✅ Чек-лист успешно обновлен!"))
		userState[userID] = ""
		tempMessage[userID] = ""

	case "rewrite_checklist":
		userState[userID] = "waiting_checklist"
		bot.Send(tgbotapi.NewMessage(chatID, "📝 Введи информацию заново"))

	case "main_menu":
		sendMainMenu(bot, chatID, userID)
		userState[userID] = ""

	case "show_intent":
		intent, _, _ := getLimits()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**Intent лимиты:**\n\n%s", intent))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = limitsKeyboard()
		bot.Send(msg)

	case "show_p2p":
		_, p2p, _ := getLimits()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**P2P лимиты:**\n\n%s", p2p))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = limitsKeyboard()
		bot.Send(msg)

	case "edit_limits":
		if !isAdmin(userID) {
			return
		}
		msg := tgbotapi.NewMessage(chatID, "✏️ **Обновить лимиты**\n\nВыберите тип для редактирования:")
		msg.ReplyMarkup = limitsEditKeyboard()
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	case "copy_intent":
		intent, _, _ := getLimits()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**📋 Текущие Intent лимиты (скопируй и отредактируй):**\n\n`%s`", intent))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		userState[userID] = "waiting_intent_update"

	case "copy_p2p":
		_, p2p, _ := getLimits()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**📋 Текущие P2P лимиты (скопируй и отредактируй):**\n\n`%s`", p2p))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		userState[userID] = "waiting_p2p_update"

	case "scripts_intent":
		msg := tgbotapi.NewMessage(chatID, "💳 **Intent скрипты:**")
		msg.ReplyMarkup = intentScriptsKeyboard()
		bot.Send(msg)

	case "scripts_p2p":
		sendP2PScript(bot, chatID)

	case "scripts_back":
		msg := tgbotapi.NewMessage(chatID, "📜 **Выберите тип скриптов:**")
		msg.ReplyMarkup = scriptsKeyboard()
		bot.Send(msg)

	case "script_1win":
		send1winScript(bot, chatID)

	case "script_4ra":
		send4raScript(bot, chatID)

	}
}

func isAdmin(userID int64) bool {
	return userID == adminID
}

func sendChecklist(bot *tgbotapi.BotAPI, chatID int64) {

	text := "📋 Чек-лист смены:\n\n" + getChecklist()

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	var msg tgbotapi.MessageConfig

	if isAdmin(userID) {
		msg = tgbotapi.NewMessage(chatID, "Админ меню")
	} else {
		msg = tgbotapi.NewMessage(chatID, "Меню пользователя")
	}

	if isAdmin(userID) {
		msg.ReplyMarkup = getAdminKeyboard()
	} else {
		msg.ReplyMarkup = getMainKeyboard()
	}
	bot.Send(msg)
}
func sendTodayNotifications(bot *tgbotapi.BotAPI, chatID int64) {
	notifications, err := getTodayNotifications()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки"))
		return
	}

	if len(notifications) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "📢 Уведомлений за сегодня нет"))
		return
	}

	msgText := "📢 Уведомления за сегодня (" + strconv.Itoa(len(notifications)) + "):\n\n"
	for i, notif := range notifications {
		msgText += fmt.Sprintf("%d. %s\n\n", i+1, notif)
	}

	bot.Send(tgbotapi.NewMessage(chatID, msgText))
}

func broadcastMessage(bot *tgbotapi.BotAPI, text string) {

	users, err := getAllUsers()
	fmt.Println("USERS:", users)
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}

	for _, userID := range users {

		msg := tgbotapi.NewMessage(userID, text)

		_, err := bot.Send(msg)
		if err != nil {
			fmt.Println("Failed to send to:", userID, err)
		}
	}
}
func getChecklist() string {
	var text string
	err := db.QueryRow(`SELECT text FROM checklist WHERE id = 1`).Scan(&text)
	if err != nil || text == "" {
		return "📋 Чек-лист пока пуст"
	}
	return text
}
func getLimits() (string, string, error) {
	var intent, p2p string
	err := db.QueryRow(`
		SELECT intent, p2p FROM limits WHERE id = 1
	`).Scan(&intent, &p2p)

	if err != nil {
		intent = `Бади: 300-7000
Ген: 300-7000
Врр: 300-10000
Рх: 300-7000
Глобал: 300-10000
Глидех: 300-5000
Бета: 300-10000`
		p2p = `Бетих: 300-20000
Силк: 300-20000
Ф2: 300-20000
Бест: 300-20000
15: 500-25000`
		return intent, p2p, nil
	}
	return intent, p2p, nil
}
func saveLimits(intent, p2p string) error {
	_, err := db.Exec(`
		INSERT INTO limits (id, intent, p2p) VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET 
			intent = excluded.intent, 
			p2p = excluded.p2p
	`, intent, p2p)
	return err
}
func send1winScript(bot *tgbotapi.BotAPI, chatID int64) {
	intent, _, _ := getLimits()

	bot.Send(tgbotapi.NewMessage(chatID, "Deposit intent channel TD and H2H are submitted normally. (Min 300 - Max 10 000)"))
	bot.Send(tgbotapi.NewMessage(chatID, "Withdrawal channel is submitted normally. (Min 1000 - Max 150 000)"))
	bot.Send(tgbotapi.NewMessage(chatID, "Deposit intent channel TD and H2HTD are under maintenance, wait for next notifications."))
	bot.Send(tgbotapi.NewMessage(chatID, "Withdrawal channel is under maintenance, wait for next notifications."))
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**Intent лимиты:**\n\n%s", intent))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func send4raScript(bot *tgbotapi.BotAPI, chatID int64) {
	intent, _, _ := getLimits()

	bot.Send(tgbotapi.NewMessage(chatID, "Deposit channel TD submitted normally. (Min 400 - Max 10 000)"))
	bot.Send(tgbotapi.NewMessage(chatID, "Withdrawal channel is submitted normally. (Min 500 - Max 200 000)"))
	bot.Send(tgbotapi.NewMessage(chatID, "Deposit channel TD is under maintenance, wait for next notifications."))
	bot.Send(tgbotapi.NewMessage(chatID, "Withdrawal channel is under maintenance, wait for next notifications."))

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**Intent лимиты:**\n\n%s", intent))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func sendP2PScript(bot *tgbotapi.BotAPI, chatID int64) {
	_, p2p, _ := getLimits()

	bot.Send(tgbotapi.NewMessage(chatID, "Deposit channel P2P and P2PH2H are submitted normally. (Min 300 - Max 20 000)"))
	bot.Send(tgbotapi.NewMessage(chatID, "Deposit channel P2P and P2PH2H are under maintenance, wait for next notifications."))

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("**P2P лимиты:**\n\n%s", p2p))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
func calculateAndSend(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	providers := []string{"Бади", "Ген", "Врр", "Рх", "Глобал", "Глидех"}

	totalLimit, totalPoured := 0.0, 0.0
	data := balanceData[userID]

	for _, p := range providers {
		totalLimit += data[fmt.Sprintf("limit_%s", p)]
		totalPoured += data[fmt.Sprintf("poured_%s", p)]
	}

	remain := totalLimit - totalPoured
	w1 := remain * 0.76
	ls := remain * 0.13
	topx := remain * 0.11

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("`Добрый день! Коллеги, у вас осталось %.0f 1W трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.`", w1)))
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("`Добрый день! Коллеги, у вас осталось %.0f LS INR трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.`", ls)))
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("`Добрый день! Коллеги, у вас осталось %.0f TOPX трафика на депозитных TD и H2HTD интент каналах. Пожалуйста, используйте его.`", topx)))

	delete(balanceStep, userID)
	delete(balanceData, userID)
}
func startScheduler(bot *tgbotapi.BotAPI) {
	log.Println("🚀 Scheduler запущен")
	rand.Seed(time.Now().UnixNano())

	tickerPush := time.NewTicker(8 * time.Hour)
	go func() {
		for range tickerPush.C {
			pushText := `Не забудь пропушить провайдеров!

Colleagues, deposit channels TD and H2HTD are working smoothly. Please increase traffic.
Коллеги, депозитные каналы TD и H2HTD работают штатно. Пожалуйста, увеличьте трафик.`
			broadcastMessage(bot, pushText)
			log.Println("✅ Напоминание отправлено")
		}
	}()

	tickerCompliment := time.NewTicker(12 * time.Hour)
	go func() {
		compliments := []string{
			"🌟 Вы сегодня супер команда!",
			"💪 Продолжайте в том же духе!",
			"⭐ Отличная работа!",
			"🔥 Вы — лучшие!",
		}
		for range tickerCompliment.C {
			msg := compliments[rand.Intn(len(compliments))]
			broadcastMessage(bot, msg)
			log.Println("✅ Комплимент отправлен")
		}
	}()

	select {}
}
