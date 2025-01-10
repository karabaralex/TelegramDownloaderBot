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
	"strings"

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

	const (
		loginFormKey    = "login"
		loginFormValue  = "%C2%F5%EE%E4"
		usernameFormKey = "login_username"
		passwordFormKey = "login_password"
	)

	form := url.Values{}
	form.Add(usernameFormKey, USER_NAME)
	form.Add(passwordFormKey, USER_PASSWORD)
	form.Add(loginFormKey, loginFormValue)

	res, err := http.PostForm("https://rutracker.org/forum/login.php", form)
	if err != nil {
		log.Panic(err)
		return nil
	}

	if res.StatusCode != 200 {
		log.Panicf("status code error: %d %s", res.StatusCode, res.Status)
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
		log.Panic(err)
	}

	for i := 0; i < len(authCookie); i++ {
		req.AddCookie(authCookie[i])
	}

	res, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}

	if res.StatusCode != 200 {
		log.Panicf("status code error: %d %s", res.StatusCode, res.Status)
	}

	return res, nil
}

func SearchEverywhere(what string) ([]TorrentItem, error) {
	return searchItems(searchEverywhere(what))
}

func SearchAudioBooks(what string) ([]TorrentItem, error) {
	return searchItems(searchAudioBooks(what))
}

func SearchMovies(what string) ([]TorrentItem, error) {
	return searchItems(searchMovies(what))
}

func SearchSeries(what string) ([]TorrentItem, error) {
	return searchItems(searchSeries(what))
}

func SearchBooks(what string) ([]TorrentItem, error) {
	return searchItems(searchTextBooks(what))
}

func searchItems(uri string) ([]TorrentItem, error) {
	res, err := makeRequest(func() (*http.Request, error) {
		return http.NewRequest("GET", uri, nil)
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
			iSeeds := seedsToInt(items[i].Seeds) - dvdCategoryPenalty(items[i].Category)
			jSeeds := seedsToInt(items[j].Seeds) - dvdCategoryPenalty(items[j].Category)
			if iSeeds > jSeeds {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func dvdCategoryPenalty(category string) int {
	if strings.Contains(strings.ToLower(category), "dvd") {
		return 100
	} else {
		return 0
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

// &o=10&s=2 means sort by seeds descending
const rutrackerBaseUrl = "https://rutracker.org/forum/tracker.php"

// search everywhere
func searchEverywhere(what string) string {
	return fmt.Sprintf("%s?nm=%s&o=10&s=2", rutrackerBaseUrl, what)
}

// to find categories: go to rutracker, search any word, select categories, copy paste list from url
func searchTextBooks(what string) string {
	allCategories := "1037,1101,1238,1325,1335,1337,1341,1349,1353,1400,1410,1411,1412,1415,1418,1422,1423,1424,1425,1426,1427,1428,1429,1430,1431,1432,1433,1436,1445,1446,1447,1477,1520,1523,1528,1575,1680,1681,1683,1684,1685,1686,1687,1688,1689,1696,1801,1802,1961,1967,2019,2020,2021,2022,2023,2024,2026,2027,2028,2029,2030,2031,2032,2033,2034,2037,2038,2039,2041,2042,2043,2044,2045,2046,2047,2048,2049,2054,2055,2056,2074,2080,2086,2099,21,2114,2125,2129,2130,2131,2132,2133,2141,2157,2189,2190,2191,2192,2193,2194,2195,2196,2202,2215,2216,2217,2218,2223,2224,2252,2253,2254,2313,2314,2315,2319,2320,2336,2337,2349,2375,2376,2386,2418,2422,2424,2427,2432,2433,2434,2435,2436,2437,2438,2439,2440,2441,2442,2443,2444,2445,2446,2447,2452,2453,2458,2461,2462,2463,2464,2465,2468,2469,2470,2471,2472,2473,2476,2477,2494,2515,2516,2517,2518,2519,2520,2521,2524,2525,2526,2527,2528,2543,281,295,31,39,565,667,669,745,753,754,764,765,767,768,769,770,862,919,944,946,977,980,995"
	return fmt.Sprintf("%s?f=%s&nm=%s&o=10&s=2", rutrackerBaseUrl, allCategories, what)
}

func searchAudioBooks(what string) string {
	allCategories := "2348,2387,2388,2389,661,2127,2137,2327,399,402,467,490,499,695,1279,1350,2165,2328,401,403,716,1909"
	return fmt.Sprintf("%s?f=%s&nm=%s&o=10&s=2", rutrackerBaseUrl, allCategories, what)
}

func searchMovies(what string) string {
	// numbers are subdirectories of rutracker.org like https://rutracker.org/forum/viewforum.php?f=93
	//106,1666,22,376,941
	//1235,166,185,187,1950,2090,2091,2092,2093,212,2200,2221,2459,252,2540,505,7,934
	//124,1543,1577,709
	//1247,140,1457,194,2198,2199,2201,2339,312,313
	//1908
	//1936
	// excluded DVD 100,101,1576,1670,2220,572,877,905,93
	// excluded audio cover 185
	allCategories := "106,1666,22,376,941,1235,166,187,1950,2090,2091,2092,2093,212,2200,2221,2459,252,2540,505,7,934,124,1543,1577,709,1247,140,1457,194,2198,2199,2201,2339,312,313,1908,1936"
	return fmt.Sprintf("%s?f=%s&nm=%s&o=10&s=2", rutrackerBaseUrl, allCategories, what)
}

func searchSeries(what string) string {
	all := "81,920,842,235,242,1531,1102,387,195,119,1803,266,193,1459,1288,1498,864,315"
	return fmt.Sprintf("%s?f=%s&nm=%s&o=10&s=2", rutrackerBaseUrl, all, what)
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

func DownloadTorrentFileToStream(topicId string) (io.ReadCloser, error) {
	res, err := makeRequest(func() (*http.Request, error) {
		return http.NewRequest("GET", downloadCall(topicId), nil)
	})

	if err != nil {
		return nil, err
	}

	// defer res.Body.Close()
	return res.Body, err
}

type TorrentItem struct {
	Title    string
	Size     string
	Seeds    string
	TopicId  string
	Category string
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
		if len(titleTag.Nodes) == 0 {
			return
		}

		title := titleTag.Nodes[0].FirstChild.Data
		category := string(decodeWindows1251([]uint8(row.Find(".ts-text").Nodes[0].FirstChild.Data)))
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
		item := TorrentItem{Title: title, Size: size, Seeds: seeds, TopicId: topicId, Category: category}
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
