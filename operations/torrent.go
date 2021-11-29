package operations

import (
	"fmt"
	"time"

	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

type OperationResult struct {
	Id    int
	Text  string
	Items []rutracker.TorrentItem
	Err   error
}

func DownloadTorrentByPostId(channel chan OperationResult, id int, topicId string, destination string) {
	err := rutracker.DownloadTorrentFile(destination, topicId)
	if err != nil {
		fmt.Println("Error download ", err)
		channel <- OperationResult{Id: id, Text: "cannot download", Err: err}
		return
	}

	channel <- OperationResult{Id: id, Text: "scheduled"}
}

func SearchTorrent(channel chan OperationResult, id int, what string) {
	items, err := rutracker.SearchItems(what)
	if err != nil {
		fmt.Println("Error searching ", err)
		channel <- OperationResult{Id: id, Text: "cannot search", Err: err}
		return
	}

	channel <- OperationResult{Id: id, Text: "this is what I found", Items: items}
}

func Download(channel chan OperationResult, id int, url string, destination string) {
	start := time.Now()
	fmt.Println("Download...")
	err := DownloadFile(destination, url)
	elapsed := time.Now().Sub(start)
	if err != nil {
		fmt.Println("Error downloading ", err)
		channel <- OperationResult{Id: id, Text: "error", Err: err}
		return
	}
	fmt.Println("Download took ", elapsed)
	channel <- OperationResult{Id: id, Text: "scheduled"}
}
