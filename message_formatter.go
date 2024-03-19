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

func convertItemsToPrompt(items []rutracker.TorrentItem, searchQuery string) string {
	MAX_RES := 15
	if MAX_RES > len(items) {
		MAX_RES = len(items)
	}

	lines := ""
	for i := 0; i < MAX_RES; i++ {
		nextItem := fmt.Sprintf("%d) Title:%s,Size:%s,Seeds:%s,Category:%s, /%s\n",
			i,
			items[i].Title,
			items[i].Size,
			items[i].Seeds,
			items[i].Category,
			items[i].TopicId)
		lines += nextItem
	}

	prompt := fmt.Sprintf(`you are given a search result for a search query \"%s\", you need to pick best candidate based on following criteria (ordered by priority from high to low):
	0. it is a movie
	1. should not mention DVD
	2. the more seeds the better
	3. if this is single movie then reasonable size is about 10 gb
	4. if series then reasonable size is about 40 gb
	
	If there are no suitable candidates then you can skip some of requirement  above. Respond with exact item from the list and don't change anything in the item.
	%s`, searchQuery, lines)

	return prompt
}
