package main

import (
	"fmt"
	"regexp"
	"runtime/debug"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	"github.com/telegram-command-reader/operations/ai"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
	"github.com/telegram-command-reader/operations/storage"
	transmission "github.com/telegram-command-reader/operations/transmission"
)

// safely call function without panic
func safeCall(f func(), onError func(string)) {
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
				onError(result)
			}
		}
	}()
	f()
}

func main() {
	version := "Telegram downloader version 13"
	fmt.Println(version)
	envConfig, envError := config.Read()
	if envError != nil {
		fmt.Println("Load config error ", envError)
		return
	}

	activeFolder := transmission.New(envConfig.ActiveTorrentFilesPath)
	finishedFolder := transmission.New(envConfig.FinishedFolder)

	rutracker.USER_NAME = envConfig.RuTrackerUserName
	rutracker.USER_PASSWORD = envConfig.RuTrackerPassword
	bot.API_TOKEN = envConfig.TelegramBotToken
	storage.API_KEY = envConfig.KVDBToken
	ai.API_KEY = envConfig.GeminiApiKey

	outputChannel := make(chan bot.OutMessage)

	bot.AddHandler(bot.NewCommandMatcher("/search_([A-Za-z0-9+/]+={0,2})"), func(message *bot.Info) {
		re := regexp.MustCompile("^/search_([A-Za-z0-9+/]+={0,2})$")
		match1 := re.FindStringSubmatch(message.Text)
		if len(match1) > 0 {
			movie_name := bot.DecodeStringFromCommand(match1[1])
			searchTorrent(message, movie_name, outputChannel)
		}
	})

	bot.AddHandler(bot.NewCommandMatcher("/delete_([A-Za-z0-9+/]+={0,2})"), func(message *bot.Info) {
		re := regexp.MustCompile("^/delete_([A-Za-z0-9+/]+={0,2})$")
		match1 := re.FindStringSubmatch(message.Text)
		if len(match1) > 0 {
			if storage.DeleteKey(match1[1]) {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Deleted:" + bot.DecodeStringFromCommand(match1[1])}
				outputChannel <- reply
			}
		}
	})

	bot.AddHandler(bot.NewCommandMatcher("/save_([A-Za-z0-9+/]+={0,2})"), func(message *bot.Info) {
		// read all files and send them to output channel
		go safeCall(func() {
			re := regexp.MustCompile("^/save_([A-Za-z0-9+/]+={0,2})$")
			match1 := re.FindStringSubmatch(message.Text)
			if len(match1) > 0 {
				movie_name := bot.DecodeStringFromCommand(match1[1])
				if storage.SetKeyValue(bot.EncodeString(movie_name), movie_name) {
					reply := bot.OutMessage{OriginalMessage: message, Text: "Saved:" + movie_name}
					outputChannel <- reply
				}
			}
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/saved"), func(message *bot.Info) {
		// read all files and send them to output channel
		go safeCall(func() {
			list := storage.GetAllKeys()

			// convert list to string, if list is empty then send message "No files"
			if len(list) == 0 {
				outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Nothing saved"}
			} else {
				result := ""
				for _, item := range list {
					result += (bot.DecodeStringFromCommand(item) + "\n")
					result += ("/search_" + item + "\n")
					result += ("/delete_" + item + "\n\n")
				}

				reply := bot.OutMessage{OriginalMessage: message, Text: result}
				outputChannel <- reply
			}
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/downloading"), func(message *bot.Info) {
		// read all files and send them to output channel
		go safeCall(func() {
			list := activeFolder.ReadAllFilesInFolder()

			// convert list to string, if list is empty then send message "No files"
			if len(list) == 0 {
				outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Nothing downloading"}
			} else {
				result := ""
				for _, item := range list {
					result += item + "\n\n"
				}

				reply := bot.OutMessage{OriginalMessage: message, Text: result}
				outputChannel <- reply
			}
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/finished"), func(message *bot.Info) {
		// read all files and send them to output channel
		go safeCall(func() {
			list := finishedFolder.ReadAllFilesInFolder()

			// convert list to string, if list is empty then send message "No files"
			if len(list) == 0 {
				outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Nothing finished"}
			} else {
				result := ""
				for _, item := range list {
					result += item + "\n\n"
				}

				reply := bot.OutMessage{OriginalMessage: message, Text: result}
				outputChannel <- reply
			}
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/version"), func(message *bot.Info) {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: version}
	})

	bot.AddHandler(bot.NewCommandMatcher("/[0-9]+"), func(originalMessage *bot.Info) {
		topicId := originalMessage.Text[1:]
		fmt.Println("Command /[0-9]+", topicId)
		bot.SendTypingStatus(originalMessage)
		destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, topicId+".torrent")
		go safeCall(func() {
			reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "Что делаем?", UseInlineKeyboard: true, InlineKeyboard: bot.DownloadActionKeyboard, ReplyCallback: func(data string) {
				if data == bot.DownloadActionFile {
					operations.DownloadTorrentByPostIdToStream(topicId, func(result operations.OperationResult) {
						if result.Err != nil {
							fmt.Println(result.Text)
							reply := bot.OutMessage{OriginalMessage: originalMessage, Text: result.Err.Error()}
							outputChannel <- reply
						} else {
							fmt.Println("saved torrent file to stream")
							reply := bot.OutMessage{OriginalMessage: originalMessage, Text: topicId + ".torrent", FileStream: result.FileStream}
							outputChannel <- reply
						}
					})
				} else if data == bot.DownloadActionServer {
					operations.DownloadTorrentByPostId(topicId, destinationPath, func(result operations.OperationResult) {
						if result.Err != nil {
							fmt.Println(result.Text)
						} else {
							fmt.Println("saved torrent file to ", destinationPath)
							go monitorTorrentUpdates(activeFolder, originalMessage, outputChannel, finishedFolder)
						}
					})
				}
			}}
			outputChannel <- reply
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: originalMessage, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewTextMatcher(".*"), func(message *bot.Info) {
		searchTorrent(message, message.Text, outputChannel)
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

func monitorTorrentUpdates(activeFolder transmission.WatchedFolder, originalMessage *bot.Info, outputChannel chan bot.OutMessage, finishedFolder transmission.WatchedFolder) {
	newFileName, err := activeFolder.WaitForNewFileWithRetry(25)
	if err != nil {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "Waited for torrent to start loading, but error: " + err.Error()}
		outputChannel <- reply
	} else {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "Start loading: " + newFileName}
		outputChannel <- reply
	}

	newFileName, err = finishedFolder.WaitForNewFileWithRetry(60 * 60 * 24)
	if err != nil {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "Waited for torrent to finish, but error: " + err.Error()}
		outputChannel <- reply
	} else {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: fmt.Sprintf("%s finished", newFileName)}
		outputChannel <- reply
	}
}

func searchTorrent(originalMessage *bot.Info, searchText string, outputChannel chan bot.OutMessage) {
	fmt.Println("Command .*", searchText)
	bot.SendTypingStatus(originalMessage)
	go safeCall(func() {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "Где искать?", UseInlineKeyboard: true, InlineKeyboard: bot.CategoriesKeyboard, ReplyCallback: func(data string) {
			operations.SearchTorrent(searchText, data, func(result operations.OperationResult) {
				if result.Err != nil {
					fmt.Println(result.Text)
					reply := bot.OutMessage{OriginalMessage: originalMessage, Text: result.Text}
					outputChannel <- reply
				} else {
					fmt.Println("search result")
					// check if items nil or empty
					if result.Items == nil || len(result.Items) == 0 {
						save := bot.EncodeStringToCommand("save", searchText)
						reply := bot.OutMessage{OriginalMessage: originalMessage, Text: fmt.Sprintf("No results, %s", save)}
						outputChannel <- reply
					} else {
						textBlocks := convertItemsToText(result.Items)
						for _, textBlock := range textBlocks {
							reply := bot.OutMessage{OriginalMessage: originalMessage, Text: textBlock, Html: true}
							outputChannel <- reply
						}

						// go makeAiResponse(result, searchText, originalMessage, outputChannel)
					}
				}
			})
		}}
		outputChannel <- reply
	}, func(s string) {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: s}
		outputChannel <- reply
	})
}

func makeAiResponse(result operations.OperationResult, searchText string, originalMessage *bot.Info, outputChannel chan bot.OutMessage) {
	prompt := convertItemsToPrompt(result.Items, searchText)
	fmt.Println(prompt)
	ai_result, ai_error := ai.GenerateAiResponse(prompt)
	if ai_error != nil {
		fmt.Println(ai_error)
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: ai_error.Error()}
		outputChannel <- reply
	} else {
		fmt.Println(ai_result)
		aiOutput := fmt.Sprintf("<b>Совет искуственного интеллекта:</b>\n%s", ai_result)
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: aiOutput, Html: true}
		outputChannel <- reply
	}
}
