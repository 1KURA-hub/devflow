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
					if deadLetterErr := broker.DeadLetter(ctx, mq.QueueFeedDistribute, delivery, "invalid_json: "+err.Error()); deadLetterErr != nil {
						log.Printf("feed worker dead-letter invalid JSON failed: err=%v", deadLetterErr)
					}
					continue
				}
				if err := validatePostPublishedEvent(event); err != nil {
					log.Printf("feed worker invalid event: event_id=%s err=%v", event.EventID, err)
					if deadLetterErr := broker.DeadLetter(ctx, mq.QueueFeedDistribute, delivery, "invalid_event: "+err.Error()); deadLetterErr != nil {
						log.Printf("feed worker dead-letter invalid event failed: event_id=%s err=%v", event.EventID, deadLetterErr)
					}
					continue
				}
				if err := posts.DistributeFeedNow(ctx, event.AuthorID, event.PostID, event.CreatedAt); err != nil {
					log.Printf("feed worker distribute failed: event_id=%s author_id=%d post_id=%d err=%v",
						event.EventID, event.AuthorID, event.PostID, err)
					if retryErr := broker.RetryOrDeadLetter(ctx, mq.QueueFeedDistribute, delivery, err); retryErr != nil {
						log.Printf("feed worker retry/dead-letter failed: event_id=%s err=%v", event.EventID, retryErr)
					}
					continue
				}
				if err := delivery.Ack(false); err != nil {
					log.Printf("feed worker ack failed: event_id=%s err=%v", event.EventID, err)
				}
			}
		}
	}()
	return nil
}
