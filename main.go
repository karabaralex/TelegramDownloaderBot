package main

import (
	"fmt"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func main() {
	fmt.Println("Telegram downloader ver 3")
	envConfig, envError := config.Read()
	if envError != nil {
		fmt.Println("Load config error ", envError)
		return
	}

	rutracker.USER_NAME = envConfig.RuTrackerUserName
	rutracker.USER_PASSWORD = envConfig.RuTrackerPassword
	bot.API_TOKEN = envConfig.TelegramBotToken
	idCounter := 0
	idToMessage := make(map[int]bot.BotMessage)
	inputChannel := make(chan bot.BotMessage)
	outputChannel := make(chan bot.OutMessage)
	operationChannel := make(chan operations.OperationResult)
	go bot.Sender(outputChannel)
	go bot.RequestUpdates(inputChannel)
	for {
		select {
		case message := <-inputChannel:
			fmt.Println(message.Text)
			idCounter++
			idToMessage[idCounter] = message
			if message.Command == "FILE" {
				destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, message.FileName)
				go operations.Download(operationChannel, idCounter, message.FileUrl, destinationPath)
			}
			if message.Command == "SEARCH" {
				go operations.SearchTorrent(operationChannel, idCounter, message.Text)
			}
			if message.Command == "DOWNLOAD_BY_ID" {
				destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, message.Text+".torrent")
				go operations.DownloadTorrentByPostId(operationChannel, idCounter, message.Text, destinationPath)
			}
			if message.Command == "WATCH" {
				go operations.WatchTorrent(operationChannel, idCounter, message.Text)
			}
		case result := <-operationChannel:
			message := idToMessage[result.Id]
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
		}
	}
}

func convertItemsToText(items []rutracker.TorrentItem, outputChannel chan bot.OutMessage, message bot.BotMessage) {
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
