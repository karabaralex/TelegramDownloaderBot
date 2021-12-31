package operations

import (
	"fmt"
	"time"

	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

type OperationResult struct {
	Text  string
	Items []rutracker.TorrentItem
	Err   error
}

// define callback function
type Callback func(result OperationResult)

func DownloadTorrentByPostId(topicId string, destination string, callback Callback) {
	err := rutracker.DownloadTorrentFile(destination, topicId)
	if err != nil {
		fmt.Println("Error download ", err)
		callback(OperationResult{Text: "cannot download", Err: err})
		return
	}

	callback(OperationResult{Text: "scheduled"})
}

func WatchTorrent(what string, callback Callback) {
	fmt.Printf("Watching %s\n", what)
	callback(OperationResult{Text: fmt.Sprintf("watching %s", what)})
}

func SearchTorrent(what string, callback Callback) {
	items, err := rutracker.SearchItems(what)
	if err != nil {
		fmt.Println("Error searching ", err)
		callback(OperationResult{Text: "cannot search", Err: err})
		return
	}

	callback(OperationResult{Text: "this is what I found", Items: items})
}

func Download(url string, destination string, callback Callback) {
	start := time.Now()
	fmt.Println("Download...")
	err := DownloadFile(destination, url)
	elapsed := time.Now().Sub(start)
	if err != nil {
		fmt.Println("Error downloading ", err)
		callback(OperationResult{Text: "error", Err: err})
		return
	}
	fmt.Println("Download took ", elapsed)
	callback(OperationResult{Text: "scheduled"})
}
