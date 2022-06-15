package transmission

import (
	"fmt"
	"log"
	"os"
	"time"
)

var fileMap map[string]bool
var folderPath string

// init fileMap with all files in folder argument
func Create(path string) {
	folderPath = path
	fileMap = make(map[string]bool)
	list := ReadAllFilesInFolder()

	// Loop through the files
	for _, filename := range list {
		fmt.Println(filename)
		fileMap[filename] = true
	}
}

// return new file or error
func WaitForNewFile() (string, error) {
	maxRetry := 5
	for i := 0; i < maxRetry; i++ {
		list := ReadAllFilesInFolder()

		// Loop through the files
		for _, filename := range list {
			// if file is not in map, add it and send to channel
			if _, ok := fileMap[filename]; !ok {
				fileMap[filename] = true
				fmt.Println("NewFile:" + filename)
				return filename, nil
			}
		}

		// sleep for a second
		time.Sleep(time.Second)
	}

	return "", fmt.Errorf("No new file found")
}

// returns list of filenames
func ReadAllFilesInFolder() []string {
	// Open the file
	file, err := os.Open(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Get a list of all of the files in the directory
	list, err := file.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	result := make([]string, len(list))

	// Loop through the files
	for _, f := range list {
		// add the file name to result
		result = append(result, f.Name())
	}

	return result
}
