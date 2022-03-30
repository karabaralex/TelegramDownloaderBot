package operations

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

/**
send string to https://paste.rs/ and return url
*/
func SendStringToPastebin(body string) (string, error) {
	url := "https://paste.rs/"

	// Create the request
	req, err := http.NewRequest("POST", url,
		strings.NewReader(body))
	if err != nil {
		return "", err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 201 && resp.StatusCode != 206 {
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	pastedUrl, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Close the response body
	defer resp.Body.Close()
	convertedString := string(pastedUrl)
	// trim and remove new lines
	convertedString = strings.Trim(convertedString, "\n")
	return convertedString, nil
}
