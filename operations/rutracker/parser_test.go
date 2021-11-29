package operations

import (
	"fmt"
	"os"
	"testing"
)

func TestSearchArg(t *testing.T) {
	expected := "https://rutracker.org/forum/tracker.php?nm=fallout"
	actual := searchCall("fallout")
	if expected != actual {
		t.Fatalf("expected:%s, actual:%s", expected, actual)
	}
}

func TestEmptySearchResultList(t *testing.T) {
	reader, err := os.Open("test_data/not_found.html")
	if err != nil {
		t.Error(err)
	}

	items, err := parseItemListPage(reader)
	if len(items) != 0 {
		t.Fatalf("expected 0, actual %d", len(items))
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseList(t *testing.T) {
	reader, err := os.Open("test_data/item_list.html")
	if err != nil {
		t.Error(err)
	}

	items, _ := parseItemListPage(reader)
	if len(items) != 50 {
		t.Fatalf("expected 50, actual %d", len(items))
	}

	err = testItem(items[0], TorrentItem{Title: "Fallout 3. Game of The Year Edition [P] [RUS + ENG / RUS + ENG] (2009) (1.7, build 7447090 + 5 DLC)",
		Size: "7.3 GB ↓", TopicId: "6131488"})

	if err != nil {
		t.Error(err)
	}

	err = testItem(items[11], TorrentItem{Title: "Fallout 3 [L] [ENG / ENG] (2008) (1.0.0.12 / 1.7.0.3)",
		Size: "8.42 GB ↓", TopicId: "6097187"})

	if err != nil {
		t.Error(err)
	}
}

func TestRussianChars(t *testing.T) {
	reader, err := os.Open("test_data/item_list.html")
	if err != nil {
		t.Error(err)
	}

	items, _ := parseItemListPage(reader)

	err = testItem(items[2], TorrentItem{Title: "[Mod] Сборка модификаций F4NH для игры [Fallout 4 / 1.10.163.0.1 + DLC] [RUS]",
		Size: "26.36 GB ↓", TopicId: "6091039"})

	if err != nil {
		t.Error(err)
	}
}

func testItem(actualItem TorrentItem, expectedItem TorrentItem) error {
	if actualItem.Title != expectedItem.Title {
		return fmt.Errorf("expected %s, actual %s", expectedItem.Title, actualItem.Title)
	}

	if actualItem.Size != expectedItem.Size {
		return fmt.Errorf("expected %s, actual %s", expectedItem.Size, actualItem.Size)
	}

	if actualItem.TopicId != expectedItem.TopicId {
		return fmt.Errorf("expected %s, actual %s", expectedItem.TopicId, actualItem.TopicId)
	}

	return nil
}

// func TestDownloadFile(t *testing.T) {
// 	err := DownloadTorrentFile("out.torrent", TorrentItem{TopicId: "6131488"})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestGetList(t *testing.T) {
// 	// DownloadFile("res.html", "https://rutracker.org/forum/tracker.php?nm=fallout")
// 	items, err := SearchItems("fallout")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	for i := 0; i < len(items); i++ {
// 		t.Log(items[i].Title)
// 	}
// }
