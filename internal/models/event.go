package models

import "github.com/bluesky-social/indigo/api/bsky"

type Event struct {
	Type string
	// May be large, so copying it every time through the channel is wasteful
	Post *bsky.FeedPost
	DID  string
}
