package main

import (
	"fmt"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func main() {
	version := "Telegram downloader ver 5"
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

	// add bot handlers
	bot.AddHandler(bot.NewCommandMatcher("/details[0-9]+"), func(message *bot.Info) {
		// get topic id from message text
		topicId := message.Text[len("/details"):]
		instantView := fmt.Sprintf("https://t.me/iv?url=https://rutracker.org/forum/viewtopic.php?t=%s&rhash=4625e276e6dfbf", topicId)
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: instantView}
	})

	bot.AddHandler(bot.NewCommandMatcher("/version"), func(message *bot.Info) {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: version}
	})

	bot.AddHandler(bot.NewCommandMatcher("/[0-9]+"), func(message *bot.Info) {
		topicId := message.Text[1:]
		fmt.Println("Command /[0-9]+", topicId)
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, topicId+".torrent")
		go operations.DownloadTorrentByPostId(topicId, destinationPath, func(result operations.OperationResult) {
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
	})

	bot.AddHandler(bot.NewTextMatcher(".*"), func(message *bot.Info) {
		fmt.Println("Command .*", message.Text)
		go operations.SearchTorrent(message.Text, func(result operations.OperationResult) {
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
					text := convertItemsToText(result.Items)
					reply := bot.OutMessage{OriginalMessage: message, Text: text}
					outputChannel <- reply
				}
			}
		})
	})

	bot.AddHandler(bot.NewFileNameMatcher(), func(message *bot.Info) {
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, message.FileName)
		go operations.Download(message.FileUrl, destinationPath, func(result operations.OperationResult) {
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
	})

	go bot.Sender(outputChannel)
	bot.RequestUpdates()
}
