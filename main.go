package main

import (
	"fmt"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func main() {
	fmt.Println("Telegram downloader ver 4")
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
	bot.AddHandler(bot.NewCommandMatcher("/[0-9]+"), func(message *bot.Info) {
		fmt.Println("Command /[0-9]+", message.Text)
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, message.Text+".torrent")
		go operations.DownloadTorrentByPostId(message.Text, destinationPath, func(result operations.OperationResult) {
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
					convertItemsToText(result.Items, outputChannel, message)
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

func convertItemsToText(items []rutracker.TorrentItem, outputChannel chan bot.OutMessage, message *bot.Info) {
	MAX_RES := 15
	if MAX_RES > len(items) {
		MAX_RES = len(items)
	}

	lines := ""
	for i := 0; i < MAX_RES; i++ {
		lines = lines + fmt.Sprintf("%s\nSize:%s,Seeds:%s\n/%s\n\n",
			items[i].Title,
			items[i].Size,
			items[i].Seeds,
			items[i].TopicId)
	}

	reply := bot.OutMessage{OriginalMessage: message, Text: lines}
	outputChannel <- reply
}
