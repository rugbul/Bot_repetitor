package main

import (
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botToken       = "7625504322:AAFj7rHYZRm9jbG8YgJStcXAevazQiiCChU" // Замените на токен вашего бота
	channel1       = "@OdnaZdizavtra"                                 // Укажите username канала (начинается с @)
	channel2       = "@doShkolnik22"                                  // Укажите username канала (начинается с @)
	correctAnswers = []int{2, 1, 3, 3, 1, 4}                          // Правильные ответы на вопросы
	userAnswers    = make(map[int64][]int)                            // Ответы пользователей

	// Ссылки на видео из VK Cloud
	welcomeVideoURL = "https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/%D0%A1%D0%B0%D1%88%D0%BA%D0%B0%20%D0%BF%D1%80%D0%BE%D1%81%D0%B8%D1%82%20%D0%BF%D0%BE%D0%BC%D0%BE%D1%87%D1%8C%20%D0%B1%D0%B5%D0%B7%20%D0%BF%D1%80%D0%B8%D0%B3%D0%BB%D0%B0%D1%88%D0%B5%D0%BD%D0%B8%D1%8F.mp4"
	videoURLs       = []string{
		"https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/%D1%83%D0%BF%D0%B0%D0%BB%20%D0%BA%D1%80%D0%BE%D0%B2%D0%BE%D1%82%D0%BE%D1%87.mp4",
		"https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/10%20%D0%9F%D0%BE%D0%B4%D0%B0%D0%B2%D0%B8%D0%BB%D1%81%D1%8F.mp4",
		"https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/10%20%D0%A1%D0%B0%D1%88%D0%BA%D0%B0%20%D0%B7%D0%B0%D0%BC%D0%B5%D1%80%D0%B7.mp4",
		"https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/10%20%D0%A1%D0%B0%D1%88%D0%BA%D0%B0%20%D0%BE%D0%B6%D0%BE%D0%B3.mp4",
		"https://my-video-bucket.hb.ru-msk.vkcloud-storage.ru/16%20%D0%97%D0%BC%D0%B5%D1%8F%20%D1%83%D0%BA%D1%83%D1%81%D0%B8%D0%BB%D0%B0.mp4",
	}

	// Вопросы
	questions = [][]string{
		{
			"Что делать при носовом кровотечении?\n1. Поднять голову, чтобы не пачкать одежду\n2. Зажать нос и наклонить голову вниз\n3. Громко кричать и звать на помощь\n4. Засунуть нос в холодильник",
			"А когда сильно ударился ногой?\n1. Покой и холод\n2. Ползти домой\n3. Приложить горячую грелку\n4. Наложить жгут",
		},
		{
			"Подавился и кашляет, что же ему делать?\n1. Бить по спине\n2. Давить на живот\n3. Ничего не делать, человек справится сам\n4. Продолжить есть",
		},
		{
			"Что делать если сильно замерз на улице, даже пальцы онемели?\n1. Растирать щеки снегом\n2. Греться у костра\n3. Руки греть подмышкой, на щеки натянуть шарф\n4. Залезть дома в горячую ванну",
		},
		{
			"Что делать при ожоге?\n1. Засунуть в холодную воду\n2. Засунуть в горячую воду\n3. Намазать мазью\n4. Полить маслом",
		},
		{
			"Что делать когда укусила змея?\n1. Заорать и убежать\n2. Укусить змею\n3. Наложить жгут и отсосать яд\n4. Запомнить змею, снять все тугие вещи (кольца, браслеты), не шевелить укушенной конечностью и обратиться к врачу",
		},
	}

	// Текущий шаг для каждого пользователя
	userStep = make(map[int64]int)
	// Текущий вопрос для каждого пользователя
	userQuestionStep = make(map[int64]int)
)

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("Ошибка при создании бота: ", err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if message.Text == "/start" {
		sendWelcomeVideo(bot, chatID) // Отправляем приветственное видео
	} else {
		// Обработка ответов на вопросы
		if len(userAnswers[chatID]) < len(correctAnswers) {
			// Проверяем, что ответ является цифрой от 1 до 4
			if isValidAnswer(message.Text) {
				processAnswer(bot, chatID, message.Text)
			} else {
				// Если ответ некорректный, отправляем сообщение
				msg := tgbotapi.NewMessage(chatID, "Неправильная команда. Попробуйте снова.")
				bot.Send(msg)
			}
		}
	}
}

// Проверка, что ответ является цифрой от 1 до 4
func isValidAnswer(answer string) bool {
	switch answer {
	case "1", "2", "3", "4":
		return true
	default:
		return false
	}
}

