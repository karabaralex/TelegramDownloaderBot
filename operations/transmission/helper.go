package transmission

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	return "", fmt.Errorf("no new file found")
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

func GetAllTorrentsAsString() (string, error) {
	torrents, err := GetAllTorrents()
	if err != nil {
		return "", fmt.Errorf("failed to get torrents: %w", err)
	}

	if len(torrents) == 0 {
		return "No torrents found", nil
	}
	// Sort torrents so downloading ones are last
	sort.Slice(torrents, func(i, j int) bool {
		// If either torrent doesn't have a status, treat as non-downloading
		if torrents[i].Status == nil || torrents[j].Status == nil {
			return false
		}

		// Return true if i should come before j
		iDownloading := *torrents[i].Status == 4
		jDownloading := *torrents[j].Status == 4
		return !iDownloading && jDownloading
	})

	var result strings.Builder
	for _, t := range torrents {
		if t.Name != nil {
			result.WriteString(fmt.Sprintf("%s, %.1f%%, %s, /delete_%d\n", *t.Name, *t.PercentDone*100, *t.Status, *t.ID))
		}
	}

	return result.String(), nil
}
