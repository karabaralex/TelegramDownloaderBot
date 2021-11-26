package config

import (
	"os"
	"errors"
	"path/filepath"
)

type Config struct {
	TorrentFileFolder string
	TelegramBotToken string
}

func Read() (Config, error) {
	result:=Config {}
	result.TorrentFileFolder=os.Getenv("TORRENT_FOLDER")
	result.TelegramBotToken=os.Getenv("TELEGRAM_TOKEN")
	if result.TorrentFileFolder=="" {
		return result,errors.New("no torrent folder")
	}
	if result.TelegramBotToken=="" {
		return result,errors.New("no telegram token")
	}
	return result,nil
}

func CreateFilePath(torrentFolder string, fileName string) (string) {
	return filepath.Join(torrentFolder,fileName)
}
