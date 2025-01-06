package transmission

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/hekmon/transmissionrpc/v3"
)

var client *transmissionrpc.Client
var RPC_URI string
var RPC_PORT_FROM int
var RPC_PORT_TO int
var dynamicPort int

// http://127.0.0.1:9091/transmission/rpc
func getTransmissionUriString(port int) string {
	return fmt.Sprintf("http://%s:%d/transmission/rpc", RPC_URI, port)
}

func getClient() (*transmissionrpc.Client, error) {
	if client != nil {
		ok, _ := checkRPCConnection(client)
		if ok {
			return client, nil
		} else {
			client = nil
			dynamicPort = 0
		}
	}

	if dynamicPort == 0 {
		// check if any port in range is open
		for port := RPC_PORT_FROM; port <= RPC_PORT_TO; port++ {
			endpoint := getTransmissionUriString(port)
			fmt.Println("RPC checking uri ", endpoint)
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", RPC_URI, port))
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
				fmt.Println("RPC checking error ", err)
			}
		}
	}

	return nil, fmt.Errorf("no rpc port found")
}

// uri in format http://127.0.0.1:9091/transmission/rpc
func makeClient(uri string) (*transmissionrpc.Client, error) {
	endpoint, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	tbt, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		return nil, err
	}

	ok, err := checkRPCConnection(tbt)
	if !ok {
		return nil, err
	}

	return tbt, nil
}

func CheckRPCConnection() (bool, error) {
	tbt, err := getClient()
	if err != nil {
		return false, err
	}

	ok, err := checkRPCConnection(tbt)

	if ok {
		client = tbt
		return true, nil
	} else {
		return false, err
	}
}

func checkRPCConnection(clientLocal *transmissionrpc.Client) (bool, error) {
	ok, serverVersion, serverMinimumVersion, err := clientLocal.RPCVersion(context.Background())
	if err != nil {
		return false, err
	}

	if !ok {
		return false, fmt.Errorf("remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, transmissionrpc.RPCVersion, serverMinimumVersion)
	}

	fmt.Printf("Remote transmission RPC version (v%d) is compatible with our transmissionrpc library (v%d)\n",
		serverVersion, transmissionrpc.RPCVersion)

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

func AddTorrent(magnet string) (bool, error) {
	tbt, err := getClient()
	if err != nil {
		return false, err
	}

	torrent, err := tbt.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
		Filename: &magnet,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false, err
	} else {
		// Only 3 fields will be returned/set in the Torrent struct
		fmt.Println(*torrent.ID)
		fmt.Println(*torrent.Name)
		fmt.Println(*torrent.HashString)
		return true, nil
	}
}
