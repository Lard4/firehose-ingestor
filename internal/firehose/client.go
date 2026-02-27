package firehose

import (
	"github.com/lard4/firehose-ingestor/internal/models"
)

type Client struct {
	URL    string
	Events chan models.Event
}

func NewClient() *Client {
	return &Client{
		URL:    "wss://relay1.us-east.bsky.network/xrpc/com.atproto.sync.subscribeRepos",
		Events: make(chan models.Event, 1000),
	}
}
