package config

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	os.Setenv("RUTRACKER_LOGIN", "a")
	os.Setenv("RUTRACKER_PASSWORD", "a")
	os.Setenv("KVDB_TOKEN", "a")
}

func TestFolderNotSet(t *testing.T) {
	os.Setenv("TORRENT_FOLDER", "")
	os.Setenv("TELEGRAM_TOKEN", "")
	_, error := Read()
	if error == nil {
		t.Fatalf("expected error")
	}
}

func TestFolderSet(t *testing.T) {
	os.Setenv("TORRENT_FOLDER", "my_folder")
	os.Setenv("TELEGRAM_TOKEN", "token")
	config, error := Read()
	if error != nil {
		t.Fatalf("not expected error")
	}
	if config.TorrentFileFolder != "my_folder" {
		t.Fatalf("not expected %v", config.TorrentFileFolder)
	}
	if config.TelegramBotToken != "token" {
		t.Fatalf("not expected %v", config.TelegramBotToken)
	}
}

func TestCreateFilePath(t *testing.T) {
	actual := CreateFilePath("path", "file")
	expected := "path/file"
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestCreateFilePath2(t *testing.T) {
	actual := CreateFilePath("path/", "file")
	expected := "path/file"
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestCreateFilePath3(t *testing.T) {
	actual := CreateFilePath("/a/b/path/", "file")
	expected := "/a/b/path/file"
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestCreateFilePath4(t *testing.T) {
	actual := CreateFilePath("/a/b/path/", "/file")
	expected := "/a/b/path/file"
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
