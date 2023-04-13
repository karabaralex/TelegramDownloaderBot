package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	BASE_URL = "https://kvdb.io"
)

var API_KEY string

func SetKeyValue(key string, value string) bool {
	url := fmt.Sprintf("%s/%s/%s", BASE_URL, API_KEY, key)
	data := map[string]string{"value": value}
	payload, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Error: Failed to encode payload:", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))

	if err != nil {
		fmt.Println("Error: Failed to create request:", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error: Failed to send request:", err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf(`Value for key "%s" was successfully set to "%s"\n`, key, value)
		return true
	} else {
		fmt.Println("Error: Failed to set the value")
		return false
	}
}

func GetAllKeys() []string {
	url := fmt.Sprintf("%s/%s/?values=true&format=json", BASE_URL, API_KEY)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		fmt.Println("Error: Failed to create request:", err)
		return []string{}
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error: Failed to send request:", err)
		return []string{}
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Read the response body.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return []string{}
		}
		defer resp.Body.Close()

		// Unmarshal the JSON data into a two-dimensional slice of empty interfaces.
		var data [][]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("Error parsing JSON response:", err)
			return []string{}
		}

		// Convert the two-dimensional slice to a slice of strings.
		var result []string
		for _, item := range data {
			s := fmt.Sprintf("%v ", item[0])
			// for _, v := range item {
			// 	s += fmt.Sprintf("%v ", v)
			// }
			result = append(result, s)
		}

		// Print the result.
		fmt.Println(result)
		return result
	} else {
		fmt.Println("Error: Failed to get the keys")
		return []string{}
	}
}

func DeleteKey(key string) bool {
	url := fmt.Sprintf("%s/%s/%s", BASE_URL, API_KEY, key)

	req, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		fmt.Println("Error: Failed to create request:", err)
		return false
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error: Failed to send request:", err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf(`Key "%s" was successfully deleted\n`, key)
		return true
	} else {
		fmt.Printf(`Error: Failed to delete key "%s"\n`, key)
		return false
	}
}
