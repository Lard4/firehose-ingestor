package firehose

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/repo"
	"github.com/gorilla/websocket"
	"github.com/lard4/firehose-ingestor/internal/kafka"
	"github.com/lard4/firehose-ingestor/internal/models"
)

type Runner struct {
	client   *Client
	producer *kafka.KafkaProducer
}

func NewRunner(c *Client, p *kafka.KafkaProducer) *Runner {
	return &Runner{
		client:   c,
		producer: p,
	}
}

// Connects to the firehose and starts processing events, sending them to the Events channel. It blocks until the context is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	fmt.Println("Connecting to firehose at ", r.client.URL)

	wsocket, _, err := websocket.DefaultDialer.Dial(r.client.URL, http.Header{})
	if err != nil {
		return err
	}
	defer wsocket.Close()

	fmt.Println("connected to bluesky firehose")

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(commitEvent *atproto.SyncSubscribeRepos_Commit) error {

			repoReader, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(commitEvent.Blocks))
			if err != nil {
				// skip this commit if we can't read the repo data, but log the error
				fmt.Println("Error reading repo data from commit event:", err)
				return nil
			}

			for _, op := range commitEvent.Ops {
				if op.Action == "create" && strings.HasPrefix(op.Path, "app.bsky.feed.post/") {
					// this is the `did`
					fmt.Println("Event from ", commitEvent.Repo)
					fmt.Printf("  - new post created with path %s\n", op.Path)

					_, rec, err := repoReader.GetRecord(ctx, op.Path)
					if err != nil {
						fmt.Println("Error getting record from repo reader:", err)
						continue
					}

					post, ok := rec.(*bsky.FeedPost)
					if !ok {
						fmt.Println("Error asserting record to appbsky.FeedPost:", err)
						continue
					}

					var cids []string
					if post.Embed != nil && post.Embed.EmbedImages != nil {
						for _, img := range post.Embed.EmbedImages.Images {
							cids = append(cids, img.Image.Ref.String())
						}
					}

					event := models.Event{
						Type: "post",
						DID:  commitEvent.Repo,
						Path: op.Path,
					}

					// put into a channel to provide backpressure (and we can move the publishing out eventually to a dedicated worker pool)
					// r.client.Events <- event

					if err := r.producer.Send(ctx, event); err != nil {
						fmt.Println("Error sending event to Kafka:", err)
					}
				}
			}

			return nil
		},
	}

	sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)

	// blocking call to handle the firehose stream
	events.HandleRepoStream(ctx, wsocket, sched, nil)

	<-ctx.Done()
	fmt.Println("firehose client shutting down")
	return nil
}
