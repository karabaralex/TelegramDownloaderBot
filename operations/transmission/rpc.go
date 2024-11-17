package transmission

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hekmon/transmissionrpc/v3"
)

var client *transmissionrpc.Client
var RPC_URI string

func getClient() (*transmissionrpc.Client, error) {
	if client != nil {
		return client, nil
	}

	endpoint, err := url.Parse(RPC_URI)
	if err != nil {
		return nil, err
	}

	tbt, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		return nil, err
	}

	return tbt, nil
}

func CheckRPCConnection() (bool, error) {
	tbt, err := getClient()
	if err != nil {
		return false, err
	}

	ok, serverVersion, serverMinimumVersion, err := tbt.RPCVersion(context.Background())
	if err != nil {
		return false, err
	}

	if !ok {
		return false, fmt.Errorf("remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, transmissionrpc.RPCVersion, serverMinimumVersion)
	}

	fmt.Printf("Remote transmission RPC version (v%d) is compatible with our transmissionrpc library (v%d)\n",
		serverVersion, transmissionrpc.RPCVersion)

	client = tbt
	return true, nil
}

func GetAllTorrents() ([]transmissionrpc.Torrent, error) {
	tbt, err := getClient()
	if err != nil {
		return nil, err
	}

	torrents, err := tbt.TorrentGetAll(context.Background())
	if err != nil {
		return nil, err
	}

	return torrents, nil
}

func RemoveTorrent(id int64) (bool, error) {
	tbt, err := getClient()
	if err != nil {
		return false, err
	}

	err = tbt.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{IDs: []int64{id}})
	if err != nil {
		return false, err
	}

	return true, nil
}
