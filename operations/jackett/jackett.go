package jackett

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

var JACKET_URI string
var JACKET_KEY string
var JACKET_PORT_FROM int
var JACKET_PORT_TO int
var dynamicPort int
var client *Jackett

type FetchRequest struct {
	Query      string
	Trackers   []string
	Categories []uint
}

type FetchResponse struct {
	Results  []Result
	Indexers []Indexer
}

type jackettTime struct {
	time.Time
}

func (jt *jackettTime) UnmarshalJSON(b []byte) (err error) {
	str := strings.Trim(string(b), `"`)
	if str == "0001-01-01T00:00:00" {
	} else if len(str) == 19 {
		jt.Time, err = time.Parse(time.RFC3339, str+"Z")
	} else {
		jt.Time, err = time.Parse(time.RFC3339, str)
	}
	return
}

type Result struct {
	BannerUrl            string
	BlackholeLink        string
	Category             []uint
	CategoryDesc         string
	Comments             string
	Description          string
	DownloadVolumeFactor float32
	Files                uint
	FirstSeen            jackettTime
	Gain                 float32
	Grabs                uint
	Guid                 string
	Imdb                 uint
	InfoHash             string
	Link                 string
	MagnetUri            string
	MinimumRatio         float32
	MinimumSeedTime      uint
	Peers                uint
	PublishDate          jackettTime
	RageID               uint
	Seeders              uint
	Size                 uint
	TMDb                 uint
	TVDBId               uint
	Title                string
	Tracker              string
	TrackerId            string
	UploadVolumeFactor   float32
}

type Config struct {
	Notices                   []interface{} `json:"notices"`
	Port                      int           `json:"port"`
	External                  bool          `json:"external"`
	LocalBindAddress          string        `json:"local_bind_address"`
	Cors                      bool          `json:"cors"`
	ApiKey                    string        `json:"api_key"`
	BlackholeDir              *string       `json:"blackholedir"`
	UpdateDisabled            bool          `json:"updatedisabled"`
	Prerelease                bool          `json:"prerelease"`
	Password                  string        `json:"password"`
	Logging                   bool          `json:"logging"`
	BasePathOverride          *string       `json:"basepathoverride"`
	BaseUrlOverride           *string       `json:"baseurloverride"`
	CacheEnabled              bool          `json:"cache_enabled"`
	CacheTtl                  int           `json:"cache_ttl"`
	CacheMaxResultsPerIndexer int           `json:"cache_max_results_per_indexer"`
	FlareSolverrUrl           *string       `json:"flaresolverrurl"`
	FlareSolverrMaxTimeout    int           `json:"flaresolverr_maxtimeout"`
	OmdbKey                   *string       `json:"omdbkey"`
	OmdbUrl                   *string       `json:"omdburl"`
	AppVersion                string        `json:"app_version"`
	CanRunNetCore             bool          `json:"can_run_netcore"`
	ProxyType                 int           `json:"proxy_type"`
	ProxyUrl                  *string       `json:"proxy_url"`
	ProxyPort                 *int          `json:"proxy_port"`
	ProxyUsername             *string       `json:"proxy_username"`
	ProxyPassword             *string       `json:"proxy_password"`
}

type Indexer struct {
	Error   string
	ID      string
	Name    string
	Results uint
	Status  uint
}

type Jackett struct {
	uri string
}

// http://10.0.4.124:49158
func getTransmissionUriString(port int) string {
	return fmt.Sprintf("http://%s:%d", JACKET_URI, port)
}

func GetClient() (*Jackett, error) {
	if client != nil {
		err := client.pingServer(context.Background())
		if err == nil {
			return client, nil
		} else {
			client = nil
			dynamicPort = 0
		}
	}

	if dynamicPort == 0 {
		// check if any port in range is open
		for port := JACKET_PORT_FROM; port <= JACKET_PORT_TO; port++ {
			endpoint := getTransmissionUriString(port)
			fmt.Println("jackett checking uri ", endpoint)
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", JACKET_URI, port))
			if err == nil {
				conn.Close()
				client, err = makeClient(endpoint)
				if err != nil {
					continue
				} else {
					dynamicPort = port
					return client, nil
				}
			} else {
				fmt.Println("jackett checking error ", err)
			}
		}
	}

	return nil, fmt.Errorf("no jackett port found")
}

// uri in format http://127.0.0.1:9091/transmission/rpc
func makeClient(uri string) (*Jackett, error) {
	_, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	jackett := newJackett(uri)
	err = jackett.pingServer(context.Background())
	if err != nil {
		return nil, err
	}

	return jackett, nil
}

func newJackett(uri string) *Jackett {
	return &Jackett{uri: uri}
}

func (j *Jackett) generateFetchURL(fr *FetchRequest) (string, error) {
	u, err := url.Parse(j.uri)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse apiURL %q", j.uri)
	}
	u.Path = "/api/v2.0/indexers/all/results"
	q := u.Query()
	q.Set("apikey", JACKET_KEY)
	for _, t := range fr.Trackers {
		q.Add("Tracker[]", t)
	}
	for _, c := range fr.Categories {
		q.Add("Category[]", fmt.Sprintf("%v", c))
	}
	if fr.Query != "" {
		q.Add("Query", fr.Query)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (j *Jackett) Fetch(ctx context.Context, fr *FetchRequest) (*FetchResponse, error) {
	u, err := j.generateFetchURL(fr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate fetch url")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make fetch request")
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to invoke fetch request")
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read fetch data")
	}
	var fres FetchResponse
	err = json.Unmarshal(data, &fres)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal fetch data with url=%v and data=%v", u, string(data))
	}
	return &fres, nil
}

func (j *Jackett) generateCapsURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse apiURL %q", uri)
	}
	u.Path = "/api/v2.0/server/config"
	q := u.Query()
	q.Set("apikey", JACKET_KEY)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (j *Jackett) pingServer(ctx context.Context) error {
	u, err := j.generateCapsURL(j.uri)
	if err != nil {
		return errors.Wrap(err, "failed to generate fetch url")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return errors.Wrap(err, "failed to make caps request")
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to invoke caps request")
	}
	defer res.Body.Close()
	if res.Header["Server"][0] != "Kestrel" { // seems like this is what jackett have
		return errors.Errorf("incorrect server")
	}

	return nil
}

// func init() {
// 	if v, ok := os.LookupEnv("JACKETT_API_URL"); ok {
// 		apiURL = v
// 	}
// 	if v, ok := os.LookupEnv("JACKETT_API_KEY"); ok {
// 		apiKey = v
// 	}
// }

func DownloadTorrentFile(filepath string, url string) error {
	res, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}

	// Initialize an HTTP client
	client := &http.Client{}

	// Execute the request
	resp, err := client.Do(res)
	if err != nil {
		return err
	}
	// defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	fmt.Printf("saving file for %s to %s\n", url, filepath)
	saveToFile(resp.Body, filepath)
	return err
}

func saveToFile(body io.ReadCloser, file string) {
	// Create the file
	out, _ := os.Create(file)
	defer out.Close()

	// Write the body to file
	io.Copy(out, body)
}

func DownloadTorrentFileToStream(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// Initialize an HTTP client
	client := &http.Client{}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Status code: %d", resp.StatusCode)
	}

	// defer res.Body.Close()
	return resp.Body, nil
}
