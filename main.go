package main

import (
	"context"
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/config"
	"github.com/telegram-command-reader/operations"
	"github.com/telegram-command-reader/operations/ai"
	"github.com/telegram-command-reader/operations/jackett"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
	"github.com/telegram-command-reader/operations/storage"
	transmission "github.com/telegram-command-reader/operations/transmission"
)

var (
	jacketClient              *jackett.Jackett       // this is lib to search all torrent providers
	lastJackettRequestResults map[int]jackett.Result = make(map[int]jackett.Result)
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
	version := "Telegram downloader version 16"
	fmt.Println(version)
	envConfig, envError := config.Read()
	if envError != nil {
		fmt.Println("Load config error ", envError)
		return
	}

	jackett.JACKET_KEY = envConfig.JackettApiKey
	jackett.JACKET_PORT_FROM = envConfig.JackettPortFrom
	jackett.JACKET_PORT_TO = envConfig.JackettPortTo
	jackett.JACKET_URI = envConfig.JackettApiURL
	activeFolder := transmission.New(envConfig.ActiveTorrentFilesPath)
	finishedFolder := transmission.New(envConfig.FinishedFolder)

	rutracker.USER_NAME = envConfig.RuTrackerUserName
	rutracker.USER_PASSWORD = envConfig.RuTrackerPassword
	bot.API_TOKEN = envConfig.TelegramBotToken
	storage.API_KEY = envConfig.KVDBToken
	ai.API_KEY = envConfig.GeminiApiKey
	transmission.RPC_URI = envConfig.TransmissionUri
	transmission.RPC_PORT_FROM = envConfig.TransmissionPortFrom
	transmission.RPC_PORT_TO = envConfig.TransmissionPortTo

	outputChannel := make(chan bot.OutMessage)

	bot.AddHandler(bot.NewCommandMatcher("/search_([A-Za-z0-9+/]+={0,2})"), func(message *bot.Info) {
		re := regexp.MustCompile("^/search_([A-Za-z0-9+/]+={0,2})$")
		match1 := re.FindStringSubmatch(message.Text)
		if len(match1) > 0 {
			movie_name := bot.DecodeStringFromCommand(match1[1])
			searchTorrent(message, movie_name, outputChannel)
		}
	})

	bot.AddHandler(bot.NewCommandMatcher("/download_([0-9]+)"), func(message *bot.Info) {
		idStr := message.Text[10:]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			reply := bot.OutMessage{OriginalMessage: message, Text: "Invalid torrent ID: " + err.Error()}
			outputChannel <- reply
			return
		}

		go safeCall(func() {
			magnetUri := lastJackettRequestResults[id].MagnetUri
			linkUri := lastJackettRequestResults[id].Link
			if magnetUri == "" && linkUri == "" {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Try search again, no magnet URI found for ID: " + idStr}
				outputChannel <- reply
				return
			}

			if magnetUri != "" {
				result, err := transmission.AddTorrent(magnetUri)
				if err != nil {
					reply := bot.OutMessage{OriginalMessage: message, Text: "Error adding torrent: " + err.Error()}
					outputChannel <- reply
					return
				}
				if result {
					outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Downloading from magnet"}
					return
				}
			}

			if linkUri != "" {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Что делаем?", UseInlineKeyboard: true, InlineKeyboard: bot.DownloadActionKeyboard, ReplyCallback: func(data string) {
					fileTitle := strings.Map(func(r rune) rune {
						if unicode.IsLetter(r) || unicode.IsNumber(r) {
							return r
						}
						return '_'
					}, lastJackettRequestResults[id].Title)
					if data == bot.DownloadActionFile {
						operations.DownloadJackettTorrentByUriToStream(linkUri, func(result operations.OperationResult) {
							if result.Err != nil {
								fmt.Println(result.Text)
								reply := bot.OutMessage{OriginalMessage: message, Text: result.Err.Error()}
								outputChannel <- reply
							} else {
								fmt.Println("saved torrent file to stream")
								reply := bot.OutMessage{OriginalMessage: message, Text: fileTitle + ".torrent", FileStream: result.FileStream}
								outputChannel <- reply
							}
						})
					} else if data == bot.DownloadActionServer {
						destinationPath := config.CreateFilePath(envConfig.TorrentFileFolder, fileTitle+".torrent")
						operations.DownloadJackettTorrentByUri(linkUri, destinationPath, func(result operations.OperationResult) {
							if result.Err != nil {
								fmt.Println(result.Text)
							} else {
								fmt.Println("saved torrent file to ", destinationPath)
								go monitorTorrentUpdates(activeFolder, message, outputChannel, finishedFolder)
							}
						})
					}
				}}
				outputChannel <- reply
			}
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/delete_([A-Za-z0-9+/]+={0,2})"), func(message *bot.Info) {
		re := regexp.MustCompile("^/delete_([A-Za-z0-9+/]+={0,2})$")
		match1 := re.FindStringSubmatch(message.Text)
		if len(match1) > 0 {
			id, err := strconv.ParseInt(match1[1], 10, 64)
			if err != nil {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Invalid torrent ID"}
				outputChannel <- reply
				return
			}
			ok, err := transmission.RemoveTorrent(id)
			if err != nil {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Error deleting torrent: " + err.Error()}
				outputChannel <- reply
				return
			}
			if ok {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Deleted:" + bot.DecodeStringFromCommand(match1[1])}
				outputChannel <- reply
			} else {
				reply := bot.OutMessage{OriginalMessage: message, Text: "Failed to delete torrent"}
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
			showTorrentList(message, outputChannel)
		}, func(result string) {
			reply := bot.OutMessage{OriginalMessage: message, Text: result}
			outputChannel <- reply
		})
	})

	bot.AddHandler(bot.NewCommandMatcher("/finished"), func(message *bot.Info) {
		// read all files and send them to output channel
		go safeCall(func() {
			showTorrentList(message, outputChannel)
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
						searchJackett(searchText, originalMessage, outputChannel)
						// save := bot.EncodeStringToCommand("save", searchText)
						// reply := bot.OutMessage{OriginalMessage: originalMessage, Text: fmt.Sprintf("No results, %s", save)}
						// outputChannel <- reply
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

func searchJackett(searchText string, originalMessage *bot.Info, outputChannel chan bot.OutMessage) {
	jacketClient, err := jackett.GetClient()
	if err != nil {
		fmt.Println("No jackett ", err)
	} else {
		fmt.Println("Jackett found")
	}

	lastJackettRequestResults = make(map[int]jackett.Result)
	ctx := context.Background()
	input := &jackett.FetchRequest{Query: searchText}
	response, err := jacketClient.Fetch(ctx, input)
	if err != nil {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: fmt.Sprintf("Error searching: %v", err)}
		outputChannel <- reply
		return
	}

	if len(response.Results) == 0 {
		reply := bot.OutMessage{OriginalMessage: originalMessage, Text: "No results found"}
		outputChannel <- reply
		return
	}

	var results []string
	id := 0
	for _, result := range response.Results {
		lastJackettRequestResults[id] = result
		results = append(results, fmt.Sprintf("Title: %s\nSize: %d\nSeeders: %d\nDownload: /download_%d",
			result.Title, result.Size, result.Seeders, id))
		id = id + 1
	}

	reply := bot.OutMessage{OriginalMessage: originalMessage, Text: strings.Join(results, "\n\n")}
	outputChannel <- reply
}

func showTorrentList(message *bot.Info, outputChannel chan bot.OutMessage) {
	ok, err := transmission.CheckRPCConnection()
	if err != nil {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Error connecting to transmission: " + err.Error()}
		return
	}
	if !ok {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: "Could not connect to transmission"}
		return
	}
	torrents, err := transmission.GetAllTorrentsAsString()
	if err != nil {
		outputChannel <- bot.OutMessage{OriginalMessage: message, Text: err.Error()}
		return
	}
	outputChannel <- bot.OutMessage{OriginalMessage: message, Text: torrents}
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
