package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	TorrentFileFolder      string // folder to store torrent files, which will be downloaded by transmission as a result
	TelegramBotToken       string
	RuTrackerUserName      string
	RuTrackerPassword      string
	ActiveTorrentFilesPath string // folder which currently downloading, transmission will move torrent files to this folder
	FinishedFolder         string // folder with downloaded content
	KVDBToken              string // api token for kvdb.io
	GeminiApiKey           string
	TransmissionUri        string // URI for connecting to transmission RPC server
	TransmissionPortFrom   int
	TransmissionPortTo     int
}

func Read() (Config, error) {
	result := Config{}
	result.TransmissionUri = os.Getenv("TRANSMISSION_URI")
	if result.TransmissionUri == "" {
		result.TransmissionUri = "127.0.0.1"
	}
	result.TransmissionPortFrom = parseIntOrDefault(os.Getenv("TRANSMISSION_PORT_FROM"), 0)
	result.TransmissionPortTo = parseIntOrDefault(os.Getenv("TRANSMISSION_PORT_TO"), 0)
	result.TorrentFileFolder = os.Getenv("TORRENT_FOLDER")
	result.TelegramBotToken = os.Getenv("TELEGRAM_TOKEN")
	if result.TorrentFileFolder == "" {
		return result, errors.New("no torrent folder")
	}
	if result.TelegramBotToken == "" {
		return result, errors.New("no telegram token")
	}

	result.RuTrackerUserName = os.Getenv("RUTRACKER_LOGIN")
	result.RuTrackerPassword = os.Getenv("RUTRACKER_PASSWORD")
	result.ActiveTorrentFilesPath = os.Getenv("ACTIVE_TORRENT_FILES_PATH")
	result.FinishedFolder = os.Getenv("FINISHED_FOLDER")
	result.KVDBToken = os.Getenv("KVDB_TOKEN")
	result.GeminiApiKey = os.Getenv("GEMINI_AI_API_TOKEN")
	if result.RuTrackerUserName == "" || result.RuTrackerPassword == "" || result.KVDBToken == "" {
		return result, errors.New("missing arguments")
	}
	return result, nil
}

func CreateFilePath(torrentFolder string, fileName string) string {
	return filepath.Join(torrentFolder, fileName)
}

func parseIntOrDefault(str string, defaultValue int) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return num
}
