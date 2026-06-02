package worker

import (
	"context"
	"encoding/json"
	"log"

	"devflow/internal/mq"
	"devflow/internal/service"
)

func StartFeedConsumer(ctx context.Context, broker *mq.Broker, posts *service.PostService) error {
	if broker == nil {
		return nil
	}
	deliveries, err := broker.Consume(mq.QueueFeedDistribute, "devflow-feed-worker")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					return
				}
				var event mq.PostPublishedEvent
				if err := json.Unmarshal(delivery.Body, &event); err != nil {
					log.Printf("feed worker decode event failed: body_len=%d err=%v", len(delivery.Body), err)
					_ = delivery.Nack(false, false)
					continue
				}
				if err := posts.DistributeFeedNow(ctx, event.AuthorID, event.PostID, event.CreatedAt); err != nil {
					log.Printf("feed worker distribute failed: event_id=%s author_id=%d post_id=%d err=%v",
						event.EventID, event.AuthorID, event.PostID, err)
					_ = delivery.Nack(false, false)
					continue
				}
				_ = delivery.Ack(false)
			}
		}
	}()
	return nil
}
