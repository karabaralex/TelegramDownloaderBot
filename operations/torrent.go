package operations

import (
	"fmt"
	"io"
	"time"

	"github.com/telegram-command-reader/bot"
	"github.com/telegram-command-reader/operations/jackett"
	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

type OperationResult struct {
	Text       string
	Items      []rutracker.TorrentItem
	Err        error
	FileStream io.ReadCloser
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

func DownloadTorrentByPostIdToStream(topicId string, callback Callback) {
	stream, err := rutracker.DownloadTorrentFileToStream(topicId)
	if err != nil {
		fmt.Println("Error download ", err)
		callback(OperationResult{Text: "cannot download", Err: err})
		return
	}

	callback(OperationResult{Text: "scheduled", FileStream: stream})
}

func DownloadJackettTorrentByUri(uri string, destination string, callback Callback) {
	err := jackett.DownloadTorrentFile(destination, uri)
	if err != nil {
		fmt.Println("Error download ", err)
		callback(OperationResult{Text: "cannot download", Err: err})
		return
	}

	callback(OperationResult{Text: "scheduled"})
}

func DownloadJackettTorrentByUriToStream(uri string, callback Callback) {
	stream, err := jackett.DownloadTorrentFileToStream(uri)
	if err != nil {
		fmt.Println("Error download ", err)
		callback(OperationResult{Text: "cannot download", Err: err})
		return
	}

	callback(OperationResult{Text: "scheduled", FileStream: stream})
}

func WatchTorrent(what string, callback Callback) {
	fmt.Printf("Watching %s\n", what)
	callback(OperationResult{Text: fmt.Sprintf("watching %s", what)})
}

func SearchTorrent(what string, where string, callback Callback) {
	var items []rutracker.TorrentItem
	var err error
	if where == bot.All {
		items, err = rutracker.SearchEverywhere(what)
	} else if where == bot.Audiobooks {
		items, err = rutracker.SearchAudioBooks(what)
	} else if where == bot.Movies {
		items, err = rutracker.SearchMovies(what)
	} else if where == bot.Series {
		items, err = rutracker.SearchSeries(what)
	} else if where == bot.TextBooks {
		items, err = rutracker.SearchBooks(what)
	} else {
		fmt.Println("Incorrect search destination:" + where)
		items, err = rutracker.SearchEverywhere(what)
	}

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
