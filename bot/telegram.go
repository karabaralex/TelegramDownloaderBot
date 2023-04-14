package bot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Info struct {
	Text     string
	FileName string
	FileUrl  string
	source   *tgbotapi.Message
	callback *tgbotapi.CallbackQuery
}

var API_TOKEN string
var botInstance *tgbotapi.BotAPI

// create bot, print error if any, do not panic
// retry if error
func createBot() *tgbotapi.BotAPI {
	if botInstance != nil {
		return botInstance
	}

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
type InlineResponseMatcher struct {
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

func NewInlineResponseMatcher() *InlineResponseMatcher {
	return &InlineResponseMatcher{}
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

func SendTypingStatus(info *Info) {
	bot := createBot()

	msg := tgbotapi.NewChatAction(info.source.Chat.ID, tgbotapi.ChatTyping)
	bot.Send(msg)
}

// public enum of values for inline response
const (
	Movies     = "Movies"
	Series     = "Series"
	Audiobooks = "Audiobooks"
	All        = "All"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Везде", All),
		tgbotapi.NewInlineKeyboardButtonData("Фильмы", Movies),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Сериалы", Series),
		tgbotapi.NewInlineKeyboardButtonData("Аудиокниги", Audiobooks),
	),
)

func RequestUpdates() {
	bot := createBot()
	bot.Debug = false

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
		if update.Message != nil {
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
		} else if update.CallbackQuery != nil {
			// find if we have reply function for the message in hash and call it
			replyCallback, ok := replyCallbacks[update.CallbackQuery.Message.MessageID]
			if ok {
				replyCallback(update.CallbackQuery.Data)
			}

			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error())
				bot.Send(msg)
				continue
			}

			// originalText := update.CallbackQuery.Message.ReplyToMessage.Text
			// // And finally, send a message containing the data received.
			// msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, originalText)
			// if _, err := bot.Send(msg); err != nil {
			// 	fmt.Println(err.Error())
			// 	continue
			// }
		} else {
			fmt.Println("update without message")
			continue
		}
	}
}

type OutMessage struct {
	OriginalMessage *Info
	Text            string
	Html            bool
	InlineKeyboard  bool
	ReplyCallback   func(string)
}

// map of reply callbacks
var replyCallbacks = make(map[int]func(string))

func Sender(sendChannel chan OutMessage) {
	bot := createBot()

	for toSend := range sendChannel {
		telegramMessage := toSend.OriginalMessage.source

		// if toSend.text is more than 4096 symbols, trim it
		if len(toSend.Text) > 4096 {
			toSend.Text = toSend.Text[:4096]
		}

		// Now that we know we've gotten a new message, we can construct a
		// reply! We'll take the Chat ID and Text from the incoming message
		// and use it to create a new message.
		msg := tgbotapi.NewMessage(telegramMessage.Chat.ID, toSend.Text)
		if toSend.Html {
			msg.ParseMode = "HTML"
		}

		// We'll also say that this message is a reply to the previous message.
		// For any other specifications than Chat ID or Text, you'll need to
		// set fields on the `MessageConfig`.
		msg.ReplyToMessageID = telegramMessage.MessageID
		if toSend.InlineKeyboard {
			msg.ReplyMarkup = numericKeyboard
		}

		for i := 0; i < 3; i++ {
			sentMessage, err := bot.Send(msg)
			if err != nil {
				fmt.Println("error send message ", err)
				time.Sleep(time.Second * 3)
				continue
			}

			if toSend.ReplyCallback != nil {
				// we need to wait for user reply, add message to hashmap by id
				replyCallbacks[sentMessage.MessageID] = toSend.ReplyCallback
			}

			break
		}
	}
}

// encode string
func EncodeString(value string) string {
	// Replace spaces in encoded value with %&
	encodedValue := strings.ReplaceAll(value, " ", "1X2U1")
	return encodedValue
}

// encode string, to make it telegram command
func EncodeStringToCommand(command string, value string) string {
	// Replace spaces in encoded value with %&
	encodedValue := EncodeString(value)

	// Combine command and encoded value with underscore
	return fmt.Sprintf("/%s_%s", command, encodedValue)
}

func DecodeStringFromCommand(value string) string {
	// Replace spaces in encoded value with %&
	decodedValue := strings.ReplaceAll(value, "1X2U1", " ")

	// Combine command and encoded value with underscore
	return decodedValue
}
