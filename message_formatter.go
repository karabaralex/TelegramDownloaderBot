package main

import (
	"fmt"

	rutracker "github.com/telegram-command-reader/operations/rutracker"
)

func convertItemsToText(items []rutracker.TorrentItem) []string {
	MAX_RES := 15
	if MAX_RES > len(items) {
		MAX_RES = len(items)
	}

	result := []string{}
	lines := ""
	for i := 0; i < MAX_RES; i++ {
		url := fmt.Sprintf("https://t.me/iv?url=https://rutracker.org/forum/viewtopic.php?t=%s&rhash=4625e276e6dfbf", items[i].TopicId)
		nextItem := fmt.Sprintf("%s\n<b>Size:%s</b>,Seeds:%s,%s\n/%s			<a href=\"%s\">details</a>\n\n",
			items[i].Title,
			items[i].Size,
			items[i].Seeds,
			items[i].Category,
			items[i].TopicId,
			url)
		// if lenght of line + next item is more than 4096 symbols, then start new string
		if len(lines)+len(nextItem) > 4096 {
			result = append(result, lines)
			lines = ""
		}

		lines += nextItem
	}

	if len(lines) > 0 {
		result = append(result, lines)
	}

	return result
}
