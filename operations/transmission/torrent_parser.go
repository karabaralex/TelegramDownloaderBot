package transmission

import (
	"os"

	"github.com/anacrolix/torrent/metainfo"
)

// GetTorrentTitle takes the path to a .torrent file and returns its title.
func GetTorrentTitle(path string) (string, error) {
	// Open .torrent file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Parse .torrent file
	metainfo, err := metainfo.Load(file)
	if err != nil {
		return "", err
	}

	// Get title from .torrent file
	info, err := metainfo.UnmarshalInfo()
	if err != nil {
		return "", err
	}

	// Return title
	return info.Name, nil
}
