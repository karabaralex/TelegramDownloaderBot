package main

import (
	"fmt"
	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/operations"
	"github.com/telegram-command-reader/config"
)

func main() {
	fmt.Println("Telegram downloader ver 1")
	envConfig,envError:=config.Read()
	if envError!=nil {
		fmt.Println("Load config error ", envError)
		return
	}

	bot.API_TOKEN = envConfig.TelegramBotToken
	idCounter:=0
	idToMessage := make(map[int]bot.BotMessage)
	inputChannel:=make(chan bot.BotMessage)
	outputChannel:=make(chan bot.OutMessage)
	operationChannel:=make(chan operations.OperationResult)
	go bot.Sender(outputChannel)
	go bot.RequestUpdates(inputChannel)
	for {
		select {
		case message:=<-inputChannel:
			fmt.Println(message.Text)
			idCounter++
			idToMessage[idCounter]=message
			if message.Command == "FILE" {
				destinationPath:=config.CreateFilePath(envConfig.TorrentFileFolder,message.FileName)
				go operations.Download(operationChannel, idCounter, message.FileUrl, destinationPath)
			}
		case result:=<-operationChannel:
			fmt.Println(result.Text)
			message:=idToMessage[result.Id]
			reply := bot.OutMessage {OriginalMessage:message, Text: result.Text}
			outputChannel <- reply
		}
	}
}