func sendWelcomeVideo(bot *tgbotapi.BotAPI, chatID int64) {
	// Отправляем приветственное видео
	video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(welcomeVideoURL))
	video.Caption = "Привет, я Сашка и мне нужны твои знания по курсу первой помощи!\nПодпишись на каналы, чтобы продолжить:"
	video.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Канал 1", "https://t.me/OdnaZdizavtra"),
			tgbotapi.NewInlineKeyboardButtonURL("Канал 2", "https://t.me/doShkolnik22"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Проверить подписку", "check_subscription"),
		),
	)
	_, err := bot.Send(video)
	if err != nil {
		log.Println("Ошибка при отправке приветственного видео: ", err)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID

	switch callbackQuery.Data {
	case "check_subscription":
		channel1ID, err := getChatIDByUsername(bot, channel1)
		if err != nil {
			log.Println("Ошибка при получении ID канала 1: ", err)
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при проверке подписки. Попробуйте позже.")
			bot.Send(msg)
			return
		}

		channel2ID, err := getChatIDByUsername(bot, channel2)
		if err != nil {
			log.Println("Ошибка при получении ID канала 2: ", err)
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при проверке подписки. Попробуйте позже.")
			bot.Send(msg)
			return
		}

		isSubscribed, err := checkSubscription(bot, chatID, channel1ID, channel2ID)
		if err != nil {
			log.Println("Ошибка при проверке подписки: ", err)
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при проверке подписки. Попробуйте позже.")
			bot.Send(msg)
			return
		}

		if !isSubscribed {
			msg := tgbotapi.NewMessage(chatID, "Вы не подписаны на все каналы. Пожалуйста, подпишитесь и нажмите 'Проверить подписку' снова.")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Проверить подписку", "check_subscription"),
				),
			)
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Отлично, тогда давай начнем! Выбирай правильный ответ и вписывай его цифрой.")
		_, err = bot.Send(msg)
		if err != nil {
			log.Println("Ошибка при отправке сообщения: ", err)
			return
		}

		time.Sleep(2 * time.Second)

		video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(videoURLs[0]))
		_, err = bot.Send(video)
		if err != nil {
			log.Println("Ошибка при отправке первого видео: ", err)
			return
		}

		userStep[chatID] = 0
		userQuestionStep[chatID] = 0
		sendNextQuestion(bot, chatID)
	}
}

func getChatIDByUsername(bot *tgbotapi.BotAPI, username string) (int64, error) {
	chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			SuperGroupUsername: username,
		},
	})
	if err != nil {
		return 0, err
	}
	return chat.ID, nil
}

// Проверка подписки на каналы
func checkSubscription(bot *tgbotapi.BotAPI, userID int64, channel1ID, channel2ID int64) (bool, error) {
	// Проверяем подписку на первый канал
	member1, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: channel1ID,
			UserID: userID,
		},
	})
	if err != nil {
		return false, err
	}

	// Проверяем подписку на второй канал
	member2, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: channel2ID,
			UserID: userID,
		},
	})
	if err != nil {
		return false, err
	}

	// Проверяем статус пользователя в обоих каналах
	if (member1.Status == "member" || member1.Status == "administrator" || member1.Status == "creator") &&
		(member2.Status == "member" || member2.Status == "administrator" || member2.Status == "creator") {
		return true, nil
	}

	return false, nil
}

func processAnswer(bot *tgbotapi.BotAPI, chatID int64, answer string) {
	// Записываем ответ пользователя
	userAnswers[chatID] = append(userAnswers[chatID], parseAnswer(answer))

	// Переходим к следующему вопросу
	userQuestionStep[chatID]++

	// Если вопросы для текущего видео закончились, переходим к следующему видео
	if userQuestionStep[chatID] >= len(questions[userStep[chatID]]) {
		userStep[chatID]++
		userQuestionStep[chatID] = 0
		sendNextVideo(bot, chatID)
	} else {
		// Иначе отправляем следующий вопрос
		sendNextQuestion(bot, chatID)
	}
}

func parseAnswer(answer string) int {
	switch answer {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	default:
		return 0
	}
}

func sendNextVideo(bot *tgbotapi.BotAPI, chatID int64) {
	step := userStep[chatID]

	// Если все видео пройдены, отправляем итоговое сообщение
	if step >= len(videoURLs) {
		sendFinalMessage(bot, chatID)
		return
	}

	// Отправляем видео
	video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(videoURLs[step]))
	_, err := bot.Send(video)
	if err != nil {
		log.Println("Ошибка при отправке видео: ", err)
		return
	}

	// Отправляем первый вопрос для этого видео
	sendNextQuestion(bot, chatID)
}

func sendNextQuestion(bot *tgbotapi.BotAPI, chatID int64) {
	step := userStep[chatID]
	questionStep := userQuestionStep[chatID]

	// Отправляем вопрос
	msg := tgbotapi.NewMessage(chatID, questions[step][questionStep])
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке вопроса: ", err)
	}
}

func sendFinalMessage(bot *tgbotapi.BotAPI, chatID int64) {
	correctCount := 0
	for i, answer := range userAnswers[chatID] {
		if answer == correctAnswers[i] {
			correctCount++
		}
	}

	// Сообщение с количеством правильных ответов
	resultMessage := tgbotapi.NewMessage(chatID, "Правильно "+strconv.Itoa(correctCount)+" из 6.")
	_, err := bot.Send(resultMessage)
	if err != nil {
		log.Println("Ошибка при отправке результата: ", err)
	}

	// Итоговое сообщение
	var finalMessage string
	if correctCount == len(correctAnswers) {
		finalMessage = "Ты просто профи! Приглашаю на курс: \"Первая помощь с Сашкой\" для детей 6-12 лет."
	} else if correctCount >= len(correctAnswers)/2 {
		finalMessage = "Молодец, ты много знаешь. Приглашаю на курс: \"Первая помощь с Сашкой\" для детей 6-12 лет."
	} else {
		finalMessage = "Не беда, главное начало. Приглашаю на курс: \"Первая помощь с Сашкой\" для детей 6-12 лет."
	}

	finalMessage += "\nПодробности и ссылка на меня: @JuliaGorodovikova"
	msg := tgbotapi.NewMessage(chatID, finalMessage)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке итогового сообщения: ", err)
	}
}
