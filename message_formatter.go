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
		url := fmt.Sprintf("https://t.me/iv?url=https://rutracker.org/forum/viewtopic.php?t=%s&rhash=4625e276e6dfbf", items[i].TopicId)
		lines = lines + fmt.Sprintf("%s\n<b>Size:%s</b>,Seeds:%s\n/%s			<a href=\"%s\">details</a>\n\n",
			items[i].Title,
			items[i].Size,
			items[i].Seeds,
			items[i].TopicId,
			url)
	}

	return lines
}
