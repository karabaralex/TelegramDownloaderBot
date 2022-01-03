package main

import (
	"fmt"
	"testing"

	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func TestConvertItemsToText(t *testing.T) {
	items := []rutracker.TorrentItem{
		{
			Title:   "title1",
			Size:    "3.3 GB",
			Seeds:   "1333",
			TopicId: "123",
		},
		{
			Title:   "title2",
			Size:    "1.3 GB",
			Seeds:   "222",
			TopicId: "321",
		},
	}

	first := fmt.Sprintf("%s\nSize:%s,Seeds:%s\n/%s	/details%s\n\n", items[0].Title, items[0].Size, items[0].Seeds, items[0].TopicId, items[0].TopicId)
	second := fmt.Sprintf("%s\nSize:%s,Seeds:%s\n/%s	/details%s\n\n", items[1].Title, items[1].Size, items[1].Seeds, items[1].TopicId, items[1].TopicId)

	text := convertItemsToText(items)
	expected := first + second
	if text != expected {
		t.Errorf("expected:\n%s, actual:\n%s", expected, text)
	}
}
