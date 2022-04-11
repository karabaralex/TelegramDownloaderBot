package main

import (
	"fmt"
	"runtime/debug"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

// safely call function without panic
func safeCall(f func(), reply func(string)) {
	defer func() {
		if r := recover(); r != nil {
			// print recovered stack trace
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			fmt.Println("Recovered in f", r)
			stackTraceUrl, sterr := operations.SendStringToPastebin(string(debug.Stack()))
			if sterr != nil {
				fmt.Println("Send stacktrace error ", sterr)
			} else {
				result := fmt.Sprintf("Error: %s, Stacktrace url: %s", r, stackTraceUrl)
				reply(result)
			}
		}
	}()
	f()
}

func main() {
	version := "Telegram downloader version 7"
	fmt.Println(version)
	envConfig, envError := config.Read()
	if envError != nil {
		fmt.Println("Load config error ", envError)
		return
	}

	rutracker.USER_NAME = envConfig.RuTrackerUserName
	rutracker.USER_PASSWORD = envConfig.RuTrackerPassword
	bot.API_TOKEN = envConfig.TelegramBotToken

	outputChannel := make(chan bot.OutMessage)

	bot.AddHandler(bot.NewCommandMatcher("/version"), func(message *bot.Info) {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: version}
	})

	bot.AddHandler(bot.NewCommandMatcher("/[0-9]+"), func(message *bot.Info) {
		topicId := message.Text[1:]
		fmt.Println("Command /[0-9]+", topicId)
		bot.SendTypingStatus(message)
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, topicId+".torrent")
		go safeCall(func() {
			operations.DownloadTorrentByPostId(topicId, destinationPath, func(result operations.OperationResult) {
				if result.Err != nil {
					fmt.Println(result.Text)
					reply := bot.OutMessage{OriginalMessage: message, Text: result.Text}
					outputChannel <- reply
				} else {
					fmt.Println("saved torrent file to ", destinationPath)
					reply := bot.OutMessage{OriginalMessage: message, Text: result.Text}
					outputChannel <- reply
				}
			})
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewTextMatcher(".*"), func(message *bot.Info) {
		fmt.Println("Command .*", message.Text)
		bot.SendTypingStatus(message)
		go safeCall(func() {
			operations.SearchTorrent(message.Text, func(result operations.OperationResult) {
				if result.Err != nil {
					fmt.Println(result.Text)
					reply := bot.OutMessage{OriginalMessage: message, Text: result.Text}
					outputChannel <- reply
				} else {
					fmt.Println("search result")
					// check if items nil or empty
					if result.Items == nil || len(result.Items) == 0 {
						reply := bot.OutMessage{OriginalMessage: message, Text: fmt.Sprintf("No results, watch /watch%s", message.Text)}
						outputChannel <- reply
					} else {
						textBlocks := convertItemsToText(result.Items)
						for _, textBlock := range textBlocks {
							reply := bot.OutMessage{OriginalMessage: message, Text: textBlock, Html: true}
							outputChannel <- reply
						}
					}
				}
			})
		}, func(s string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: s}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewFileNameMatcher(), func(message *bot.Info) {
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, message.FileName)
		bot.SendTypingStatus(message)
		go safeCall(func() {
			operations.Download(message.FileUrl, destinationPath, func(result operations.OperationResult) {
				if result.Err != nil {
					fmt.Println(result.Text)
					reply := bot.OutMessage{OriginalMessage: message, Text: result.Text}
					outputChannel <- reply
				} else {
					fmt.Println("saved torrent file to ", destinationPath)
					reply := bot.OutMessage{OriginalMessage: message, Text: result.Text}
					outputChannel <- reply
				}
			})
		}, func(s string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: s}
			outputChannel <- reply
		})
	})

	go bot.Sender(outputChannel)
	bot.RequestUpdates()
}
