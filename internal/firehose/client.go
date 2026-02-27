package firehose

import "github.com/bluesky-social/indigo/api/bsky"

type Event struct {
	Type string
	Post bsky.FeedPost
	DID  string
}

type Client struct {
	URL    string
	Events chan Event
}

func NewClient() *Client {
	return &Client{
		URL:    "wss://relay1.us-east.bsky.network/xrpc/com.atproto.sync.subscribeRepos",
		Events: make(chan Event, 1000),
	}
}
