package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/lard4/firehose-ingestor/internal/http"
	"github.com/lard4/firehose-ingestor/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	client *xrpc.Client
	server *http.Server
}

func NewKafkaConsumer() *KafkaConsumer {
	return &KafkaConsumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:           []string{"localhost:9092"},
			GroupID:           "bluesky-post-reader-go",
			Topic:             "bluesky-posts",
			SessionTimeout:    30 * time.Second,
			HeartbeatInterval: 3 * time.Second,
		}),
		&xrpc.Client{
			Host: "https://public.api.bsky.app",
		},
		http.NewServer(),
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	kc.server.Start(ctx)

	for {
		// consideration: if the websocket write fails, we lose the message.
		// We could add retry logic here, but for simplicity we'll just log the error and move on to the next message.
		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var event models.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			fmt.Println("Error unmarshalling event:", err)
			continue
		}

		post, err := kc.fetchPost(ctx, event)
		if err != nil {
			fmt.Println("Error fetching post details:", err)
			continue
		}

		// Send message to http server for website to consume
		jsonPayload, err := json.Marshal(post)
		if err != nil {
			fmt.Printf("Error marshalling post to JSON: %v\n", err)
		}

		//fmt.Println("Sending post to HTTP server: ", string(jsonPayload))
		kc.server.Send(jsonPayload)
	}

	return nil
}

func (kc *KafkaConsumer) fetchPost(ctx context.Context, event models.Event) (*models.Post, error) {
	uri := fmt.Sprintf("at://%s/%s", event.DID, event.Path)
	return kc.fetchPostRecursively(ctx, uri, nil)
}

// fetches the post details for the given URI, and if it's a reply, it also fetches the parent post details recursively.
func (kc *KafkaConsumer) fetchPostRecursively(ctx context.Context, uri string, parent *models.Post) (*models.Post, error) {
	posts, err := bsky.FeedGetPosts(ctx, kc.client, []string{uri})
	if err != nil {
		return nil, err
	}

	if len(posts.Posts) == 0 {
		fmt.Println("No posts found for URI:", uri)
		return nil, fmt.Errorf("no posts found for URI: %s", uri)
	}

	post := posts.Posts[0] // we only requested one post, so this should be safe

	record, ok := post.Record.Val.(*bsky.FeedPost)
	if !ok {
		return nil, fmt.Errorf("unexpected record type: %T", post.Record.Val)
	}

	var imageUrls []string
	if post.Embed != nil && post.Embed.EmbedImages_View != nil {
		for _, img := range post.Embed.EmbedImages_View.Images {
			imageUrls = append(imageUrls, img.Fullsize)
		}
	}

	postEvent := models.Post{
		Handle:    post.Author.Handle,
		Content:   record.Text,
		Likes:     int(*post.LikeCount),
		Quotes:    int(*post.QuoteCount),
		Replies:   int(*post.ReplyCount),
		Reposts:   int(*post.RepostCount),
		ImageUrls: imageUrls,
		Parent:    nil,
	}

	if record.Reply != nil && record.Reply.Parent != nil {
		postEvent.Parent, err = kc.fetchPostRecursively(ctx, record.Reply.Parent.Uri, parent)
		if err != nil {
			fmt.Println("Error fetching parent post:", err)
			return nil, err
		}
		return &postEvent, nil
	} else {
		return &postEvent, nil
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
