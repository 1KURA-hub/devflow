package worker

import (
	"context"
	"encoding/json"
	"log"

	"devflow/internal/mq"
	"devflow/internal/service"
)

func StartNotificationConsumer(ctx context.Context, broker *mq.Broker, notifications *service.NotificationService) error {
	if broker == nil {
		return nil
	}
	deliveries, err := broker.Consume(mq.QueueNotificationCreate, "devflow-notification-worker")
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
				var event mq.NotificationEvent
				if err := json.Unmarshal(delivery.Body, &event); err != nil {
					log.Printf("notification worker decode event: %v", err)
					_ = delivery.Nack(false, false)
					continue
				}
				if err := notifications.CreateNow(ctx, service.CreateNotificationInput{
					UserID:    event.UserID,
					ActorID:   event.ActorID,
					Type:      event.Type,
					PostID:    event.PostID,
					CommentID: event.CommentID,
					Content:   event.Content,
				}); err != nil {
					log.Printf("notification worker create notification: event_id=%s err=%v", event.EventID, err)
					_ = delivery.Nack(false, false)
					continue
				}
				_ = delivery.Ack(false)
			}
		}
	}()
	return nil
}
