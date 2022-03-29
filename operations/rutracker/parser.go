package operations

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
)

var authCookie []*http.Cookie
var USER_NAME string
var USER_PASSWORD string

func authorize() error {
	if authCookie != nil {
		log.Println("already authorized, skipping")
		return nil
	}

	if USER_NAME == "" || USER_PASSWORD == "" {
		return errors.New("no rutracker auth params")
	}

	form := url.Values{}
	form.Add("login_username", USER_NAME)
	form.Add("login_password", USER_PASSWORD)
	form.Add("login", "%C2%F5%EE%E4")

	res, err := http.PostForm("https://rutracker.org/forum/login.php", form)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	authCookie = res.Request.Response.Cookies()

	defer res.Body.Close()
	return nil
}

type createRequest func() (*http.Request, error)

func makeRequest(create createRequest) (*http.Response, error) {
	err := authorize()
	if err != nil {
		return nil, err
	}

	// Declare http client
	client := &http.Client{}

	req, err := create()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(authCookie); i++ {
		req.AddCookie(authCookie[i])
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	return res, nil
}

func SearchItems(what string) ([]TorrentItem, error) {
	res, err := makeRequest(func() (*http.Request, error) {
		return http.NewRequest("GET", searchCall(what), nil)
	})

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	items, err := parseItemListPage(res.Body)
	if err != nil {
		return nil, err
	}

	sortListOfTorrentsBySeeders(items)
	return items, nil
}

func sortListOfTorrentsBySeeders(items []TorrentItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if seedsToInt(items[i].Seeds) < seedsToInt(items[j].Seeds) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// convert string number to number
func seedsToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		// return max int
		return 2147483647
	}
	return i
}

func searchCall(what string) string {
	return fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s", what)
}

func downloadCall(topicId string) string {
	return fmt.Sprintf("https://rutracker.org/forum/dl.php?t=%s", topicId)
}

func DownloadTorrentFile(filepath string, topicId string) error {
	res, err := makeRequest(func() (*http.Request, error) {
		return http.NewRequest("GET", downloadCall(topicId), nil)
	})

	if err != nil {
		return err
	}

	defer res.Body.Close()
	fmt.Printf("saving file for %s to %s\n", topicId, filepath)
	saveToFile(res.Body, filepath)
	return err
}

type TorrentItem struct {
	Title   string
	Size    string
	Seeds   string
	TopicId string
}

func parseItemListPage(body io.Reader) ([]TorrentItem, error) {
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	selector := doc.Find("#logged-in-username")
	if len(selector.Nodes) == 0 {
		return nil, errors.New("not logged in")
	}

	var items []TorrentItem

	rows := doc.Find(".hl-tr")
	if len(rows.Nodes) == 0 {
		// return empty list
		return items, nil
	}

	// Find the items items
	rows.Each(func(i int, row *goquery.Selection) {
		titleTag := row.Find(".hl-tags")
		title := titleTag.Nodes[0].FirstChild.Data
		size := row.Find(".tr-dl").Nodes[0].FirstChild.Data
		topicId := ""
		for i := range titleTag.Nodes[0].Attr {
			if titleTag.Nodes[0].Attr[i].Key == "data-topic_id" {
				topicId = titleTag.Nodes[0].Attr[i].Val
				break
			}
		}

		if topicId == "" {
			log.Println("not found id")
			return
		}

		seeds := "new"
		if len(row.Find(".seedmed").Nodes) != 0 {
			seeds = row.Find(".seedmed").Nodes[0].FirstChild.Data
		}

		title = string(decodeWindows1251([]uint8(title)))
		item := TorrentItem{Title: title, Size: size, Seeds: seeds, TopicId: topicId}
		items = append(items, item)
	})

	return items, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func saveToFile(body io.ReadCloser, file string) {
	// Create the file
	out, _ := os.Create(file)
	defer out.Close()

	// Write the body to file
	io.Copy(out, body)
}

func decodeWindows1251(ba []uint8) []uint8 {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(ba)
	return out
}
