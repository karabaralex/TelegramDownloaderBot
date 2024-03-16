package transmission

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type WatchedFolder struct {
	folderPath string
	fileMap    map[string]bool
	allFiles   []string
}

// init fileMap with all files in folder argument
func New(path string) WatchedFolder {
	fileMap := make(map[string]bool)
	allFiles := readAllFilesInFolder(path)

	// Loop through the files
	for _, filename := range allFiles {
		fmt.Println(filename)
		fileMap[filename] = true
	}

	return WatchedFolder{path, fileMap, allFiles}
}

// return new file or error
func (folder WatchedFolder) WaitForNewFileWithRetry(seconds int) (string, error) {
	maxRetry := seconds
	for i := 0; i < maxRetry; i++ {
		list := folder.ReadAllFilesInFolder()

		// Loop through the files
		for _, filename := range list {
			// if file is not in map, add it and send to channel
			if _, ok := folder.fileMap[filename]; !ok {
				folder.fileMap[filename] = true
				fmt.Println("NewFile:" + filename)
				return filename, nil
			}
		}

		// sleep for a second
		time.Sleep(time.Second)
	}

	return "", fmt.Errorf("No new file found")
}

// return new file or error
func (folder WatchedFolder) WaitForNewFile() (string, error) {
	// infinite loop
	for {
		list := folder.ReadAllFilesInFolder()

		// Loop through the files
		for _, filename := range list {
			// if file is not in map, add it and send to channel
			if _, ok := folder.fileMap[filename]; !ok {
				folder.fileMap[filename] = true
				fmt.Println("NewFile:" + filename)
				return filename, nil
			}
		}

		// sleep for a second
		time.Sleep(time.Second)
	}
}

// returns list of filenames
func (folder WatchedFolder) ReadAllFilesInFolder() []string {
	return readAllFilesInFolder(folder.folderPath)
}

// returns list of filenames
func readAllFilesInFolder(path string) []string {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	// Get a list of all of the files in the directory
	list, err := file.Readdir(-1)
	if err != nil {
		log.Panic(err)
	}

	result := make([]string, len(list))

	// Loop through the files
	for _, f := range list {
		fullPath := filepath.Join(path, f.Name())
		title, err := GetTorrentTitle(fullPath)
		if err != nil || len(title) == 0 {
			result = append(result, f.Name())
		} else {
			result = append(result, title)
		}
	}

	return result
}
