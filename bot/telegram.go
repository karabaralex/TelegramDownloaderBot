package bot

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotMessage struct {
	Command  string
	Text     string
	FileName string
	FileUrl  string
	source   *tgbotapi.Message
}

var API_TOKEN string

// create bot, print error if any, do not panic
// retry if error
func createBot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(API_TOKEN)
	if err != nil {
		fmt.Println("error create bot ", err)
		fmt.Println("retry in 15 seconds")
		time.Sleep(time.Second * 15)
		return createBot()
	}
	return bot
}

func RequestUpdates(received chan BotMessage) {
	bot := createBot()
	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			fmt.Println("update without message")
			continue
		}

		if update.Message.Document != nil {
			fmt.Println("document ", update.Message.Document.FileID)
			fileUrl, error := bot.GetFileDirectURL(update.Message.Document.FileID)
			if error != nil {
				fmt.Println("error get url ", error)
				continue
			}

			fmt.Println("doc url ", fileUrl)
			received <- BotMessage{
				Command:  "FILE",
				Text:     fileUrl,
				FileName: update.Message.Document.FileName,
				FileUrl:  fileUrl,
				source:   update.Message}
		} else if len(update.Message.Entities) > 0 && update.Message.Entities[0].Type == "bot_command" {
			// if text starts with /watch
			if update.Message.Text == "/watch" {
				fmt.Println("watch")
				received <- BotMessage{
					Command: "WATCH",
					Text:    update.Message.Text,
					source:  update.Message}
			} else {
				fmt.Println("search for ", update.Message.Text)
				received <- BotMessage{
					Command: "DOWNLOAD_BY_ID",
					Text:    update.Message.Text[1:],
					source:  update.Message}
			}
		} else {
			fmt.Println("search for ", update.Message.Text)
			received <- BotMessage{
				Command: "SEARCH",
				Text:    update.Message.Text,
				source:  update.Message}
		}
	}
}

type OutMessage struct {
	OriginalMessage BotMessage
	Text            string
}

func Sender(sendChannel chan OutMessage) {
	bot := createBot()

	for toSend := range sendChannel {
		telegramMessage := toSend.OriginalMessage.source

		// Now that we know we've gotten a new message, we can construct a
		// reply! We'll take the Chat ID and Text from the incoming message
		// and use it to create a new message.
		msg := tgbotapi.NewMessage(telegramMessage.Chat.ID, toSend.Text)
		// We'll also say that this message is a reply to the previous message.
		// For any other specifications than Chat ID or Text, you'll need to
		// set fields on the `MessageConfig`.
		msg.ReplyToMessageID = telegramMessage.MessageID

		// Okay, we're sending our message off! We don't care about the message
		// we just sent, so we'll discard it.
		if _, err := bot.Send(msg); err != nil {
			// Note that panics are a bad way to handle errors. Telegram can
			// have service outages or network errors, you should retry sending
			// messages or more gracefully handle failures.
			panic(err)
		}
	}
}
