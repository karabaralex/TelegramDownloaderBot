package main

import (
	"fmt"

	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func convertItemsToText(items []rutracker.TorrentItem) string {
	MAX_RES := 15
	if MAX_RES > len(items) {
		MAX_RES = len(items)
	}

	lines := ""
	for i := 0; i < MAX_RES; i++ {
		lines = lines + fmt.Sprintf("%s\nSize:%s,Seeds:%s\n/%s	/details%s\n\n",
			items[i].Title,
			items[i].Size,
			items[i].Seeds,
			items[i].TopicId,
			items[i].TopicId)
	}

	return lines
}
