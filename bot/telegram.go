package bot

import (
	"fmt"
	"regexp"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Info struct {
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

type Hanlder func(message *Info)

type Matcher interface {
	match(update *tgbotapi.Update) bool
}

type CommandMatcher struct {
	regex string
}

type TextMatcher struct {
	regex string
}
type FileNameMatcher struct {
}

// create new CommandMatcher
func NewCommandMatcher(regex string) *CommandMatcher {
	return &CommandMatcher{regex: "^" + regex + "$"}
}

func NewTextMatcher(regex string) *TextMatcher {
	return &TextMatcher{regex: "^" + regex + "$"}
}

func NewFileNameMatcher() *FileNameMatcher {
	return &FileNameMatcher{}
}

var pair = make(map[Matcher]Hanlder)

// add handler to list
func AddHandler(matcher Matcher, handler Hanlder) {
	pair[matcher] = handler
}

func findHandlerForUpdate(update *tgbotapi.Update) (Hanlder, bool) {
	// iterate through pairs of matcher-handler
	for matcher, handler := range pair {
		if matcher.match(update) {
			return handler, true
		}
	}

	return nil, false
}

func (matcher *CommandMatcher) match(update *tgbotapi.Update) bool {
	if len(update.Message.Entities) > 0 && update.Message.Entities[0].Type == "bot_command" {
		// check if matcher.regex match update.Message.Text
		return regexp.MustCompile(matcher.regex).MatchString(update.Message.Text)
	}

	return false
}

func (matcher *TextMatcher) match(update *tgbotapi.Update) bool {
	// check not entity with type bot_command
	if len(update.Message.Entities) > 0 && update.Message.Entities[0].Type == "bot_command" {
		return false
	}

	if update.Message.Document != nil {
		return false
	}

	return regexp.MustCompile(matcher.regex).MatchString(update.Message.Text)
}

func (matcher *FileNameMatcher) match(update *tgbotapi.Update) bool {
	return update.Message.Document != nil
}

func RequestUpdates() {
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

		// check if matchers match
		handler, ok := findHandlerForUpdate(&update)
		if !ok {
			fmt.Println("no handler for update")
			continue
		}

		var fileUrl string
		var fileName string

		if update.Message.Document != nil {
			var err error = nil
			fileUrl, err = bot.GetFileDirectURL(update.Message.Document.FileID)
			fileName = update.Message.Document.FileName
			if err != nil {
				fmt.Println("error get url ", err)
				continue
			}
		}

		handler(&Info{
			Text:     update.Message.Text,
			FileName: fileName,
			FileUrl:  fileUrl,
			source:   update.Message,
		})

		// if update.Message.Document != nil {
		// 	fmt.Println("document ", update.Message.Document.FileID)
		// 	fileUrl, error := bot.GetFileDirectURL(update.Message.Document.FileID)
		// 	if error != nil {
		// 		fmt.Println("error get url ", error)
		// 		continue
		// 	}

		// 	fmt.Println("doc url ", fileUrl)
		// 	received <- BotMessage{
		// 		Command:  "FILE",
		// 		Text:     fileUrl,
		// 		FileName: update.Message.Document.FileName,
		// 		FileUrl:  fileUrl,
		// 		source:   update.Message}
		// } else if len(update.Message.Entities) > 0 && update.Message.Entities[0].Type == "bot_command" {
		// 	// if text starts with /watch
		// 	if update.Message.Text == "/watch" {
		// 		fmt.Println("watch")
		// 		received <- BotMessage{
		// 			Command: "WATCH",
		// 			Text:    update.Message.Text,
		// 			source:  update.Message}
		// 	} else {
		// 		fmt.Println("search for ", update.Message.Text)
		// 		received <- BotMessage{
		// 			Command: "DOWNLOAD_BY_ID",
		// 			Text:    update.Message.Text[1:],
		// 			source:  update.Message}
		// 	}
		// } else {
		// 	fmt.Println("search for ", update.Message.Text)
		// 	received <- BotMessage{
		// 		Command: "SEARCH",
		// 		Text:    update.Message.Text,
		// 		source:  update.Message}
		// }
	}
}

type OutMessage struct {
	OriginalMessage *Info
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

		for i := 0; i < 3; i++ {
			_, err := bot.Send(msg)
			if err != nil {
				fmt.Println("error send message ", err)
				time.Sleep(time.Second * 3)
				continue
			}
			break
		}
	}
}
